package webhooks

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/cloudnative-pg/barman-cloud/pkg/api"
	"github.com/cloudnative-pg/barman-cloud/pkg/utils"
)

// ValidateBackupConfiguration validates the backup configuration
func ValidateBackupConfiguration(obj api.BarmanObjectStoreWebhookGetter) field.ErrorList {
	allErrors := field.ErrorList{}
	barmanObjectStore := obj.GetBarmanObjectStore()

	if barmanObjectStore == nil {
		return nil
	}

	basePath := buildFieldPath(obj.GetBarmanObjectStorePath()...)

	credentialsCount := 0
	if barmanObjectStore.BarmanCredentials.Azure != nil {
		credentialsCount++
		allErrors = barmanObjectStore.BarmanCredentials.Azure.ValidateAzureCredentials(
			basePath.Child("azureCredentials"),
		)
	}
	if barmanObjectStore.BarmanCredentials.AWS != nil {
		credentialsCount++
		allErrors = barmanObjectStore.BarmanCredentials.AWS.ValidateAwsCredentials(
			basePath.Child("awsCredentials"),
		)
	}
	if barmanObjectStore.BarmanCredentials.Google != nil {
		credentialsCount++
		allErrors = barmanObjectStore.BarmanCredentials.Google.ValidateGCSCredentials(
			field.NewPath("spec", "backupConfiguration", "googleCredentials"))
	}
	if credentialsCount == 0 {
		allErrors = append(allErrors, field.Invalid(
			basePath,
			barmanObjectStore,
			"missing credentials. "+
				"One and only one of azureCredentials, s3Credentials and googleCredentials are required",
		))
	}
	if credentialsCount > 1 {
		allErrors = append(allErrors, field.Invalid(
			basePath,
			barmanObjectStore,
			"too many credentials. "+
				"One and only one of azureCredentials, s3Credentials and googleCredentials are required",
		))
	}

	if obj.GetRetentionPolicy() != "" {
		_, err := utils.ParsePolicy(obj.GetRetentionPolicy())
		if err != nil {
			allErrors = append(allErrors, field.Invalid(
				buildFieldPath(obj.GetRetentionPolicyPath()...),
				obj.GetRetentionPolicy(),
				"not a valid retention policy",
			))
		}
	}

	return allErrors
}

func buildFieldPath(paths ...string) *field.Path {
	if len(paths) == 0 {
		return field.NewPath("")
	}
	path := field.NewPath(paths[0])
	for _, s := range paths {
		path = path.Child(s)
	}
	return path
}
