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

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
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
	switch {
	case credentials.AWS != nil:
		options = append(
			options,
			"--cloud-provider",
			"aws-s3")
	case credentials.Azure != nil:
		options = append(
			options,
			"--cloud-provider",
			"azure-blob-storage")

		// deprecated, to be removed in future versions
		if useDefaultAzureCredentials(ctx) {
			options = append(
				options,
				"--credential",
				"default")
			break
		}

		if credentials.Azure.UseDefaultAzureCredentials {
			options = append(
				options,
				"--credential",
				"default")
			break
		}

		if credentials.Azure.InheritFromAzureAD {
			options = append(
				options,
				"--credential",
				"managed-identity")
			break
		}
	case credentials.Google != nil:
		options = append(
			options,
			"--cloud-provider",
			"google-cloud-storage")
	}

	return options, nil
}

type contextKey string

// contextKeyUseDefaultAzureCredentials contains a bool indicating if the default azure credentials should be used
const contextKeyUseDefaultAzureCredentials contextKey = "useDefaultAzureCredentials"

func useDefaultAzureCredentials(ctx context.Context) bool {
	v := ctx.Value(contextKeyUseDefaultAzureCredentials)
	if v == nil {
		return false
	}
	result, ok := v.(bool)
	if !ok {
		return false
	}
	return result
}

// ContextWithDefaultAzureCredentials creates a context that contains the contextKeyUseDefaultAzureCredentials flag.
// When set to true barman-cloud will use the default Azure credentials.
//
// Deprecated: Use AzureCredentials.UseDefaultAzureCredentials instead.
func ContextWithDefaultAzureCredentials(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, enabled)
}
