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

// Package catalog is the implementation of a backup catalog
package catalog

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/cloudnative-pg/machinery/pkg/types"
)

// Catalog is a list of backup infos belonging to the same server
type Catalog struct {
	// The list of backups
	List []BarmanBackup `json:"backups_list"`
}

// NewCatalogFromBarmanCloudBackupList parses the output of barman-cloud-backup-list
func NewCatalogFromBarmanCloudBackupList(rawJSON string) (*Catalog, error) {
	result := &Catalog{}
	err := json.Unmarshal([]byte(rawJSON), result)
	if err != nil {
		return nil, err
	}

	for idx := range result.List {
		if err := result.List[idx].deserializeBackupTimeStrings(); err != nil {
			return nil, err
		}
	}

	// Sort the list of backups in order of time
	sort.Sort(result)

	return result, nil
}

var currentTLIRegex = regexp.MustCompile("^(|latest)$")

// LatestBackupInfo gets the information about the latest successful backup
func (catalog *Catalog) LatestBackupInfo() *BarmanBackup {
	if catalog.Len() == 0 {
		return nil
	}

	// the code below assumes the catalog to be sorted, therefore, we enforce it first
	sort.Sort(catalog)

	// Skip errored backups and return the latest valid one
	for i := len(catalog.List) - 1; i >= 0; i-- {
		if catalog.List[i].isBackupDone() {
			return &catalog.List[i]
		}
	}

	return nil
}

// GetLastSuccessfulBackupTime gets the end time of the last successful backup or nil if no backup was successful
func (catalog *Catalog) GetLastSuccessfulBackupTime() *time.Time {
	var lastSuccessfulBackup *time.Time
	if lastSuccessfulBackupInfo := catalog.LatestBackupInfo(); lastSuccessfulBackupInfo != nil {
		return &lastSuccessfulBackupInfo.EndTime
	}
	return lastSuccessfulBackup
}

// GetBackupIDs returns the list of backup IDs in the catalog
func (catalog *Catalog) GetBackupIDs() []string {
	backupIDs := make([]string, len(catalog.List))
	for idx, barmanBackup := range catalog.List {
		backupIDs[idx] = barmanBackup.ID
	}
	return backupIDs
}

// FirstRecoverabilityPoint gets the start time of the first backup in
// the catalog
func (catalog *Catalog) FirstRecoverabilityPoint() *time.Time {
	if catalog.Len() == 0 {
		return nil
	}

	// the code below assumes the catalog to be sorted, therefore, we enforce it first
	sort.Sort(catalog)

	// Skip errored backups and return the first valid one
	for i := 0; i < len(catalog.List); i++ {
		if !catalog.List[i].isBackupDone() {
			continue
		}

		return &catalog.List[i].EndTime
	}

	return nil
}

// GetFirstRecoverabilityPoint see FirstRecoverabilityPoint. This is needed to adhere to the common backup interface.
func (catalog *Catalog) GetFirstRecoverabilityPoint() *time.Time {
	return catalog.FirstRecoverabilityPoint()
}

// GetBackupMethod returns the backup method
func (catalog Catalog) GetBackupMethod() string {
	return "barmanObjectStore"
}

type recoveryTargetAdapter interface {
	GetBackupID() string
	GetTargetTime() string
	GetTargetLSN() string
	GetTargetTLI() string
}

// FindBackupInfo finds the backup info that should be used to file
// a PITR request via target parameters specified within `RecoveryTarget`
func (catalog *Catalog) FindBackupInfo(
	recoveryTarget recoveryTargetAdapter,
) (*BarmanBackup, error) {
	// Check that BackupID is not empty. In such case, always use the
	// backup ID provided by the user.
	if recoveryTarget.GetBackupID() != "" {
		return catalog.findBackupFromID(recoveryTarget.GetBackupID())
	}

	// The user has not specified any backup ID. As a result we need
	// to automatically detect the backup from which to start the
	// recovery process.

	// Set the timeline
	targetTLI := recoveryTarget.GetTargetTLI()

	// Sort the catalog, as that's what the code below expects
	sort.Sort(catalog)

	// The first step is to check any time based research
	if t := recoveryTarget.GetTargetTime(); t != "" {
		return catalog.findClosestBackupFromTargetTime(t, targetTLI)
	}

	// The second step is to check any LSN based research
	if t := recoveryTarget.GetTargetLSN(); t != "" {
		return catalog.findClosestBackupFromTargetLSN(t, targetTLI)
	}

	// The fallback is to use the latest available backup in chronological order
	return catalog.findLatestBackupFromTimeline(targetTLI), nil
}

func (catalog *Catalog) findClosestBackupFromTargetLSN(
	targetLSNString string,
	targetTLI string,
) (*BarmanBackup, error) {
	targetLSN := types.LSN(targetLSNString)
	if _, err := targetLSN.Parse(); err != nil {
		return nil, fmt.Errorf("while parsing recovery target targetLSN: %s", err.Error())
	}
	for i := len(catalog.List) - 1; i >= 0; i-- {
		barmanBackup := catalog.List[i]
		if !barmanBackup.isBackupDone() {
			continue
		}
		if (strconv.Itoa(barmanBackup.TimeLine) == targetTLI ||
			// if targetTLI is not an integer, it will be ignored actually
			currentTLIRegex.MatchString(targetTLI)) &&
			types.LSN(barmanBackup.EndLSN).Less(targetLSN) {
			return &catalog.List[i], nil
		}
	}
	return nil, nil
}

func (catalog *Catalog) findClosestBackupFromTargetTime(
	targetTimeString string,
	targetTLI string,
) (*BarmanBackup, error) {
	targetTime, err := types.ParseTargetTime(nil, targetTimeString)
	if err != nil {
		return nil, fmt.Errorf("while parsing recovery target targetTime: %s", err.Error())
	}
	for i := len(catalog.List) - 1; i >= 0; i-- {
		barmanBackup := catalog.List[i]
		if !barmanBackup.isBackupDone() {
			continue
		}
		if (strconv.Itoa(barmanBackup.TimeLine) == targetTLI ||
			// if targetTLI is not an integer, it will be ignored actually
			currentTLIRegex.MatchString(targetTLI)) &&
			!barmanBackup.EndTime.After(targetTime) {
			return &catalog.List[i], nil
		}
	}
	return nil, nil
}

func (catalog *Catalog) findLatestBackupFromTimeline(targetTLI string) *BarmanBackup {
	for i := len(catalog.List) - 1; i >= 0; i-- {
		barmanBackup := catalog.List[i]
		if !barmanBackup.isBackupDone() {
			continue
		}
		if strconv.Itoa(barmanBackup.TimeLine) == targetTLI ||
			// if targetTLI is not an integer, it will be ignored actually
			currentTLIRegex.MatchString(targetTLI) {
			return &catalog.List[i]
		}
	}

	return nil
}

func (catalog *Catalog) findBackupFromID(backupID string) (*BarmanBackup, error) {
	if backupID == "" {
		return nil, fmt.Errorf("no backupID provided")
	}
	for _, barmanBackup := range catalog.List {
		if !barmanBackup.isBackupDone() {
			continue
		}
		if barmanBackup.ID == backupID {
			return &barmanBackup, nil
		}
	}
	return nil, fmt.Errorf("no backup found with ID %s", backupID)
}

// BarmanBackup represent a backup as created
// by Barman
type BarmanBackup struct {
	// The backup name, can be used as a way to identify a backup.
	// Populated only if the backup was executed with barman 3.3.0+.
	BackupName string `json:"backup_name,omitempty"`

	// The backup label
	Label string `json:"backup_label"`

	// The moment where the backup started
	BeginTimeString string `json:"begin_time"`

	// The moment where the backup ended
	EndTimeString string `json:"end_time"`

	// The moment where the backup started in ISO format
	BeginTimeISOString string `json:"begin_time_iso"`

	// The moment where the backup ended in ISO format
	EndTimeISOString string `json:"end_time_iso"`

	// The moment where the backup ended
	BeginTime time.Time

	// The moment where the backup ended
	EndTime time.Time

	// The WAL where the backup started
	BeginWal string `json:"begin_wal"`

	// The WAL where the backup ended
	EndWal string `json:"end_wal"`

	// The LSN where the backup started
	BeginLSN string `json:"begin_xlog"`

	// The LSN where the backup ended
	EndLSN string `json:"end_xlog"`

	// The systemID of the cluster
	SystemID string `json:"systemid"`

	// The ID of the backup
	ID string `json:"backup_id"`

	// The error output if present
	Error string `json:"error"`

	// The TimeLine
	TimeLine int `json:"timeline"`
}

type barmanBackupShow struct {
	Cloud BarmanBackup `json:"cloud,omitempty"`
}

// NewBackupFromBarmanCloudBackupShow parses the output of barman-cloud-backup-show
func NewBackupFromBarmanCloudBackupShow(rawJSON string) (*BarmanBackup, error) {
	result := &barmanBackupShow{}
	err := json.Unmarshal([]byte(rawJSON), result)
	if err != nil {
		return nil, err
	}

	if err := result.Cloud.deserializeBackupTimeStrings(); err != nil {
		return nil, err
	}

	return &result.Cloud, nil
}

// barmanTimeLayout is the format that is being used to parse
// the backupInfo from barman-cloud-backup-list
const (
	barmanTimeLayout = "Mon Jan 2 15:04:05 2006"
)

func (b *BarmanBackup) deserializeBackupTimeStrings() error {
	var err error
	b.BeginTime, err = tryParseISOOrCtimeTime(b.BeginTimeISOString, b.BeginTimeString)
	if err != nil {
		return err
	}

	b.EndTime, err = tryParseISOOrCtimeTime(b.EndTimeISOString, b.EndTimeString)
	if err != nil {
		return err
	}

	return nil
}

func tryParseISOOrCtimeTime(isoValue, ctimeOrISOValue string) (time.Time, error) {
	if isoValue != "" {
		return time.Parse(time.RFC3339, isoValue)
	}

	if ctimeOrISOValue != "" {
		// Barman 3.12.0 incorrectly puts an ISO-formatted time in the ctime-formatted field.
		// So in case of parsing failure we try again parsing it as an ISO time,
		// discarding an eventual failure
		return parseTimeWithFallbackLayout(ctimeOrISOValue, barmanTimeLayout, time.RFC3339)
	}

	return time.Time{}, nil
}

func parseTimeWithFallbackLayout(value string, primaryLayout string, fallbackLayout string) (time.Time, error) {
	result, err := time.Parse(primaryLayout, value)
	if err == nil {
		return result, nil
	}

	result, errFallback := time.Parse(fallbackLayout, value)
	if errFallback == nil {
		return result, nil
	}

	return result, err
}

func (b *BarmanBackup) isBackupDone() bool {
	return !b.BeginTime.IsZero() && !b.EndTime.IsZero()
}

// NewCatalog creates a new sorted backup catalog, given a list of backup infos
// belonging to the same server.
func NewCatalog(list []BarmanBackup) *Catalog {
	result := &Catalog{
		List: list,
	}
	sort.Sort(result)

	return result
}
