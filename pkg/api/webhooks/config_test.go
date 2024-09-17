package webhooks

import (
	api "github.com/cloudnative-pg/barman-cloud/pkg/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type resource struct {
	Spec resourceSpec
}
type resourceSpec struct {
	Backup *backupConfiguration
}
type backupConfiguration struct {
	BarmanObjectStore *api.BarmanObjectStoreConfiguration
	RetentionPolicy   string
}

func (r resource) GetBarmanObjectStore() *api.BarmanObjectStoreConfiguration {
	if r.Spec.Backup == nil {
		return nil
	}
	return r.Spec.Backup.BarmanObjectStore
}

func (r resource) GetRetentionPolicy() string {
	if r.Spec.Backup == nil {
		return ""
	}
	return r.Spec.Backup.RetentionPolicy
}

func (r resource) GetBarmanObjectStorePath() []string {
	return []string{"spec", "backupConfiguration", "barmanObjectStore"}
}

func (r resource) GetRetentionPolicyPath() []string {
	return []string{"spec", "backupConfiguration", "retentionPolicy"}
}

var _ = Describe("Backup validation", func() {
	It("complain if there's no credentials", func() {
		res := &resource{
			Spec: resourceSpec{
				Backup: &backupConfiguration{
					BarmanObjectStore: &api.BarmanObjectStoreConfiguration{},
				},
			},
		}
		err := ValidateBackupConfiguration(res)
		Expect(err).To(HaveLen(1))
	})

	It("doesn't complain if given policy is not provided", func() {
		res := &resource{
			Spec: resourceSpec{
				Backup: &backupConfiguration{},
			},
		}
		err := ValidateBackupConfiguration(res)
		Expect(err).To(BeNil())
	})

	It("doesn't complain if given policy is valid", func() {
		res := &resource{
			Spec: resourceSpec{
				Backup: &backupConfiguration{
					RetentionPolicy: "90d",
				},
			},
		}
		err := ValidateBackupConfiguration(res)
		Expect(err).To(BeNil())
	})

	It("complain if a given policy is not valid", func() {
		res := &resource{
			Spec: resourceSpec{
				Backup: &backupConfiguration{
					BarmanObjectStore: &api.BarmanObjectStoreConfiguration{},
					RetentionPolicy:   "09",
				},
			},
		}
		err := ValidateBackupConfiguration(res)
		Expect(err).To(HaveLen(2))
	})
})
