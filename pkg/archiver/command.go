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
	"fmt"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanCommand "github.com/cloudnative-pg/barman-cloud/pkg/command"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

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
