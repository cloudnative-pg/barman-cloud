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

package webhooks

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/cloudnative-pg/barman-cloud/pkg/api"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// ValidateBackupConfiguration validates the backup configuration
func ValidateBackupConfiguration(
	barmanObjectStore *api.BarmanObjectStoreConfiguration,
	path *field.Path,
) field.ErrorList {
	allErrors := field.ErrorList{}

	if barmanObjectStore == nil {
		return nil
	}

	credentialsCount := 0
	if barmanObjectStore.Azure != nil {
		credentialsCount++
		allErrors = barmanObjectStore.Azure.ValidateAzureCredentials(
			path.Child("azureCredentials"),
		)
	}
	if barmanObjectStore.AWS != nil {
		credentialsCount++
		allErrors = barmanObjectStore.AWS.ValidateAwsCredentials(
			path.Child("awsCredentials"),
		)
	}
	if barmanObjectStore.Google != nil {
		credentialsCount++
		allErrors = barmanObjectStore.Google.ValidateGCSCredentials(
			field.NewPath("spec", "backupConfiguration", "googleCredentials"))
	}
	if credentialsCount == 0 {
		allErrors = append(allErrors, field.Invalid(
			path,
			barmanObjectStore,
			"missing credentials. "+
				"One and only one of azureCredentials, s3Credentials and googleCredentials are required",
		))
	}
	if credentialsCount > 1 {
		allErrors = append(allErrors, field.Invalid(
			path,
			barmanObjectStore,
			"too many credentials. "+
				"One and only one of azureCredentials, s3Credentials and googleCredentials are required",
		))
	}

	return allErrors
}

// ValidateRetentionPolicy validates a Barman retention policy
func ValidateRetentionPolicy(retentionPolicy string, path *field.Path) field.ErrorList {
	allErrors := field.ErrorList{}

	if retentionPolicy == "" {
		return nil
	}

	_, err := utils.ParsePolicy(retentionPolicy)
	if err != nil {
		allErrors = append(allErrors, field.Invalid(
			path,
			retentionPolicy,
			"not a valid retention policy",
		))
	}

	return allErrors
}
