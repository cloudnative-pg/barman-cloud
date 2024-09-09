package backup

import (
	"context"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/execlog"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/postgres"
	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/logging"
	"os/exec"
	"strconv"

	barmanCapabilities "github.com/cloudnative-pg/plugin-barman-cloud/pkg/capabilities"
	barmanCatalog "github.com/cloudnative-pg/plugin-barman-cloud/pkg/catalog"
	barmanCommand "github.com/cloudnative-pg/plugin-barman-cloud/pkg/command"
	barmanTypes "github.com/cloudnative-pg/plugin-barman-cloud/pkg/types"
	"github.com/cloudnative-pg/plugin-barman-cloud/pkg/utils"
)

// Command represents a barman backup command
type Command struct {
	configuration *barmanTypes.BarmanObjectStoreConfiguration
	capabilities  *barmanCapabilities.Capabilities
}

// NewBackupCommand creates a new barman backup command
func NewBackupCommand(configuration *barmanTypes.BarmanObjectStoreConfiguration, capabilities *barmanCapabilities.Capabilities) *Command {
	return &Command{
		configuration: configuration,
		capabilities:  capabilities,
	}
}

// getDataConfiguration gets the configuration in the `Data` object of the Barman configuration
func (b *Command) GetDataConfiguration(
	options []string,
) ([]string, error) {
	if b.configuration.Data == nil {
		return options, nil
	}

	if b.configuration.Data.Compression == barmanTypes.CompressionTypeSnappy && !b.capabilities.HasSnappy {
		return nil, fmt.Errorf("snappy compression is not supported in Barman %v", b.capabilities.Version)
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
	backupName string,
	serverName string,
	exec barmanCapabilities.LegacyExecutor,
) ([]string, error) {
	options := []string{
		"--user", "postgres",
	}

	if b.capabilities.ShouldExecuteBackupWithName(exec) {
		options = append(options, "--name", backupName)
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

	options, err = barmanCommand.AppendCloudProviderOptionsFromConfiguration(options, b.configuration)
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
	exec barmanCapabilities.LegacyExecutor,
	env []string,
) (*barmanCatalog.BarmanBackup, error) {
	if b.capabilities.ShouldExecuteBackupWithName(exec) {
		return barmanCommand.GetBackupByName(
			ctx,
			backupName,
			serverName,
			b.configuration,
			env,
		)
	}

	// we don't know the id or the name of the executed backup so it fetches the last executed barman backup.
	// it could create issues in case of concurrent backups. It is a deprecated way of detecting the backup.
	return barmanCommand.GetLatestBackup(
		ctx,
		serverName,
		b.configuration,
		env,
	)
}

// IsCompatible checks if barman can back up this version of PostgreSQL
func (b *Command) IsCompatible(postgresVers semver.Version) error {
	switch {
	case postgresVers.Major == 15 && b.capabilities.Version.Major < 3:
		return fmt.Errorf(
			"PostgreSQL %d is not supported by Barman %d.x",
			postgresVers.Major,
			b.capabilities.Version.Major,
		)
	default:
		return nil
	}
}

// Take takes a backup
func (b *Command) Take(
	ctx context.Context,
	backupName string,
	serverName string,
	env []string,
	legacyExecutor barmanCapabilities.LegacyExecutor,
) error {
	log := logging.FromContext(ctx)

	options, backupErr := b.GetBarmanCloudBackupOptions(backupName, serverName, legacyExecutor)
	if backupErr != nil {
		log.Error(backupErr, "while getting barman-cloud-backup options")
		return backupErr
	}

	// record the backup beginning
	log.Info("Starting barman-cloud-backup", "options", options)

	cmd := exec.Command(barmanCapabilities.BarmanCloudBackup, options...) // #nosec G204
	cmd.Env = env
	cmd.Env = append(cmd.Env, "TMPDIR="+postgres.BackupTemporaryDirectory)
	if err := execlog.RunStreaming(cmd, barmanCapabilities.BarmanCloudBackup); err != nil {
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

	return nil
}
