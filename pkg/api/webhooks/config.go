/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
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

	allErrors = append(allErrors, validateSSECustomerKey(barmanObjectStore, path)...)

	return allErrors
}

// validateSSECustomerKey checks that SSE-C is not combined with the other
// server-side encryption modes, which barman-cloud rejects
func validateSSECustomerKey(
	barmanObjectStore *api.BarmanObjectStoreConfiguration,
	path *field.Path,
) field.ErrorList {
	if barmanObjectStore.AWS == nil || barmanObjectStore.AWS.SSECustomerKey == nil {
		return nil
	}

	const message = "cannot be used together with s3Credentials.sseCustomerKey"
	allErrors := field.ErrorList{}
	if barmanObjectStore.Data != nil && barmanObjectStore.Data.Encryption != "" {
		allErrors = append(allErrors, field.Invalid(
			path.Child("data", "encryption"),
			barmanObjectStore.Data.Encryption,
			message,
		))
	}
	if barmanObjectStore.Wal != nil && barmanObjectStore.Wal.Encryption != "" {
		allErrors = append(allErrors, field.Invalid(
			path.Child("wal", "encryption"),
			barmanObjectStore.Wal.Encryption,
			message,
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
