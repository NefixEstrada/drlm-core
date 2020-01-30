// SPDX-License-Identifier: AGPL-3.0-only

package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/db"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
)

// Plugin is an individual plugin that's installed in an Agent
type Plugin struct {
	gorm.Model
	Repo      string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Version   string    `gorm:"not null"`
	AgentHost string    `gorm:"not null"`
	Agent     *Agent    `gorm:"foreignkey:Host;association_foreignkey:AgentHost"`
	Arch      []os.Arch `gorm:"-"`
	OS        []os.OS   `gorm:"-"`
}

func (p *Plugin) String() string {
	return p.Repo + "/" + p.Name
}

// Add adds a new plugin in the DB
func (p *Plugin) Add() error {
	if err := db.DB.Create(p).Error; err != nil {
		return fmt.Errorf("error adding the plugin to the DB: %v", err)
	}

	return nil
}
