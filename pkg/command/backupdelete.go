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

package command

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/cloudnative-pg/machinery/pkg/log"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanUtils "github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// DeleteBackupsByPolicy executes a command that deletes backups, given the Barman object store configuration,
// the retention policies, the server name and the environment variables
func DeleteBackupsByPolicy(
	ctx context.Context,
	barmanConfiguration *barmanApi.BarmanObjectStoreConfiguration,
	serverName string,
	env []string,
	retentionPolicy string,
) error {
	contextLogger := log.FromContext(ctx).WithName("barman")

	var options []string
	if barmanConfiguration.EndpointURL != "" {
		options = append(options, "--endpoint-url", barmanConfiguration.EndpointURL)
	}

	options, err := AppendCloudProviderOptionsFromConfiguration(ctx, options, barmanConfiguration)
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
	cmd := exec.Command(barmanUtils.BarmanCloudBackupDelete, options...) // #nosec G204
	cmd.Env = env
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err = cmd.Run()
	if err != nil {
		contextLogger.Error(err,
			"Error invoking "+barmanUtils.BarmanCloudBackupDelete,
			"options", options,
			"stdout", stdoutBuffer.String(),
			"stderr", stderrBuffer.String())
		return err
	}

	return nil
}
