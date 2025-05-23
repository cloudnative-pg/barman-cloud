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

package backup

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/cloudnative-pg/machinery/pkg/execlog"
	"github.com/cloudnative-pg/machinery/pkg/log"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanCatalog "github.com/cloudnative-pg/barman-cloud/pkg/catalog"
	barmanCommand "github.com/cloudnative-pg/barman-cloud/pkg/command"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// Command represents a barman backup command
type Command struct {
	configuration *barmanApi.BarmanObjectStoreConfiguration
}

// NewBackupCommand creates a new barman backup command
func NewBackupCommand(
	configuration *barmanApi.BarmanObjectStoreConfiguration,
) *Command {
	return &Command{
		configuration: configuration,
	}
}

// GetDataConfiguration gets the configuration in the `Data` object of the Barman configuration
func (b *Command) GetDataConfiguration(
	options []string,
) ([]string, error) {
	if b.configuration.Data == nil {
		return options, nil
	}

	if len(b.configuration.Data.Compression) != 0 {
		options = append(
			options,
			fmt.Sprintf("--%v", b.configuration.Data.Compression))
	}

	if len(b.configuration.Data.Encryption) != 0 {
		options = append(
			options,
			"--encryption",
			string(b.configuration.Data.Encryption))
	}

	if b.configuration.Data.ImmediateCheckpoint {
		options = append(
			options,
			"--immediate-checkpoint")
	}

	if b.configuration.Data.Jobs != nil {
		options = append(
			options,
			"--jobs",
			strconv.Itoa(int(*b.configuration.Data.Jobs)))
	}

	return b.configuration.Data.AppendAdditionalCommandArgs(options), nil
}

// GetBarmanCloudBackupOptions extract the list of command line options to be used with
// barman-cloud-backup
func (b *Command) GetBarmanCloudBackupOptions(
	ctx context.Context,
	backupName string,
	serverName string,
) ([]string, error) {
	options := []string{
		"--user", "postgres",
		"--name", backupName,
	}

	options, err := b.GetDataConfiguration(options)
	if err != nil {
		return nil, err
	}

	if len(b.configuration.Tags) > 0 {
		tags, err := utils.MapToBarmanTagsFormat("--tags", b.configuration.Tags)
		if err != nil {
			return nil, err
		}
		options = append(options, tags...)
	}

	if len(b.configuration.EndpointURL) > 0 {
		options = append(
			options,
			"--endpoint-url",
			b.configuration.EndpointURL)
	}

	options, err = barmanCommand.AppendCloudProviderOptionsFromConfiguration(ctx, options, b.configuration)
	if err != nil {
		return nil, err
	}

	options = append(
		options,
		b.configuration.DestinationPath,
		serverName)

	return options, nil
}

// GetExecutedBackupInfo get the status information about the executed backup
func (b *Command) GetExecutedBackupInfo(
	ctx context.Context,
	backupName string,
	serverName string,
	env []string,
) (*barmanCatalog.BarmanBackup, error) {
	return barmanCommand.GetBackupByName(
		ctx,
		backupName,
		serverName,
		b.configuration,
		env,
	)
}

// Take takes a backup
func (b *Command) Take(
	ctx context.Context,
	backupName string,
	serverName string,
	env []string,
	backupTemporaryDirectory string,
) error {
	log := log.FromContext(ctx)

	options, backupErr := b.GetBarmanCloudBackupOptions(ctx, backupName, serverName)
	if backupErr != nil {
		log.Error(backupErr, "while getting barman-cloud-backup options")
		return backupErr
	}

	// record the backup beginning
	log.Info("Starting barman-cloud-backup", "options", options)

	cmd := exec.Command(utils.BarmanCloudBackup, options...) // #nosec G204
	cmd.Env = env
	cmd.Env = append(cmd.Env, "TMPDIR="+backupTemporaryDirectory)
	if err := execlog.RunStreaming(cmd, utils.BarmanCloudBackup); err != nil {
		const badArgumentsErrorCode = "3"
		if err.Error() == badArgumentsErrorCode {
			descriptiveError := errors.New("invalid arguments for barman-cloud-backup. " +
				"Ensure that the additionalCommandArgs field is correctly populated")
			log.Error(descriptiveError, "error while executing barman-cloud-backup",
				"arguments", options)
			return descriptiveError
		}
		return err
	}

	log.Info("Completed barman-cloud-backup", "options", options)

	return nil
}
