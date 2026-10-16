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

package api

import (
	"reflect"

	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// referencingAllSecrets fills every SecretKeySelector field of the passed
// struct pointer, and returns the names it used. It fails when a new field is
// not listed in the SecretNames method of the type, which would leave
// the consumers unable to read that secret.
func referencingAllSecrets(credentials any) []string {
	var names []string
	value := reflect.ValueOf(credentials).Elem()
	for i := range value.NumField() {
		if value.Field(i).Type() != reflect.TypeFor[*machineryapi.SecretKeySelector]() {
			continue
		}
		name := value.Type().Field(i).Name
		value.Field(i).Set(reflect.ValueOf(&machineryapi.SecretKeySelector{
			LocalObjectReference: machineryapi.LocalObjectReference{Name: name},
		}))
		names = append(names, name)
	}
	return names
}

var _ = Describe("Credentials secret names", func() {
	It("lists every secret referenced by the S3 credentials", func() {
		credentials := &S3Credentials{}
		names := referencingAllSecrets(credentials)
		Expect(names).To(ContainElement("SSECustomerKey"))
		Expect(credentials.SecretNames()).To(ConsistOf(names))
	})

	It("lists every secret referenced by the Azure credentials", func() {
		credentials := &AzureCredentials{}
		names := referencingAllSecrets(credentials)
		Expect(credentials.SecretNames()).To(ConsistOf(names))
	})

	It("lists no Azure secret when the pod identity is used", func() {
		credentials := &AzureCredentials{InheritFromAzureAD: true}
		referencingAllSecrets(credentials)
		Expect(credentials.SecretNames()).To(BeEmpty())
	})

	It("lists every secret referenced by the Google credentials", func() {
		credentials := &GoogleCredentials{}
		names := referencingAllSecrets(credentials)
		Expect(credentials.SecretNames()).To(ConsistOf(names))
	})

	It("collects the secrets of all the providers, and tolerates nil ones", func() {
		Expect(BarmanCredentials{}.SecretNames()).To(BeEmpty())

		credentials := BarmanCredentials{AWS: &S3Credentials{}, Google: &GoogleCredentials{}}
		names := referencingAllSecrets(credentials.AWS)
		names = append(names, referencingAllSecrets(credentials.Google)...)
		Expect(credentials.SecretNames()).To(ConsistOf(names))
	})
})
