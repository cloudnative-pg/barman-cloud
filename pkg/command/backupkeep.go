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

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanCapabilities "github.com/cloudnative-pg/barman-cloud/pkg/capabilities"
	"github.com/cloudnative-pg/machinery/pkg/log"
)

func KeepBackup(
	ctx context.Context,
	barmanConfiguration *barmanApi.BarmanObjectStoreConfiguration,
	backupName string,
	serverName string,
	keep bool,
	keepTarget string,
	env []string) error {
	contextLogger := log.FromContext(ctx).WithName("barman")

	options := make([]string, 0)

	if barmanConfiguration.EndpointURL != "" {
		options = append(options, "--endpoint-url", barmanConfiguration.EndpointURL)
	}

	options, err := AppendCloudProviderOptionsFromConfiguration(ctx, options, barmanConfiguration)
	if err != nil {
		return err
	}

	if keep {
		options = append(options, "--target", keepTarget)
	} else {
		options = append(options, "--release")
	}

	options = append(options, barmanConfiguration.DestinationPath, serverName, backupName)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd := exec.Command(barmanCapabilities.BarmanCloudBackupKeep, options...) // #nosec G204
	cmd.Env = env
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err = cmd.Run()
	if err != nil {
		contextLogger.Error(err,
			"Can't set keep target on backup",
			"command", barmanCapabilities.BarmanCloudBackupKeep,
			"options", options,
			"stdout", stdoutBuffer.String(),
			"stderr", stderrBuffer.String())
		return err
	}

	return nil
}
