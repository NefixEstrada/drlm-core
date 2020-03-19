// SPDX-License-Identifier: AGPL-3.0-only

package migrations

import (
	"sync"
	"time"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gormigrate.v1"
)

// Migrate runs all the DB migrations
func Migrate(ctx *context.Context) {
	m := gormigrate.New(ctx.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201907161002",
			Migrate: func(tx *gorm.DB) error {
				type Type int

				type User struct {
					gorm.Model
					Username string `gorm:"unique;not null"`
					Password string `gorm:"not null"`
					AuthType Type   `gorm:"not null"`
				}

				return tx.AutoMigrate(&User{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("users").Error
			},
		},
		{
			ID: "201910161152",
			Migrate: func(tx *gorm.DB) error {
				type Arch int
				type OS int

				type Agent struct {
					gorm.Model
					Host     string `gorm:"unique;not null"`
					Accepted bool   `gorm:"not null"`

					MinioKey string `gorm:"not null"`
					Secret   string `gorm:"unique;not null"` // The secret is used for authentication

					SSHPort     int `gorm:"not null"`
					SSHUser     string
					SSHHostKeys string `gorm:"size:9999"` // The different keys are splitted with `|||` between each one

					Version       string
					Arch          Arch
					OS            OS
					OSVersion     string
					Distro        string
					DistroVersion string
				}

				return tx.AutoMigrate(&Agent{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("agents").Error
			},
		},
		{
			ID: "201910161153",
			Migrate: func(tx *gorm.DB) error {
				type JobStatus int

				type Job struct {
					gorm.Model

					PluginID   uint      `gorm:"not null"`
					AgentHost  string    `gorm:"not null"`
					Status     JobStatus `gorm:"not null"`
					Time       time.Time
					Config     string `gorm:"not null"`
					BucketName string `gorm:"not null;unique"`
					Info       string

					Mux            sync.Mutex `gorm:"-"`
					ReconnAttempts int
				}

				return tx.AutoMigrate(&Job{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("jobs").Error
			},
		},
		{
			ID: "202001271226",
			Migrate: func(tx *gorm.DB) error {
				type Arch int
				type OS int

				type Agent struct {
					gorm.Model
					Host     string `gorm:"unique;not null"`
					Accepted bool   `gorm:"not null"`

					MinioKey string `gorm:"not null"`
					Secret   string `gorm:"unique;not null"` // The secret is used for authentication

					SSHPort     int `gorm:"not null"`
					SSHUser     string
					SSHHostKeys string `gorm:"size:9999"` // The different keys are splitted with `|||` between each one

					Version       string
					Arch          Arch
					OS            OS
					OSVersion     string
					Distro        string
					DistroVersion string
				}

				type Plugin struct {
					gorm.Model
					Repo      string `gorm:"not null"`
					Name      string `gorm:"not null"`
					Version   string `gorm:"not null"`
					AgentHost string `gorm:"not null"`
					Agent     *Agent `gorm:"foreignkey:Host;association_foreignkey:AgentHost"`
					// TODO: This should be stored in the DB
					Arch []Arch `gorm:"-"`
					OS   []OS   `gorm:"-"`
				}

				return tx.AutoMigrate(&Plugin{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("plugins").Error
			},
		},
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("error running the DB migrations: %v", err)
	}

	log.Info("successfully run the DB migrations")
}
