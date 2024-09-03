package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

var _ = Describe("WalBackupConfiguration.AppendAdditionalCommandArgs", func() {
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

	It("contain only one of storage account key and SAS token", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
			StorageSasToken: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "sasToken",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())

		azureCredentials = AzureCredentials{
			StorageAccount: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey:      nil,
			StorageSasToken: nil,
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).ToNot(BeEmpty())
	})

	It("is correct when the storage key is used", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageKey",
			},
			StorageSasToken: nil,
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("is correct when the sas token is used", func() {
		azureCredentials := AzureCredentials{
			StorageAccount: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
			StorageKey: nil,
			StorageSasToken: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "sasToken",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("is correct even if only the connection string is specified", func() {
		azureCredentials := AzureCredentials{
			ConnectionString: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "connectionString",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})

	It("it is not correct when the connection string is specified with other parameters", func() {
		azureCredentials := AzureCredentials{
			ConnectionString: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "connectionString",
			},
			StorageAccount: &SecretKeySelector{
				LocalObjectReference: LocalObjectReference{
					Name: "azure-config",
				},
				Key: "storageAccount",
			},
		}
		Expect(azureCredentials.ValidateAzureCredentials(path)).To(BeEmpty())
	})
})
