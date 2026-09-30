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

package credentials

import (
	"os"
	"strings"

	machineryapi "github.com/cloudnative-pg/machinery/pkg/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	barmanApi "github.com/cloudnative-pg/barman-cloud/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Google credentials", func() {
	const namespace = "default"
	var c client.Client

	newSecret := func(name, content string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Data:       map[string][]byte{"credentials": []byte(content)},
		}
	}
	credentialsFrom := func(secretName string) *barmanApi.GoogleCredentials {
		return &barmanApi.GoogleCredentials{
			ApplicationCredentials: &machineryapi.SecretKeySelector{
				LocalObjectReference: machineryapi.LocalObjectReference{Name: secretName},
				Key:                  "credentials",
			},
		}
	}
	credentialsFile := func(env []string) string {
		for _, entry := range env {
			if value, found := strings.CutPrefix(entry, "GOOGLE_APPLICATION_CREDENTIALS="); found {
				return value
			}
		}
		return ""
	}

	BeforeEach(func() {
		previousDirectory := googleCredentialsDirectory
		googleCredentialsDirectory = GinkgoT().TempDir()
		DeferCleanup(func() { googleCredentialsDirectory = previousDirectory })

		c = fake.NewClientBuilder().WithObjects(
			newSecret("store-a", "credentials-a"),
			newSecret("store-b", "credentials-b"),
		).Build()
	})

	It("keeps the credentials of each object store in its own file", func(ctx SpecContext) {
		envA, err := envSetGoogleCredentials(ctx, c, namespace, credentialsFrom("store-a"), nil)
		Expect(err).ToNot(HaveOccurred())
		envB, err := envSetGoogleCredentials(ctx, c, namespace, credentialsFrom("store-b"), nil)
		Expect(err).ToNot(HaveOccurred())

		fileA, fileB := credentialsFile(envA), credentialsFile(envB)
		Expect(fileA).ToNot(Equal(fileB))
		Expect(os.ReadFile(fileA)).To(BeEquivalentTo("credentials-a")) // #nosec G304
		Expect(os.ReadFile(fileB)).To(BeEquivalentTo("credentials-b")) // #nosec G304
	})

	It("leaves the other object stores' credentials alone in a GKE environment", func(ctx SpecContext) {
		envA, err := envSetGoogleCredentials(ctx, c, namespace, credentialsFrom("store-a"), nil)
		Expect(err).ToNot(HaveOccurred())

		envGKE, err := envSetGoogleCredentials(ctx, c, namespace,
			&barmanApi.GoogleCredentials{GKEEnvironment: true}, []string{"FOO=bar"})
		Expect(err).ToNot(HaveOccurred())
		Expect(envGKE).To(Equal([]string{"FOO=bar"}))
		Expect(os.ReadFile(credentialsFile(envA))).To(BeEquivalentTo("credentials-a"))
	})

	It("fails when the referenced secret does not exist", func(ctx SpecContext) {
		_, err := envSetGoogleCredentials(ctx, c, namespace, credentialsFrom("missing"), nil)
		Expect(err).To(HaveOccurred())
	})
})
