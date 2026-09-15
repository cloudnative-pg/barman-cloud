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
	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"
	"k8s.io/apimachinery/pkg/util/validation/field"

	api "github.com/cloudnative-pg/barman-cloud/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup validation", func() {
	It("complain if there's no credentials", func() {
		err := ValidateBackupConfiguration(
			&api.BarmanObjectStoreConfiguration{},
			field.NewPath("spec", "backupConfiguration", "retentionPolicy"))
		Expect(err).To(HaveLen(1))
	})

	It("doesn't complain if given policy is not provided", func() {
		err := ValidateBackupConfiguration(nil, nil)
		Expect(err).To(BeEmpty())
	})

	Context("with an SSE-C customer key", func() {
		var configuration *api.BarmanObjectStoreConfiguration
		BeforeEach(func() {
			configuration = &api.BarmanObjectStoreConfiguration{
				BarmanCredentials: api.BarmanCredentials{
					AWS: &api.S3Credentials{
						InheritFromIAMRole: true,
						SSECustomerKey: &machineryapi.SecretKeySelector{
							LocalObjectReference: machineryapi.LocalObjectReference{Name: "backup-keys"},
							Key:                  "sse-c",
						},
					},
				},
			}
		})

		It("accepts it on its own", func() {
			err := ValidateBackupConfiguration(configuration, field.NewPath("spec"))
			Expect(err).To(BeEmpty())
		})

		It("rejects it together with data or wal encryption", func() {
			configuration.Data = &api.DataBackupConfiguration{Encryption: api.EncryptionTypeAES256}
			configuration.Wal = &api.WalBackupConfiguration{Encryption: api.EncryptionTypeAES256}
			err := ValidateBackupConfiguration(configuration, field.NewPath("spec"))
			Expect(err).To(HaveLen(2))
			Expect(err[0].Field).To(Equal("spec.data.encryption"))
			Expect(err[1].Field).To(Equal("spec.wal.encryption"))
		})
	})
})

var _ = Describe("Retention Policy Validation", func() {
	It("doesn't complain if given policy is valid", func() {
		err := ValidateRetentionPolicy("90d", field.NewPath("spec", "backup", "retentionPolicy"))
		Expect(err).To(BeEmpty())
	})

	It("complain if a given policy is not valid", func() {
		err := ValidateRetentionPolicy("09", field.NewPath("spec", "backup", "retentionPolicy"))
		Expect(err).To(HaveLen(1))
	})
})
