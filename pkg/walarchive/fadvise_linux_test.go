//go:build linux

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

package walarchive

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BarmanArchiver fadvise", func() {
	var archiver *BarmanArchiver
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "walarchive-test-")
		Expect(err).ToNot(HaveOccurred())

		archiver = &BarmanArchiver{}
	})

	AfterEach(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	Describe("fadviseNotUsed", func() {
		It("should succeed with a valid file", func() {
			testFile := filepath.Join(tempDir, "test-wal-file")
			err := os.WriteFile(testFile, []byte("test WAL content"), 0o600)
			Expect(err).ToNot(HaveOccurred())

			err = archiver.fadviseNotUsed(testFile)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error when file doesn't exist", func() {
			nonExistentFile := filepath.Join(tempDir, "non-existent-file")

			err := archiver.fadviseNotUsed(nonExistentFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error opening file"))
		})

		It("should handle empty files", func() {
			emptyFile := filepath.Join(tempDir, "empty-wal-file")
			err := os.WriteFile(emptyFile, []byte{}, 0o600)
			Expect(err).ToNot(HaveOccurred())

			err = archiver.fadviseNotUsed(emptyFile)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle files with path traversal characters", func() {
			testFile := filepath.Join(tempDir, "test-file")
			err := os.WriteFile(testFile, []byte("content"), 0o600)
			Expect(err).ToNot(HaveOccurred())

			// filepath.Clean should handle this safely
			fileWithDots := filepath.Join(tempDir, "..", filepath.Base(tempDir), "test-file")
			err = archiver.fadviseNotUsed(fileWithDots)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
