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

// Package capabilities stores the definition of the type for Barman capabilities
package capabilities

import (
	"github.com/blang/semver"
)

// Capabilities collects a set of boolean values that shows the possible capabilities of Barman and the version
type Capabilities struct {
	Version *semver.Version
	// this is not exported because the consumers have to use ShouldExecuteBackupWithName
	hasName                    bool
	HasAzure                   bool
	HasS3                      bool
	HasGoogle                  bool
	HasRetentionPolicy         bool
	HasTags                    bool
	HasCheckWalArchive         bool
	HasSnappy                  bool
	HasZSTD                    bool
	HasLZ4                     bool
	HasXZ                      bool
	HasErrorCodesForWALRestore bool
	HasErrorCodesForRestore    bool
	HasAzureManagedIdentity    bool
}

// LegacyExecutor allows this code to know
// if a legacy backup should be forced or not
type LegacyExecutor interface {
	ShouldForceLegacyBackup() bool
}

// ShouldExecuteBackupWithName returns true if the new backup logic should be executed
func (c *Capabilities) ShouldExecuteBackupWithName(exec LegacyExecutor) bool {
	if exec.ShouldForceLegacyBackup() {
		return false
	}

	return c.hasName
}
