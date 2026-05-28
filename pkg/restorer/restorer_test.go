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

package restorer

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("errorForExitCode", func() {
	const walName = "000000010000000000000001"

	DescribeTable(
		"wraps the expected sentinel and remains identifiable via errors.Is",
		func(exitCode int, expectedSentinel error) {
			err := errorForExitCode(exitCode, walName)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, expectedSentinel)).
				To(BeTrue(), "expected errors.Is to find %v in %q", expectedSentinel, err)
		},
		// Exit codes match the barman documentation referenced in
		// errorForExitCode; the per-Entry text labels each one.
		Entry("exit 1 -> ErrWALNotFound", 1, ErrWALNotFound),
		Entry("exit 2 -> ErrConnectivity", 2, ErrConnectivity),
		Entry("exit 3 -> ErrInvalidWALName", 3, ErrInvalidWALName),
		Entry("exit 4 -> ErrGeneric", 4, ErrGeneric),
		Entry("unrecognized exit code -> ErrUnrecognizedExitCode", 99, ErrUnrecognizedExitCode),
	)

	It("includes the WAL name in the messages that mention it", func() {
		// The two branches that interpolate walName must keep doing so;
		// the others reference the command name instead.
		Expect(errorForExitCode(1, walName).Error()).To(ContainSubstring(walName))
		Expect(errorForExitCode(3, walName).Error()).To(ContainSubstring(walName))
	})

	It("survives further wrapping via fmt.Errorf %w", func() {
		// Downstream callers (operator, plugins) often wrap the restorer's
		// error with additional context. errors.Is must still resolve the
		// original sentinel through any number of wraps.
		base := errorForExitCode(2, walName)
		outer := errors.Join(errors.New("outer context"), base)
		Expect(errors.Is(outer, ErrConnectivity)).To(BeTrue())
	})
})
