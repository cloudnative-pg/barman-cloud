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

package archiver

import (
	"strings"

	barmanTypes "github.com/cloudnative-pg/plugin-barman-cloud/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("barmanCloudWalArchiveOptions", func() {
	var config *barmanTypes.BarmanObjectStoreConfiguration

	BeforeEach(func() {
		config = &barmanTypes.BarmanObjectStoreConfiguration{
			DestinationPath: "s3://bucket-name/",
			Wal: &barmanTypes.WalBackupConfiguration{
				Compression: "gzip",
				Encryption:  "aes256",
			},
		}
	})

	It("should generate correct arguments", func(ctx SpecContext) {
		archiver, err := New(ctx, nil, "spool", "pgdata", nil)
		Expect(err).ToNot(HaveOccurred())

		extraOptions := []string{"--min-chunk-size=5MB", "--read-timeout=60", "-vv"}
		config.Wal.ArchiveAdditionalCommandArgs = extraOptions
		options, err := archiver.BarmanCloudWalArchiveOptions(ctx, config, "test-cluster")
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--gzip -e aes256 --min-chunk-size=5MB --read-timeout=60 -vv s3://bucket-name/ test-cluster",
				))
	})

	It("should not overwrite declared options if conflict", func(ctx SpecContext) {
		extraOptions := []string{
			"--min-chunk-size=5MB",
			"--read-timeout=60",
			"-vv",
			"--immediate-checkpoint=false",
			"--gzip",
			"-e",
			"aes256",
		}
		config.Wal.ArchiveAdditionalCommandArgs = extraOptions

		archiver, err := New(
			ctx, nil, "spool", "pgdata", nil)
		Expect(err).ToNot(HaveOccurred())

		options, err := archiver.BarmanCloudWalArchiveOptions(ctx, config, "test-cluster")
		Expect(err).ToNot(HaveOccurred())

		Expect(strings.Join(options, " ")).
			To(
				Equal(
					"--gzip -e aes256 --min-chunk-size=5MB --read-timeout=60 " +
						"-vv --immediate-checkpoint=false s3://bucket-name/ test-cluster",
				))
	})
})
