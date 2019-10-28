package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/db"

	"github.com/jinzhu/gorm"
)

// Job is an individual job of the scheduler
type Job struct {
	gorm.Model
	Name      string    `gorm:"not null"`
	Agent     *Agent    `gorm:"foreignkey:Host;association_foreignkey:AgentHost"`
	AgentHost string    `gorm:"not null"`
	Status    JobStatus `gorm:"not null"`
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
func JobList() ([]*Job, error) {
	jobs := []*Job{}

	if err := db.DB.Find(&jobs).Error; err != nil {
		return []*Job{}, fmt.Errorf("error getting the jobs list: %v", err)
	}

	for _, j := range jobs {
		j.Agent = &Agent{
			Host: j.AgentHost,
		}

		if err := j.Agent.Load(); err != nil {
			return []*Job{}, fmt.Errorf("error getting the agent for the job #%d: %v", j.ID, err)
		}
	}

	return jobs, nil
}

// AgentJobList returns a list with all the jobs of an specific agent
func AgentJobList(agentHost string) ([]*Job, error) {
	a := &Agent{
		Host: agentHost,
	}

	if err := a.Load(); err != nil {
		return []*Job{}, err
	}

	var jobs []*Job
	if err := db.DB.Where("agent_host = ?", agentHost).Find(&jobs).Error; err != nil {
		return []*Job{}, fmt.Errorf("error getting the jobs list: %v", err)
	}

	for _, j := range jobs {
		j.Agent = a
	}

	return jobs, nil
}

// Add creates a new job in the DB
func (j *Job) Add() error {
	if err := db.DB.Create(j).Error; err != nil {
		return fmt.Errorf("error adding the job to the DB: %v", err)
	}

	return nil
}

// Load loads the job from teh DB
func (j *Job) Load() error {
	var a Agent
	j.Agent = &a

	if err := db.DB.First(j).Related(&a, "Agent").Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error loading the job from the DB: %v", err)
	}

	return nil
}
