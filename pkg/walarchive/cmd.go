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

package walarchive

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"sync"
	"time"

	"github.com/cloudnative-pg/machinery/pkg/execlog"
	"github.com/cloudnative-pg/machinery/pkg/fileutils"
	"github.com/cloudnative-pg/machinery/pkg/log"

	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// BarmanArchiver implements a WAL archiver based
// on Barman cloud
type BarmanArchiver struct {
	Env                 []string
	Touch               func(walFile string) error
	EmptyWalArchivePath string
}

// WALArchiverResult contains the result of the archival of one WAL
type WALArchiverResult struct {
	// The WAL that have been archived
	WalName string

	// If not nil, this is the error that has been detected
	Err error

	// The time when we started barman-cloud-wal-archive
	StartTime time.Time

	// The time when end barman-cloud-wal-archive ended
	EndTime time.Time
}

// Archive archives a certain WAL file using barman-cloud-wal-archive.
// See archiveWALFileList for the meaning of the parameters
func (archiver *BarmanArchiver) Archive(
	ctx context.Context,
	walName string,
	baseOptions []string,
) error {
	contextLogger := log.FromContext(ctx)
	optionsLength := len(baseOptions)
	if optionsLength >= math.MaxInt-1 {
		return fmt.Errorf("can't archive wal file %v, options too long", walName)
	}
	options := make([]string, optionsLength, optionsLength+1)
	copy(options, baseOptions)
	options = append(options, walName)

	contextLogger.Info("Executing "+utils.BarmanCloudWalArchive,
		"walName", walName,
		"options", options,
	)

	barmanCloudWalArchiveCmd := exec.Command(utils.BarmanCloudWalArchive, options...) // #nosec G204
	barmanCloudWalArchiveCmd.Env = archiver.Env

	err := execlog.RunStreaming(barmanCloudWalArchiveCmd, utils.BarmanCloudWalArchive)
	if err != nil {
		contextLogger.Error(err, "Error invoking "+utils.BarmanCloudWalArchive,
			"walName", walName,
			"options", options,
			"exitCode", barmanCloudWalArchiveCmd.ProcessState.ExitCode(),
		)
		return fmt.Errorf("unexpected failure invoking %s: %w", utils.BarmanCloudWalArchive, err)
	}

	if err := archiver.fadviseNotUsed(walName); err != nil {
		contextLogger.Error(err, "Error issuing fadvise after archiving WAL",
			"walName", walName,
		)
	}

	// Removes the `.check-empty-wal-archive` file inside PGDATA after the
	// first successful archival of a WAL file.
	if err := fileutils.RemoveFile(archiver.EmptyWalArchivePath); err != nil {
		return fmt.Errorf("error while deleting the check WAL file flag: %w", err)
	}
	return nil
}

// ArchiveList archives a list of WAL files in parallel
func (archiver *BarmanArchiver) ArchiveList(
	ctx context.Context,
	walNames []string,
	options []string,
) (result []WALArchiverResult) {
	contextLog := log.FromContext(ctx)
	result = make([]WALArchiverResult, len(walNames))

	var waitGroup sync.WaitGroup
	for idx := range walNames {
		waitGroup.Add(1)
		go func(walIndex int) {
			defer waitGroup.Done()

			result[walIndex] = WALArchiverResult{
				WalName:   walNames[walIndex],
				StartTime: time.Now(),
				Err:       archiver.Archive(ctx, walNames[walIndex], options),
				EndTime:   time.Now(),
			}
			walStatus := &result[walIndex]

			walContextLog := contextLog.WithValues(
				"walName", walStatus.WalName,
				"startTime", walStatus.StartTime,
				"endTime", walStatus.EndTime,
				"elapsedWalTime", walStatus.EndTime.Sub(walStatus.StartTime),
			)

			if walStatus.Err != nil {
				walContextLog.Warning(
					"Failed archiving WAL: PostgreSQL will retry",
					"error", walStatus.Err)
				return
			}

			if walIndex == 0 {
				walContextLog.Info("Archived WAL file")
				return
			}

			if err := archiver.Touch(walNames[walIndex]); err != nil {
				walContextLog.Warning(
					"WAL file pre-archived, but it could not be added to the spool. PostgreSQL will retry",
					"error", err)
				return
			}

			walContextLog.Info("Pre-archived WAL file (parallel)")
		}(idx)
	}

	waitGroup.Wait()
	return result
}

// CheckWalArchiveDestination checks if the destinationObjectStore is ready perform archiving.
// Based on this ticket in Barman https://github.com/EnterpriseDB/barman/issues/432
// and its implementation https://github.com/EnterpriseDB/barman/pull/443
// The idea here is to check ONLY if we're archiving the wal files for the first time in the bucket
// since in this case the command barman-cloud-check-wal-archive will fail if the bucket exist and
// contain wal files inside
func (archiver *BarmanArchiver) CheckWalArchiveDestination(ctx context.Context, options []string) error {
	contextLogger := log.FromContext(ctx)
	contextLogger.Info("barman-cloud-check-wal-archive checking the first wal")

	contextLogger.Trace("Executing "+utils.BarmanCloudCheckWalArchive,
		"options", options,
	)

	barmanCloudWalArchiveCmd := exec.Command(utils.BarmanCloudCheckWalArchive, options...) // #nosec G204
	barmanCloudWalArchiveCmd.Env = archiver.Env

	err := execlog.RunStreaming(barmanCloudWalArchiveCmd, utils.BarmanCloudCheckWalArchive)
	if err != nil {
		contextLogger.Error(err, "Error invoking "+utils.BarmanCloudCheckWalArchive,
			"options", options,
			"exitCode", barmanCloudWalArchiveCmd.ProcessState.ExitCode(),
		)
		return fmt.Errorf("unexpected failure invoking %s: %w", utils.BarmanCloudWalArchive, err)
	}

	contextLogger.Trace("barman-cloud-check-wal-archive command execution completed")

	return nil
}
