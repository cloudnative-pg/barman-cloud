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

package backup

import (
	"strings"

	"k8s.io/utils/ptr"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetBarmanCloudBackupOptions", func() {
	var backupCommand *Command

	BeforeEach(func() {
		config := &barmanApi.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket-name/",
			Data: &barmanApi.DataBackupConfiguration{
				Compression:         "gzip",
				Encryption:          "aes256",
				ImmediateCheckpoint: true,
				Jobs:                ptr.To(int32(4)),
			},
		}
		backupCommand = &Command{configuration: config}
	})

	It("should generate correct arguments", func(ctx SpecContext) {
		extraOptions := []string{"--min-chunk-size=5MB", "--read-timeout=60", "-vv"}
		backupCommand.configuration.Data.AdditionalCommandArgs = extraOptions

		options, err := backupCommand.GetBarmanCloudBackupOptions(ctx, "test-backup", "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--user postgres --name test-backup " +
						"--gzip --encryption aes256 --immediate-checkpoint --jobs 4 " +
						"--min-chunk-size=5MB --read-timeout=60 -vv " +
						"s3://bucket-name/ test-cluster",
				))
	})

	It("should not overwrite declared options if conflict", func(ctx SpecContext) {
		extraOptions := []string{
			"--min-chunk-size=5MB",
			"--read-timeout=60",
			"-vv",
			"--encryption=aws:kms",
			"--immediate-checkpoint=false",
		}
		backupCommand.configuration.Data.AdditionalCommandArgs = extraOptions

		options, err := backupCommand.GetBarmanCloudBackupOptions(ctx, "test-backup", "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--user postgres --name test-backup " +
						"--gzip --encryption aes256 --immediate-checkpoint --jobs 4 " +
						"--min-chunk-size=5MB --read-timeout=60 -vv " +
						"s3://bucket-name/ test-cluster",
				))
	})

	It("should append tags when set", func(ctx SpecContext) {
		backupCommand.configuration.Tags = map[string]string{"tag": "foo"}

		options, err := backupCommand.GetBarmanCloudBackupOptions(ctx, "test-backup", "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--user postgres --name test-backup " +
						"--gzip --encryption aes256 --immediate-checkpoint --jobs 4 " +
						"--tags tag,foo " +
						"s3://bucket-name/ test-cluster",
				))
	})

	It("should append the endpoint URL when set", func(ctx SpecContext) {
		backupCommand.configuration.EndpointURL = "https://my-endpoint.example.com"

		options, err := backupCommand.GetBarmanCloudBackupOptions(ctx, "test-backup", "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--user postgres --name test-backup " +
						"--gzip --encryption aes256 --immediate-checkpoint --jobs 4 " +
						"--endpoint-url https://my-endpoint.example.com " +
						"s3://bucket-name/ test-cluster",
				))
	})

	It("should append the cloud provider options when credentials are set", func(ctx SpecContext) {
		backupCommand.configuration.BarmanCredentials = barmanApi.BarmanCredentials{
			AWS: &barmanApi.S3Credentials{InheritFromIAMRole: true},
		}

		options, err := backupCommand.GetBarmanCloudBackupOptions(ctx, "test-backup", "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--user postgres --name test-backup " +
						"--gzip --encryption aes256 --immediate-checkpoint --jobs 4 " +
						"--cloud-provider aws-s3 " +
						"s3://bucket-name/ test-cluster",
				))
	})
})
