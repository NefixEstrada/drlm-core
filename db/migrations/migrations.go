// SPDX-License-Identifier: AGPL-3.0-only

package migrations

import (
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"

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
				return tx.AutoMigrate(&models.User{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("users").Error
			},
		},
		{
			ID: "201910161152",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Agent{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("agents").Error
			},
		},
		{
			ID: "201910161153",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Job{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("jobs").Error
			},
		},
		{
			ID: "202001271226",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Plugin{}).Error
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
