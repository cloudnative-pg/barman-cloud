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
	"strings"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"
	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("barmanCloudWalRestoreOptions", func() {
	var storageConf *barmanApi.BarmanObjectStoreConfiguration
	BeforeEach(func() {
		storageConf = &barmanApi.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket-name/",
		}
	})

	It("should generate correct arguments without the wal stanza", func(ctx SpecContext) {
		options, err := CloudWalRestoreOptions(ctx, storageConf, "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"s3://bucket-name/ test-cluster",
				))
	})

	It("should generate correct arguments", func(ctx SpecContext) {
		extraOptions := []string{"--read-timeout=60", "-vv"}
		storageConf.Wal = &barmanApi.WalBackupConfiguration{
			RestoreAdditionalCommandArgs: extraOptions,
		}
		options, err := CloudWalRestoreOptions(ctx, storageConf, "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"s3://bucket-name/ test-cluster --read-timeout=60 -vv",
				))
	})
})

var _ = Describe("useDefaultAzureCredentials", func() {
	It("should be false by default", func(ctx SpecContext) {
		Expect(useDefaultAzureCredentials(ctx)).To(BeFalse())
	})

	It("should be false if ctx contains an invalid value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, "invalidValue")
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be false if ctx contains false value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be true only if ctx contains true value", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})
})

var _ = Describe("AppendCloudProviderOptions with Azure credentials", func() {
	var options []string

	BeforeEach(func() {
		options = []string{}
	})

	It("should use default credential when UseDefaultAzureCredentials is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				UseDefaultAzureCredentials: true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})

	It("should use managed-identity credential when InheritFromAzureAD is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				InheritFromAzureAD: true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "managed-identity",
		))
	})

	It("should not use any credential flag for explicit credentials", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				StorageAccount: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{
						Name: "test",
					},
					Key: "account",
				},
				StorageKey: &machineryapi.SecretKeySelector{
					LocalObjectReference: machineryapi.LocalObjectReference{
						Name: "test",
					},
					Key: "key",
				},
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]string{
			"--cloud-provider", "azure-blob-storage",
		}))
	})

	It("should use default credential from context when context flag is set", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{},
		}
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		result, err := appendCloudProviderOptions(newCtx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})

	It("should prioritize UseDefaultAzureCredentials over InheritFromAzureAD", func(ctx SpecContext) {
		credentials := barmanApi.BarmanCredentials{
			Azure: &barmanApi.AzureCredentials{
				UseDefaultAzureCredentials: true,
				InheritFromAzureAD:         true,
			},
		}
		result, err := appendCloudProviderOptions(ctx, options, credentials)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(ContainElements(
			"--cloud-provider", "azure-blob-storage",
			"--credential", "default",
		))
	})
})
