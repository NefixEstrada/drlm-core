// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/models"
)

var (
	jobs = jobList{v: []*models.Job{}}
	// ErrPluginNotFound gets returned if the plugin (job name) that has been requested is not found in the agent
	ErrPluginNotFound = errors.New("plugin for the job not found in the agent")
)

type jobList struct {
	v   []*models.Job
	mux sync.Mutex
}

func (j *jobList) List() []*models.Job {
	j.mux.Lock()
	defer j.mux.Unlock()

	return j.v
}

func (j *jobList) Add(jobs ...*models.Job) {
	j.mux.Lock()
	defer j.mux.Unlock()

	j.v = append(j.v, jobs...)
}

// AddJob adds a new job to the scheduler
func AddJob(ctx *context.Context, host, job, config string, t time.Time) error {
	a := &models.Agent{Host: host}
	if err := a.Load(ctx); err != nil {
		return err
	}

	if err := a.LoadPlugins(ctx); err != nil {
		return err
	}

	j := &models.Job{
		Status:    models.JobStatusScheduled,
		AgentHost: a.Host,
		Config:    config,
		Time:      t,
	}

	for _, p := range a.Plugins {
		if p.String() == job {
			j.Plugin = p
			j.PluginID = p.ID
		}
	}
	if j.PluginID == 0 {
		return ErrPluginNotFound
	}

	// TODO: Check Agent availability if the task is scheduled for now (before time.Now())

	bName, err := minio.MakeBucketForUser(ctx, "drlm-agent-"+strconv.Itoa(int(a.ID)))
	if err != nil {
		return fmt.Errorf("error adding the job: %v", err)
	}
	j.BucketName = bName

	if err := j.Add(ctx); err != nil {
		return fmt.Errorf("error adding the job: %v", err)
	}

	jobs.Add(j)

	return nil
}
