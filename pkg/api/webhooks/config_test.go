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
