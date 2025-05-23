/*
Copyright The CloudNativePG Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restorer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"sync"
	"time"

	"github.com/cloudnative-pg/machinery/pkg/execlog"
	"github.com/cloudnative-pg/machinery/pkg/log"

	"github.com/cloudnative-pg/barman-cloud/pkg/spool"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

const (
	endOfWALStreamFlagFilename = "end-of-wal-stream"
)

// ErrWALNotFound is returned when the WAL is not found in the cloud archive
var ErrWALNotFound = errors.New("WAL not found")

// WALRestorer is a structure containing every info needed to restore
// some WALs from the object storage
type WALRestorer struct {
	// The spool of WAL files to be archived in parallel
	spool *spool.WALSpool

	// The environment that should be used to invoke barman-cloud-wal-archive
	env []string
}

// Result is the structure filled by the restore process on completion
type Result struct {
	// The name of the WAL file to restore
	WalName string

	// Where to store the restored WAL file
	DestinationPath string

	// If not nil, this is the error that has been detected
	Err error

	// The time when we started barman-cloud-wal-archive
	StartTime time.Time

	// The time when end barman-cloud-wal-archive ended
	EndTime time.Time
}

// New creates a new WAL restorer
func New(
	ctx context.Context,
	env []string,
	spoolDirectory string,
) (restorer *WALRestorer, err error) {
	contextLog := log.FromContext(ctx)
	var walRecoverSpool *spool.WALSpool

	if walRecoverSpool, err = spool.New(spoolDirectory); err != nil {
		contextLog.Info("Cannot initialize the WAL spool", "spoolDirectory", spoolDirectory)
		return nil, fmt.Errorf("while creating spool directory: %w", err)
	}

	restorer = &WALRestorer{
		spool: walRecoverSpool,
		env:   env,
	}
	return restorer, nil
}

// RestoreFromSpool restores a certain file from the spool, returning a boolean flag indicating
// is the file was in the spool or not. If the file was in the spool, it will be moved into the
// specified destination path
func (restorer *WALRestorer) RestoreFromSpool(walName, destinationPath string) (wasInSpool bool, err error) {
	err = restorer.spool.MoveOut(walName, destinationPath)
	switch {
	case err == spool.ErrorNonExistentFile:
		return false, nil

	case err != nil:
		return false, err

	default:
		return true, nil
	}
}

// SetEndOfWALStream add end-of-wal-stream in the spool directory
func (restorer *WALRestorer) SetEndOfWALStream() error {
	contains, err := restorer.IsEndOfWALStream()
	if err != nil {
		return err
	}

	if contains {
		return nil
	}

	err = restorer.spool.Touch(endOfWALStreamFlagFilename)
	if err != nil {
		return err
	}

	return nil
}

// IsEndOfWALStream check whether end-of-wal-stream flag is presents in the spool directory
func (restorer *WALRestorer) IsEndOfWALStream() (bool, error) {
	isEOS, err := restorer.spool.Contains(endOfWALStreamFlagFilename)
	if err != nil {
		return false, fmt.Errorf("failed to check end-of-wal-stream flag: %w", err)
	}

	return isEOS, nil
}

// ResetEndOfWalStream remove end-of-wal-stream flag from the spool directory
func (restorer *WALRestorer) ResetEndOfWalStream() error {
	err := restorer.spool.Remove(endOfWALStreamFlagFilename)
	if err != nil {
		return fmt.Errorf("failed to remove end-of-wal-stream flag: %w", err)
	}

	return nil
}

// RestoreList restores a list of WALs. The first WAL of the list will go directly into the
// destination path, the others will be adopted by the spool
func (restorer *WALRestorer) RestoreList(
	ctx context.Context,
	fetchList []string,
	destinationPath string,
	options []string,
) (resultList []Result) {
	resultList = make([]Result, len(fetchList))
	contextLog := log.FromContext(ctx)
	var waitGroup sync.WaitGroup

	for idx := range fetchList {
		waitGroup.Add(1)
		go func(walIndex int) {
			result := &resultList[walIndex]
			result.WalName = fetchList[walIndex]
			if walIndex == 0 {
				// The WAL that PostgreSQL requested will go directly
				// to the destination path
				result.DestinationPath = destinationPath
			} else {
				result.DestinationPath = restorer.spool.FileName(result.WalName)
			}

			result.StartTime = time.Now()
			result.Err = restorer.Restore(fetchList[walIndex], result.DestinationPath, options)
			result.EndTime = time.Now()

			elapsedWalTime := result.EndTime.Sub(result.StartTime)
			if result.Err == nil {
				contextLog.Info(
					"Restored WAL file",
					"walName", result.WalName,
					"startTime", result.StartTime,
					"endTime", result.EndTime,
					"elapsedWalTime", elapsedWalTime)
			} else if walIndex == 0 {
				// We don't log errors for prefetched WALs but just for the
				// first WAL, which is the one requested by PostgreSQL.
				//
				// The implemented prefetch is speculative and this WAL may just
				// not exist, this means that this may not be a real error.
				if errors.Is(result.Err, ErrWALNotFound) {
					contextLog.Info(
						"WAL file not found in the recovery object store",
						"walName", result.WalName,
						"options", options,
						"startTime", result.StartTime,
						"endTime", result.EndTime,
						"elapsedWalTime", elapsedWalTime)
				} else {
					contextLog.Warning(
						"Failed restoring WAL file (Postgres might retry)",
						"walName", result.WalName,
						"options", options,
						"startTime", result.StartTime,
						"endTime", result.EndTime,
						"elapsedWalTime", elapsedWalTime,
						"error", result.Err)
				}
			}
			waitGroup.Done()
		}(idx)
	}

	waitGroup.Wait()
	return resultList
}

// Restore restores a WAL file from the object store
func (restorer *WALRestorer) Restore(
	walName, destinationPath string,
	baseOptions []string,
) error {
	const (
		exitCodeBucketOrWalNotFound = 1
		exitCodeConnectivityError   = 2
		exitCodeInvalidWalName      = 3
		exitCodeGeneric             = 4
	)

	optionsLength := len(baseOptions)
	if optionsLength >= math.MaxInt-2 {
		return fmt.Errorf("can't restore wal file %v, options too long", walName)
	}
	options := make([]string, optionsLength, optionsLength+2)
	copy(options, baseOptions)
	options = append(options, walName, destinationPath)

	barmanCloudWalRestoreCmd := exec.Command(
		utils.BarmanCloudWalRestore,
		options...) // #nosec G204
	barmanCloudWalRestoreCmd.Env = restorer.env

	err := execlog.RunStreaming(barmanCloudWalRestoreCmd, utils.BarmanCloudWalRestore)
	if err == nil {
		return nil
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return fmt.Errorf("unexpected failure retrieving %q with %s: %w",
			walName, utils.BarmanCloudWalRestore, err)
	}

	// nolint: lll
	// source: https://github.com/EnterpriseDB/barman/blob/26ed480cd0268dfd3f8546cc97660d4bd827772a/tests/test_barman_cloud_wal_restore.py
	exitCode := exitError.ExitCode()
	switch exitCode {
	case exitCodeBucketOrWalNotFound:
		return fmt.Errorf("object storage or file not found %s: %w", walName, ErrWALNotFound)
	case exitCodeConnectivityError:
		return fmt.Errorf("connectivity failure while executing %s, retrying",
			utils.BarmanCloudWalRestore)
	case exitCodeInvalidWalName:
		return fmt.Errorf("invalid name for a WAL file: %s", walName)
	case exitCodeGeneric:
		return fmt.Errorf("generic error code encountered while executing %s",
			utils.BarmanCloudWalRestore)
	default:
		return fmt.Errorf("encountered an unrecognized exit code error: '%d' while executing %s",
			exitCode,
			utils.BarmanCloudWalRestore)
	}
}
