// SPDX-License-Identifier: AGPL-3.0-only

package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/secret"

	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
)

// Agent (s) are the clients of DRLM Core that are installed in the servers
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
	Arch          os.Arch
	OS            os.OS
	OSVersion     string
	Distro        string
	DistroVersion string

	Jobs    []*Job    `gorm:"-"`
	Plugins []*Plugin `gorm:"-"`
}

func agentList(ctx *context.Context, accepted bool) ([]*Agent, error) {
	agents := []*Agent{}

	if err := ctx.DB.Select("id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version").
		Where("accepted = ?", accepted).Find(&agents).Error; err != nil {
		return []*Agent{}, fmt.Errorf("error getting the list of agents: %v", err)
	}

	return agents, nil
}

// AgentList returns a list with all the agents
func AgentList(ctx *context.Context) ([]*Agent, error) {
	return agentList(ctx, true)
}

// AgentRequestList returns a list with all the agents that have requested to join the DRLM Core
func AgentRequestList(ctx *context.Context) ([]*Agent, error) {
	return agentList(ctx, false)
}

// Add creates a new agent in the DB
func (a *Agent) Add(ctx *context.Context) error {
	if err := ctx.DB.Create(a).Error; err != nil {
		return fmt.Errorf("error adding the agent to the DB: %v", err)
	}

	return nil
}

// BeforeCreate is a hook that gets executed before creating an agent
func (a *Agent) BeforeCreate() error {
	if a.Secret == "" {
		var err error
		a.Secret, err = secret.New(xid.New().String())
		if err != nil {
			return fmt.Errorf("generate secret: %v", err)
		}
	}

	return nil
}

// Load loads the agent from the DB using the host
func (a *Agent) Load(ctx *context.Context) error {
	if err := ctx.DB.Where("host = ?", a.Host).First(a).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error loading the agent from the DB: %v", err)
	}

	return nil
}

// Update updates the agent in the DB
func (a *Agent) Update(ctx *context.Context) error {
	if err := ctx.DB.Save(a).Error; err != nil {
		return fmt.Errorf("error updating the agent: %v", err)
	}

	return nil
}

// Delete removes an agent from the DB
func (a *Agent) Delete(ctx *context.Context) error {
	if err := a.Load(ctx); err != nil {
		return err
	}

	return ctx.DB.Delete(a).Error
}

// LoadJobs loads all the jobs of an agent
func (a *Agent) LoadJobs(ctx *context.Context) error {
	var jobs []*Job
	if err := ctx.DB.Where("agent_host = ?", a.Host).Find(&jobs).Error; err != nil {
		return fmt.Errorf("error getting the jobs list: %v", err)
	}

	a.Jobs = jobs

	return nil
}

// LoadPlugins loads all the jobs of an agent
func (a *Agent) LoadPlugins(ctx *context.Context) error {
	var plugins []*Plugin
	if err := ctx.DB.Where("agent_host = ?", a.Host).Find(&plugins).Error; err != nil {
		return fmt.Errorf("error getting the plugins list: %v", err)
	}

	a.Plugins = plugins

	return nil
}

// MinioAccess returns the minio access key for the agent
func (a *Agent) MinioAccess() string {
	return fmt.Sprintf("drlm-agent-%d", a.ID)
}
