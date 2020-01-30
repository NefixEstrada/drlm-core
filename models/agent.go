// SPDX-License-Identifier: AGPL-3.0-only

package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/db"
	"github.com/rs/xid"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
)

// Agent (s) are the clients of DRLM Core that are installed in the servers
type Agent struct {
	gorm.Model
	Host           string `gorm:"unique;not null"`
	Port           int    `gorm:"not null"`
	User           string `gorm:"not null"`
	PublicKeyPath  string `gorm:"not null"`
	PrivateKeyPath string `gorm:"not null"`
	HostKeys       string `gorm:"size:9999;not null"` // The different keys are splitted with `|||` between each one
	MinioKey       string `gorm:"not null"`
	Secret         string `gorm:"not null"` // The secret is used for authentication

	Version       string
	Arch          os.Arch
	OS            os.OS
	OSVersion     string
	Distro        string
	DistroVersion string

	Jobs    []*Job    `gorm:"-"`
	Plugins []*Plugin `gorm:"-"`
}

// AgentList returns a list with all the agents
func AgentList() ([]*Agent, error) {
	agents := []*Agent{}

	if err := db.DB.Select("id, created_at, updated_at, host, port, user, public_key_path, private_key_path, secret, version, arch, os, os_version, distro, distro_version").Find(&agents).Error; err != nil {
		return []*Agent{}, fmt.Errorf("error getting the list of agents: %v", err)
	}

	return agents, nil
}

// Add creates a new agent in the DB
func (a *Agent) Add() error {
	if err := db.DB.Create(a).Error; err != nil {
		return fmt.Errorf("error adding the agent to the DB: %v", err)
	}

	return nil
}

// BeforeCreate is a hook that gets executed before creating an agent
func (a *Agent) BeforeCreate() error {
	if a.Secret == "" {
		a.Secret = xid.New().String()
	}

	return nil
}

// Load loads the agent from the DB using the host
func (a *Agent) Load() error {
	if err := db.DB.Where("host = ?", a.Host).First(a).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error loading the agent from the DB: %v", err)
	}

	return nil
}

// Update updates the agent in the DB
func (a *Agent) Update() error {
	if err := db.DB.Save(a).Error; err != nil {
		return fmt.Errorf("error updating the agent: %v", err)
	}

	return nil
}

// Delete removes an agent from the DB
func (a *Agent) Delete() error {
	if err := a.Load(); err != nil {
		return err
	}

	return db.DB.Delete(a).Error
}

// LoadJobs loads all the jobs of an agent
func (a *Agent) LoadJobs() error {
	var jobs []*Job
	if err := db.DB.Where("agent_host = ?", a.Host).Find(&jobs).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error getting the jobs list: %v", err)
	}

	a.Jobs = jobs

	return nil
}

// LoadPlugins loads all the jobs of an agent
func (a *Agent) LoadPlugins() error {
	var plugins []*Plugin
	if err := db.DB.Where("agent_host = ?", a.Host).Find(&plugins).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error getting the plugins list: %v", err)
	}

	a.Plugins = plugins

	return nil
}
