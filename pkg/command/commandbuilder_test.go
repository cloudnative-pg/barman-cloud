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
	"os"
	"strings"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"

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

	It("should be false if ctx contains an invalid value and not overwritten by env", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, "invalidValue")
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be false if ctx contains false value and not overwritten by env", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be true only if ctx contains true value and not overwritten by env", func(ctx SpecContext) {
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})

	// Env var should override the ctx value

	It("should be true if env var set to true even if ctx contains false value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "true")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})

	It("should be false if env var set to false even if ctx contains true value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "false")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be false if env var is empty if ctx contains false value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be true if env var is empty if ctx contains true value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})

	// Env var should override the ctx value

	It("should be true if env var set to true even if ctx contains false value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "true")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})

	It("should be false if env var set to false even if ctx contains true value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "false")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be false if env var is empty if ctx contains false value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, false)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})

	It("should be true if env var is empty if ctx contains true value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeTrue())
	})

	It("should be false if env var set with invalid value", func(ctx SpecContext) {
		os.Setenv(barmanUseDefaultAzureCredentials, "invalidValue")
		newCtx := context.WithValue(ctx, contextKeyUseDefaultAzureCredentials, true)
		Expect(useDefaultAzureCredentials(newCtx)).To(BeFalse())
	})
})
