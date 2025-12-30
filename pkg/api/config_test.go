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

package api

import (
	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DataBackupConfiguration.AppendAdditionalCommandArgs", func() {
	var options []string
	var config DataBackupConfiguration
	BeforeEach(func() {
		options = []string{"--option1", "--option2"}
		config = DataBackupConfiguration{
			AdditionalCommandArgs: []string{"--option3", "--option4"},
		}
	})

	It("should append additional command args to the options", func() {
		updatedOptions := config.AppendAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal([]string{"--option1", "--option2", "--option3", "--option4"}))
	})

	It("should return the original options if there are no additional command args", func() {
		config.AdditionalCommandArgs = nil
		updatedOptions := config.AppendAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal(options))
	})
})

var _ = Describe("WalBackupConfiguration.AppendArchiveAdditionalCommandArgs", func() {
	var options []string
	var config WalBackupConfiguration
	BeforeEach(func() {
		options = []string{"--option1", "--option2"}
		config = WalBackupConfiguration{
			ArchiveAdditionalCommandArgs: []string{"--option3", "--option4"},
		}
	})

	It("should append additional command args to the options", func() {
		updatedOptions := config.AppendArchiveAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal([]string{"--option1", "--option2", "--option3", "--option4"}))
	})

	It("should return the original options if there are no additional command args", func() {
		config.ArchiveAdditionalCommandArgs = nil
		updatedOptions := config.AppendArchiveAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal(options))
	})
})

var _ = Describe("WalBackupConfiguration.AppendRestoreAdditionalCommandArgs", func() {
	var options []string
	var config WalBackupConfiguration
	BeforeEach(func() {
		options = []string{"--option1", "--option2"}
		config = WalBackupConfiguration{
			RestoreAdditionalCommandArgs: []string{"--option3", "--option4"},
		}
	})

	It("should append additional command args to the options", func() {
		updatedOptions := config.AppendRestoreAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal([]string{"--option1", "--option2", "--option3", "--option4"}))
	})

	It("should return the original options if there are no additional command args", func() {
		config.RestoreAdditionalCommandArgs = nil
		updatedOptions := config.AppendRestoreAdditionalCommandArgs(options)
		Expect(updatedOptions).To(Equal(options))
	})
})

var _ = Describe("appendAdditionalCommandArgs", func() {
	It("should append additional command args to the options", func() {
		options := []string{"--option1", "--option2"}
		additionalCommandArgs := []string{"--option3", "--option4"}

		updatedOptions := appendAdditionalCommandArgs(additionalCommandArgs, options)
		Expect(updatedOptions).To(Equal([]string{"--option1", "--option2", "--option3", "--option4"}))
	})

	It("should add key value pairs correctly", func() {
		options := []string{"--option1", "--option2"}
		additionalCommandArgs := []string{"--option3", "--option4=value", "--option5=value2"}

		updatedOptions := appendAdditionalCommandArgs(additionalCommandArgs, options)
		Expect(updatedOptions).To(Equal([]string{
			"--option1", "--option2", "--option3",
			"--option4=value", "--option5=value2",
		}))
	})

	It("should not duplicate existing values", func() {
		options := []string{"--option1", "--option2"}
		additionalCommandArgs := []string{"--option2", "--option1"}

		updatedOptions := appendAdditionalCommandArgs(additionalCommandArgs, options)
		Expect(updatedOptions).To(Equal([]string{"--option1", "--option2"}))
	})

	It("should not overwrite existing key value pairs", func() {
		options := []string{"--option1=abc", "--option2"}
		additionalCommandArgs := []string{"--option2", "--option1=def"}

		updatedOptions := appendAdditionalCommandArgs(additionalCommandArgs, options)
		Expect(updatedOptions).To(Equal([]string{"--option1=abc", "--option2"}))
	})
})

var _ = Describe("Barman credentials", func() {
	It("can check when they are empty", func() {
		Expect(BarmanCredentials{}.ArePopulated()).To(BeFalse())
	})

	It("can check when they are not empty", func() {
		Expect(BarmanCredentials{
			Azure: &AzureCredentials{},
		}.ArePopulated()).To(BeTrue())
	})
})

var _ = Describe("azure credentials", func() {
	path := field.NewPath("spec", "backupConfiguration", "azureCredentials")

	It("it is correct when the storage key is specified", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
			StorageSasToken: nil,
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("it is correct when the SAS token is specified", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: nil,
			StorageSasToken: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "sasToken",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("it is correct if only the connection string is specified", func() {
		azureCredentials := AzureCredentials{
			ConnectionString: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "connectionString",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("it is not correct when the connection string is specified with other parameters", func() {
		azureCredentials := AzureCredentials{
			ConnectionString: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "connectionString",
			},
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})

	It("is not correct when storage key is specified without storage account", func() {
		azureCredentials := AzureCredentials{
			StorageKey: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})

	It("is not correct when SAS token is specified without storage account", func() {
		azureCredentials := AzureCredentials{
			StorageSasToken: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "sasToken",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})

	It("is not correct when both storage key and SAS token are specified", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
			StorageSasToken: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "sasToken",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())

		azureCredentials = AzureCredentials{
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey:      nil,
			StorageSasToken: nil,
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})

	It("it is correct when Azure AD authentication is used", func() {
		azureCredentials := AzureCredentials{
			InheritFromAzureAD: true,
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("is not correct when Azure AD authentication is combined with other credentials", func() {
		azureCredentials := AzureCredentials{
			InheritFromAzureAD: true,
			StorageAccount: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})
})
