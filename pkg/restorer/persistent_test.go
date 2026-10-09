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
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("persistent WAL objects", func() {
	It("builds the barman gzip key from the destination prefix and server name", func() {
		key := walObjectKey(
			"archive-prefix",
			"pg-server-a",
			"0000000200000091000000E0",
			"gzip",
		)
		Expect(key).To(Equal(
			"archive-prefix/pg-server-a/wals/0000000200000091/0000000200000091000000E0.gz",
		))
	})

	It("splits an s3 destination path", func() {
		bucket, prefix, err := splitS3URL("s3://wal-bucket/archive-prefix")
		Expect(err).NotTo(HaveOccurred())
		Expect(bucket).To(Equal("wal-bucket"))
		Expect(prefix).To(Equal("archive-prefix"))
	})

	It("gunzips a segment and writes it in one rename", func() {
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		plain := bytes.Repeat([]byte("a"), walSegmentBytes)
		_, err := writer.Write(plain)
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.Close()).To(Succeed())

		decoded, err := decodeWAL(compressed.Bytes(), "gzip")
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal(plain))

		dir := GinkgoT().TempDir()
		dest := filepath.Join(dir, "RECOVERYXLOG")
		Expect(writeAtomically(dest, decoded)).To(Succeed())
		written, err := os.ReadFile(dest)
		Expect(err).NotTo(HaveOccurred())
		Expect(written).To(Equal(plain))
		_, err = os.Stat(dest + ".partial")
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("leaves history files on the barman-cloud-wal-restore path", func() {
		Expect(isWALSegment("0000000200000091000000E0")).To(BeTrue())
		Expect(isWALSegment("00000002.history")).To(BeFalse())
	})
})
