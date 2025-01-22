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
	"context"
	"fmt"

	"github.com/cloudnative-pg/machinery/pkg/log"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	barmanCapabilities "github.com/cloudnative-pg/barman-cloud/pkg/capabilities"
)

// CloudWalRestoreOptions returns the options needed to execute the barman command successfully
func CloudWalRestoreOptions(
	ctx context.Context,
	configuration *barmanApi.BarmanObjectStoreConfiguration,
	clusterName string,
) ([]string, error) {
	var options []string
	if len(configuration.EndpointURL) > 0 {
		options = append(
			options,
			"--endpoint-url",
			configuration.EndpointURL)
	}

	options, err := AppendCloudProviderOptionsFromConfiguration(ctx, options, configuration)
	if err != nil {
		return nil, err
	}

	serverName := clusterName
	if len(configuration.ServerName) != 0 {
		serverName = configuration.ServerName
	}

	options = append(options, configuration.DestinationPath, serverName)
	options = configuration.Wal.AppendRestoreAdditionalCommandArgs(options)

	return options, nil
}

// AppendCloudProviderOptionsFromConfiguration takes an options array and adds the cloud provider specified
// in the Barman configuration object
func AppendCloudProviderOptionsFromConfiguration(
	ctx context.Context,
	options []string,
	barmanConfiguration *barmanApi.BarmanObjectStoreConfiguration,
) ([]string, error) {
	return appendCloudProviderOptions(ctx, options, barmanConfiguration.BarmanCredentials)
}

// AppendCloudProviderOptionsFromBackup takes an options array and adds the cloud provider specified
// in the Backup object
func AppendCloudProviderOptionsFromBackup(
	ctx context.Context,
	options []string,
	credentials barmanApi.BarmanCredentials,
) ([]string, error) {
	return appendCloudProviderOptions(ctx, options, credentials)
}

// appendCloudProviderOptions takes an options array and adds the cloud provider specified as arguments
func appendCloudProviderOptions(
	ctx context.Context,
	options []string,
	credentials barmanApi.BarmanCredentials,
) ([]string, error) {
	logger := log.FromContext(ctx)

	capabilities, err := barmanCapabilities.CurrentCapabilities()
	if err != nil {
		return nil, err
	}

	switch {
	case credentials.AWS != nil:
		if capabilities.HasS3 {
			options = append(
				options,
				"--cloud-provider",
				"aws-s3")
		}
	case credentials.Azure != nil:
		if !capabilities.HasAzure {
			err := fmt.Errorf(
				"barman >= 2.13 is required to use Azure object storage, current: %v",
				capabilities.Version)
			logger.Error(err, "Barman version not supported")
			return nil, err
		}

		options = append(
			options,
			"--cloud-provider",
			"azure-blob-storage")

		if !credentials.Azure.InheritFromAzureAD {
			break
		}

		if checkUseDefaultAzureCredentials(ctx) {
			break
		}

		if !capabilities.HasAzureManagedIdentity {
			err := fmt.Errorf(
				"barman >= 2.18 is required to use azureInheritFromAzureAD, current: %v",
				capabilities.Version)
			logger.Error(err, "Barman version not supported")
			return nil, err
		}

		options = append(
			options,
			"--credential",
			"managed-identity")
	case credentials.Google != nil:
		if !capabilities.HasGoogle {
			err := fmt.Errorf(
				"barman >= 2.19 is required to use Google Cloud Storage, current: %v",
				capabilities.Version)
			logger.Error(err, "Barman version not supported")
			return nil, err
		}
		options = append(
			options,
			"--cloud-provider",
			"google-cloud-storage")
	}

	return options, nil
}

type contextKey string

const useDefaultAzureCredentials contextKey = "useDefaultAzureCredentials"

func checkUseDefaultAzureCredentials(ctx context.Context) bool {
	v := ctx.Value(useDefaultAzureCredentials)
	if v == nil {
		return false
	}
	result, ok := v.(bool)
	if !ok {
		return false
	}
	return result
}
