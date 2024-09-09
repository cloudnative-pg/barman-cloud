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

// Package archiver manages the WAL archiving process
package archiver

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"

	"github.com/cloudnative-pg/plugin-barman-cloud/pkg/command"
	"github.com/cloudnative-pg/plugin-barman-cloud/pkg/spool"
	"github.com/cloudnative-pg/plugin-barman-cloud/pkg/types"
	"github.com/cloudnative-pg/plugin-barman-cloud/pkg/walarchive"
)

// WALArchiver is a structure containing every info need to archive a set of WAL files
// using barman-cloud-wal-archive
type WALArchiver struct {
	// The spool of WAL files to be archived in parallel
	spool *spool.WALSpool

	// The environment that should be used to invoke barman-cloud-wal-archive
	env []string

	pgDataDirectory string

	// this should become a grpc interface
	barmanArchiver *walarchive.BarmanArchiver

	// Supporting functions
	FileUtils spool.FileUtils
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

// New creates a new WAL archiver
func New(
	ctx context.Context,
	env []string,
	spoolDirectory string,
	pgDataDirectory string,
	fileUtils spool.FileUtils,
	runStreaming func(cmd *exec.Cmd, cmdName string) (err error),
	removeEmptyFileArchive func() error,
) (archiver *WALArchiver, err error) {
	contextLog := logging.FromContext(ctx)
	var walArchiveSpool *spool.WALSpool

	if walArchiveSpool, err = spool.New(fileUtils, spoolDirectory); err != nil {
		contextLog.Info("Cannot initialize the WAL spool", "spoolDirectory", spoolDirectory)
		return nil, fmt.Errorf("while creating spool directory: %w", err)
	}

	archiver = &WALArchiver{
		FileUtils:       fileUtils,
		spool:           walArchiveSpool,
		env:             env,
		pgDataDirectory: pgDataDirectory,
		barmanArchiver: &walarchive.BarmanArchiver{
			Env:                    env,
			RunStreaming:           runStreaming,
			Touch:                  walArchiveSpool.Touch,
			RemoveEmptyFileArchive: removeEmptyFileArchive,
		},
	}
	return archiver, nil
}

// DeleteFromSpool checks if a WAL file is in the spool and, if it is, remove it
func (archiver *WALArchiver) DeleteFromSpool(walName string) (hasBeenDeleted bool, err error) {
	var isContained bool

	// this code assumes the wal-archive command is run at most once at each instant,
	// given that PostgreSQL will call it sequentially without overlapping
	isContained, err = archiver.spool.Contains(walName)
	if !isContained || err != nil {
		return false, err
	}

	return true, archiver.spool.Remove(walName)
}

// ArchiveList archives a list of WAL files in parallel
func (archiver *WALArchiver) ArchiveList(
	ctx context.Context,
	walNames []string,
	options []string,
) (result []WALArchiverResult) {
	res := archiver.barmanArchiver.ArchiveList(ctx, walNames, options)
	for _, re := range res {
		result = append(result, WALArchiverResult{
			WalName:   re.WalName,
			Err:       re.Err,
			StartTime: re.StartTime,
			EndTime:   re.EndTime,
		})
	}
	return result
}

// CheckWalArchiveDestination checks if the destinationObjectStore is ready perform archiving.
// Based on this ticket in Barman https://github.com/EnterpriseDB/barman/issues/432
// and its implementation https://github.com/EnterpriseDB/barman/pull/443
// The idea here is to check ONLY if we're archiving the wal files for the first time in the bucket
// since in this case the command barman-cloud-check-wal-archive will fail if the bucket exist and
// contain wal files inside
func (archiver *WALArchiver) CheckWalArchiveDestination(ctx context.Context, options []string) error {
	return archiver.barmanArchiver.CheckWalArchiveDestination(ctx, options)
}

// BarmanCloudCheckWalArchiveOptions create the options needed for the `barman-cloud-check-wal-archive`
// command.
func (archiver *WALArchiver) BarmanCloudCheckWalArchiveOptions(
	configuration *types.BarmanObjectStoreConfiguration,
	clusterName string,
) ([]string, error) {
	var options []string
	if len(configuration.EndpointURL) > 0 {
		options = append(
			options,
			"--endpoint-url",
			configuration.EndpointURL)
	}

	options, err := command.AppendCloudProviderOptionsFromConfiguration(options, configuration)
	if err != nil {
		return nil, err
	}

	serverName := clusterName
	if len(configuration.ServerName) != 0 {
		serverName = configuration.ServerName
	}
	options = append(
		options,
		configuration.DestinationPath,
		serverName)
	return options, nil
}
