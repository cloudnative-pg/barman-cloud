package command

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/cloudnative-pg/cloudnative-pg-machinery/pkg/log"

	barmanCapabilities "github.com/cloudnative-pg/plugin-barman-cloud/pkg/capabilities"
	barmanTypes "github.com/cloudnative-pg/plugin-barman-cloud/pkg/types"
	barmanUtils "github.com/cloudnative-pg/plugin-barman-cloud/pkg/utils"
)

// DeleteBackupsByPolicy executes a command that deletes backups, given the Barman object store configuration,
// the retention policies, the server name and the environment variables
func DeleteBackupsByPolicy(
	ctx context.Context,
	barmanConfiguration *barmanTypes.BarmanObjectStoreConfiguration,
	serverName string,
	env []string,
	retentionPolicy string,
) error {
	contextLogger := log.FromContext(ctx).WithName("barman")

	capabilities, err := barmanCapabilities.CurrentCapabilities()
	if err != nil {
		return err
	}

	if !capabilities.HasRetentionPolicy {
		err := fmt.Errorf(
			"barman >= 2.14 is required to use retention policy, current: %v",
			capabilities.Version)
		contextLogger.Error(err, "Failed applying backup retention policies")
		return err
	}

	var options []string
	if barmanConfiguration.EndpointURL != "" {
		options = append(options, "--endpoint-url", barmanConfiguration.EndpointURL)
	}

	options, err = AppendCloudProviderOptionsFromConfiguration(ctx, options, barmanConfiguration)
	if err != nil {
		return err
	}

	parsedPolicy, err := barmanUtils.ParsePolicy(retentionPolicy)
	if err != nil {
		return err
	}

	options = append(
		options,
		"--retention-policy",
		parsedPolicy,
		barmanConfiguration.DestinationPath,
		serverName)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd := exec.Command(barmanCapabilities.BarmanCloudBackupDelete, options...) // #nosec G204
	cmd.Env = env
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err = cmd.Run()
	if err != nil {
		contextLogger.Error(err,
			"Error invoking "+barmanCapabilities.BarmanCloudBackupDelete,
			"options", options,
			"stdout", stdoutBuffer.String(),
			"stderr", stderrBuffer.String())
		return err
	}

	return nil
}
