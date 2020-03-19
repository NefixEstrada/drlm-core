// SPDX-License-Identifier: AGPL-3.0-only

package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/jinzhu/gorm"
)

// Job is an individual job of the scheduler
type Job struct {
	gorm.Model

	PluginID   uint      `gorm:"not null"`
	Plugin     *Plugin   `gorm:"-"`
	AgentHost  string    `gorm:"not null"`
	Status     JobStatus `gorm:"not null"`
	Time       time.Time
	Config     string `gorm:"not null"`
	BucketName string `gorm:"not null;unique"`
	Info       string

	Mux            sync.Mutex `gorm:"-"`
	ReconnAttempts int
}

// JobStatus is the status of a job
type JobStatus int

const (
	// JobStatusUnknown is when a job status is not known
	JobStatusUnknown JobStatus = iota
	// JobStatusScheduled is when a job has been scheduled
	JobStatusScheduled
	// JobStatusRunning is when a job is currently running
	JobStatusRunning
	// JobStatusFinished is when a job has already finished
	JobStatusFinished
	// JobStatusFailed is when a job has failed during the execution
	JobStatusFailed
	// JobStatusCancelled is when a job has been cancelled during the execution
	JobStatusCancelled
)

// JobList returns a list with all the jobs
func JobList(ctx *context.Context) ([]*Job, error) {
	jobs := []*Job{}

	if err := ctx.DB.Find(&jobs).Error; err != nil {
		return []*Job{}, fmt.Errorf("error getting the jobs list: %v", err)
	}

	return jobs, nil
}

// Add creates a new job in the DB
func (j *Job) Add(ctx *context.Context) error {
	if err := ctx.DB.Create(j).Error; err != nil {
		return fmt.Errorf("error adding the job to the DB: %v", err)
	}

	return nil
}

// Load loads the job from the DB
func (j *Job) Load(ctx *context.Context) error {
	var p Plugin
	j.Plugin = &p

	if err := ctx.DB.First(j).Related(&p, "PluginID").Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error loading the job from the DB: %v", err)
	}

	return nil
}

// Update updates the job in the DB
func (j *Job) Update(ctx *context.Context) error {
	if err := ctx.DB.Save(j).Error; err != nil {
		return fmt.Errorf("error updating the job: %v", err)
	}

	return nil
}
