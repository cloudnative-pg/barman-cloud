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

package spool

import (
	"os"
	"path"

	"github.com/cloudnative-pg/machinery/pkg/fileutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spool", func() {
	var tmpDir string
	var tmpDir2 string
	var spool *WALSpool

	_ = BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "spool-test-")
		Expect(err).NotTo(HaveOccurred())

		tmpDir2, err = os.MkdirTemp("", "spool-test-tmp-")
		Expect(err).NotTo(HaveOccurred())

		spool, err = New(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	_ = AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir2)).To(Succeed())
	})

	It("create and removes files from/into the spool", func() {
		var err error
		const walFile = "000000020000068A00000002"

		// This WAL file doesn't exist
		Expect(spool.Contains(walFile)).To(BeFalse())

		// If I try to remove a WAL file that doesn't exist, I obtain an error
		err = spool.Remove(walFile)
		Expect(err).To(Equal(ErrorNonExistentFile))

		// I add it into the spool
		err = spool.Touch(walFile)
		Expect(err).NotTo(HaveOccurred())

		// Now the file exists
		Expect(spool.Contains(walFile)).To(BeTrue())

		// I can now remove it
		err = spool.Remove(walFile)
		Expect(err).NotTo(HaveOccurred())

		// And now it doesn't exist again
		Expect(spool.Contains(walFile)).To(BeFalse())
	})

	It("can move out files from the spool", func() {
		var err error
		const walFile = "000000020000068A00000003"

		err = spool.Touch(walFile)
		Expect(err).ToNot(HaveOccurred())

		// Move out this file
		destinationPath := path.Join(tmpDir2, "testFile")
		err = spool.MoveOut(walFile, destinationPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(spool.Contains(walFile)).To(BeFalse())
		Expect(fileutils.FileExists(destinationPath)).To(BeTrue())
	})

	It("can determine names for each WAL files", func() {
		const walFile = "000000020000068A00000004"
		Expect(spool.FileName(walFile)).To(Equal(path.Join(tmpDir, walFile)))
	})

	It("returns temp file path with .tmp suffix", func() {
		const walFile = "000000020000068A00000005"
		tempPath := spool.TempFileName(walFile)
		Expect(tempPath).To(Equal(path.Join(tmpDir, walFile+".tmp")))
	})

	It("commits a temp file to its final location", func() {
		const walFile = "000000020000068A00000006"

		// Create a temp file with some content
		tempPath := spool.TempFileName(walFile)
		err := os.WriteFile(tempPath, []byte("test content"), 0o600)
		Expect(err).ToNot(HaveOccurred())

		// Temp file exists, final file does not
		Expect(fileutils.FileExists(tempPath)).To(BeTrue())
		Expect(spool.Contains(walFile)).To(BeFalse())

		// Commit the file
		err = spool.Commit(walFile)
		Expect(err).ToNot(HaveOccurred())

		// Now final file exists, temp file does not
		Expect(spool.Contains(walFile)).To(BeTrue())
		tempExists, _ := fileutils.FileExists(tempPath)
		Expect(tempExists).To(BeFalse())

		// Verify content was preserved
		content, err := os.ReadFile(spool.FileName(walFile))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal("test content"))
	})

	It("returns error when committing non-existent temp file", func() {
		const walFile = "000000020000068A00000007"

		err := spool.Commit(walFile)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to commit WAL file"))
	})

	It("cleans up temp files", func() {
		const walFile = "000000020000068A00000008"

		// Create a temp file
		tempPath := spool.TempFileName(walFile)
		err := os.WriteFile(tempPath, []byte("test content"), 0o600)
		Expect(err).ToNot(HaveOccurred())
		Expect(fileutils.FileExists(tempPath)).To(BeTrue())

		// Clean it up
		spool.CleanupTemp(walFile)

		// Temp file should be gone
		tempExists, _ := fileutils.FileExists(tempPath)
		Expect(tempExists).To(BeFalse())
	})

	It("cleanup is safe on non-existent temp file", func() {
		const walFile = "000000020000068A00000009"

		// This should not panic or error
		spool.CleanupTemp(walFile)
	})

	// These two tests verify the race condition fix:
	// temp files (.tmp) must be invisible to Contains and MoveOut

	It("Contains does NOT see temp files", func() {
		const walFile = "000000020000068A0000000A"

		// Create a temp file (simulating an in-progress download)
		tempPath := spool.TempFileName(walFile)
		err := os.WriteFile(tempPath, []byte("partial content"), 0o600)
		Expect(err).ToNot(HaveOccurred())

		// Contains should return false - temp files are invisible
		Expect(spool.Contains(walFile)).To(BeFalse())

		// Clean up
		spool.CleanupTemp(walFile)
	})

	It("MoveOut does NOT see temp files", func() {
		const walFile = "000000020000068A0000000B"

		// Create a temp file (simulating an in-progress download)
		tempPath := spool.TempFileName(walFile)
		err := os.WriteFile(tempPath, []byte("partial content"), 0o600)
		Expect(err).ToNot(HaveOccurred())

		// MoveOut should fail with ErrorNonExistentFile - temp files are invisible
		destinationPath := path.Join(tmpDir2, "testFile")
		err = spool.MoveOut(walFile, destinationPath)
		Expect(err).To(Equal(ErrorNonExistentFile))

		// Destination should not exist
		destExists, _ := fileutils.FileExists(destinationPath)
		Expect(destExists).To(BeFalse())

		// Clean up
		spool.CleanupTemp(walFile)
	})
})
