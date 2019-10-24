package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/db"

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

	Version       string
	Arch          os.Arch
	OS            os.OS
	OSVersion     string
	Distro        string
	DistroVersion string
}

// AgentList returns a list with all the agents
func AgentList() ([]*Agent, error) {
	agents := []*Agent{}

	if err := db.DB.Select("id, created_at, updated_at, host, port, user, public_key_path, private_key_path, version, arch, os, os_version, distro, distro_version").Find(&agents).Error; err != nil {
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

// Load loads the agent from the DB using the host
func (a *Agent) Load() error {
	return db.DB.Where("host = ?", a.Host).First(a).Error
}

// Delete removes an agent from the DB
func (a *Agent) Delete() error {
	if err := a.Load(); err != nil {
		return err
	}

	return db.DB.Delete(a).Error
}
