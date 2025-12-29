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
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// fadviseNotUsed issues an fadvise to the OS to inform that the file is not needed anymore.
// This is necessary because we run in a separate container from PostgreSQL in Kubernetes.
// Without this hint, archived WALs accumulate in page cache since memory pressure on large
// machines is typically insufficient to evict them, wasting memory that could be used for
// active workloads. PostgreSQL handles its own cache management, but our archiver sidecar
// needs to do the same for the files it processes.
func (archiver *BarmanArchiver) fadviseNotUsed(fileName string) (err error) {
	file, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return fmt.Errorf("error opening file %s for fadvise: %w", fileName, err)
	}

	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("error closing file %s for fadvise: %w", fileName, closeErr)
		}
	}(file)

	fd := int(file.Fd())
	if fadviseErr := unix.Fadvise(fd, 0, 0, unix.FADV_DONTNEED); fadviseErr != nil {
		return fmt.Errorf("error issuing fadvise on file %s: %w", fileName, fadviseErr)
	}

	return nil
}
