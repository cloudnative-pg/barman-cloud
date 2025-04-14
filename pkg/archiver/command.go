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

package archiver

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudnative-pg/machinery/pkg/log"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanCommand "github.com/cloudnative-pg/barman-cloud/pkg/command"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// GatherWALFilesToArchive reads from the archived status the list of WAL files
// that can be archived in parallel way.
// `requestedWALFile` is the name of the file whose archiving was requested by
// PostgreSQL, and that file is always the first of the list and is always included.
// `parallel` is the maximum number of WALs that we can archive in parallel
func (archiver *WALArchiver) GatherWALFilesToArchive(
	ctx context.Context,
	requestedWALFile string,
	parallel int,
) (walList []string) {
	contextLog := log.FromContext(ctx)
	pgWalDirectory := path.Join(os.Getenv("PGDATA"), "pg_wal")
	archiveStatusPath := path.Join(pgWalDirectory, "archive_status")
	noMoreWALFilesNeeded := errors.New("no more files needed")

	// allocate parallel + 1 only if it does not overflow. Cap otherwise
	var walListLength int
	if parallel < math.MaxInt-1 {
		walListLength = parallel + 1
	} else {
		walListLength = math.MaxInt - 1
	}
	// slightly more optimized, but equivalent to:
	// walList = []string{requestedWALFile}
	walList = make([]string, 1, walListLength)
	walList[0] = requestedWALFile

	err := filepath.WalkDir(archiveStatusPath, func(path string, d os.DirEntry, err error) error {
		// If err is set, it means the current path is a directory and the readdir raised an error
		// The only available option here is to skip the path and log the error.
		if err != nil {
			contextLog.Error(err, "failed reading path", "path", path)
			return filepath.SkipDir
		}

		if len(walList) >= parallel {
			return noMoreWALFilesNeeded
		}

		// We don't process directories beside the archive status path
		if d.IsDir() {
			// We want to proceed exploring the archive status folder
			if path == archiveStatusPath {
				return nil
			}

			return filepath.SkipDir
		}

		// We only process ready files
		if !strings.HasSuffix(path, ".ready") {
			return nil
		}

		walFileName := strings.TrimSuffix(filepath.Base(path), ".ready")

		// We are already archiving the requested WAL file,
		// and we need to avoid archiving it twice.
		// requestedWALFile is usually "pg_wal/wal_file_name" and
		// we compare it with the path we read
		if strings.HasSuffix(requestedWALFile, walFileName) {
			return nil
		}

		walList = append(walList, filepath.Join(pgWalDirectory, walFileName))
		return nil
	})

	// In this point err must be nil or noMoreWALFilesNeeded, if it is something different
	// there is a programming error
	if err != nil && err != noMoreWALFilesNeeded {
		contextLog.Error(err, "unexpected error while reading the list of WAL files to archive")
	}

	return walList
}

// BarmanCloudWalArchiveOptions calculates the set of options to be
// used with barman-cloud-wal-archive
func (archiver *WALArchiver) BarmanCloudWalArchiveOptions(
	ctx context.Context,
	configuration *barmanApi.BarmanObjectStoreConfiguration,
	clusterName string,
) ([]string, error) {
	var options []string
	if configuration.Wal != nil {
		if len(configuration.Wal.Compression) != 0 {
			options = append(
				options,
				fmt.Sprintf("--%v", configuration.Wal.Compression))
		}
		if len(configuration.Wal.Encryption) != 0 {
			options = append(
				options,
				"-e",
				string(configuration.Wal.Encryption))
		}
		options = configuration.Wal.AppendArchiveAdditionalCommandArgs(options)
	}
	if len(configuration.EndpointURL) > 0 {
		options = append(
			options,
			"--endpoint-url",
			configuration.EndpointURL)
	}

	if len(configuration.Tags) > 0 {
		tags, err := utils.MapToBarmanTagsFormat("--tags", configuration.Tags)
		if err != nil {
			return nil, err
		}
		options = append(options, tags...)
	}

	if len(configuration.HistoryTags) > 0 {
		historyTags, err := utils.MapToBarmanTagsFormat("--history-tags", configuration.HistoryTags)
		if err != nil {
			return nil, err
		}
		options = append(options, historyTags...)
	}

	options, err := barmanCommand.AppendCloudProviderOptionsFromConfiguration(ctx, options, configuration)
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
