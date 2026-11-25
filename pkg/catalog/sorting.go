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

package catalog

// Len implements sort.Interface
func (catalog *Catalog) Len() int {
	return len(catalog.List)
}

// Less implements sort.Interface: completed backups first, then the
// incomplete ones, each ordered by begin time and then by ID.
func (catalog *Catalog) Less(i, j int) bool {
	a, b := &catalog.List[i], &catalog.List[j]

	aCompleted := !a.BeginTime.IsZero() && !a.EndTime.IsZero()
	bCompleted := !b.BeginTime.IsZero() && !b.EndTime.IsZero()
	if aCompleted != bCompleted {
		return aCompleted
	}

	if !a.BeginTime.Equal(b.BeginTime) {
		return a.BeginTime.Before(b.BeginTime)
	}

	return a.ID < b.ID
}

// Swap implements sort.Interface
func (catalog *Catalog) Swap(i, j int) {
	catalog.List[j], catalog.List[i] = catalog.List[i], catalog.List[j]
}
