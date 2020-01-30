// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/minio"
	"github.com/brainupdaters/drlm-core/models"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	jobs []*models.Job
	// AgentConnections are all the active agent connections
	AgentConnections = map[string]drlm.DRLM_AgentConnectionServer{}
)

// Init starts the scheduler
func Init(ctx context.Context) {
	var err error
	jobs, err = models.JobList()
	if err != nil {
		log.Fatalf("error initializating the scheduler: %v", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	daily := time.NewTicker(24 * time.Hour)

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, j := range jobs {
					j.Mu.Lock()
					if time.Now().After(j.Time) && j.Status == models.JobStatusScheduled {

						if err := startJob(j); err != nil {
							if err == errAgentUnavailable && j.ReconnAttempts < 10 {
								j.ReconnAttempts++

							} else {
								j.Status = models.JobStatusFailed
								j.Info = err.Error()

								log.Errorf("error starting the job: %v", err)
							}

						} else {
							j.Status = models.JobStatusRunning
						}

						if err := j.Update(); err != nil {
							log.Error(err.Error())
						}
					}
					j.Mu.Unlock()
				}

			case <-daily.C:
				agents, err := models.AgentList()
				if err != nil {
					log.Fatalf("scheduler: error listing the agents: %v", err)
				}

				for _, a := range agents {
					if err := agent.Sync(a); err != nil {
						log.Errorf("error syncing the agent '%s': %v", a.Host, err)
					}
				}

			case <-ctx.Done():
				// TODO: This gets executed twice!
				// TODO: Wait for all the jobs to stop!
				ctx.Value("wg").(*sync.WaitGroup).Done()
				return
			}
		}
	}()
}

// ErrPluginNotFound gets returned if the plugin (job name) that has been requested is not found in the agent
var ErrPluginNotFound = errors.New("plugin for the job not found in the agent")

// AddJob adds a new job to the scheduler
func AddJob(host, job, config string, t time.Time) error {
	a := &models.Agent{Host: host}
	if err := a.Load(); err != nil {
		return err
	}

	if err := a.LoadPlugins(); err != nil {
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

	// TODO: Check Agent availability

	bName, err := minio.MakeBucketForUser("drlm-agent-" + strconv.Itoa(int(a.ID)))
	if err != nil {
		return fmt.Errorf("error adding the job: %v", err)
	}
	j.BucketName = bName

	if err := j.Add(); err != nil {
		return fmt.Errorf("error adding the job: %v", err)
	}

	jobs = append(jobs, j)

	return nil
}

var errAgentUnavailable = errors.New("agent unavailable")

func startJob(j *models.Job) error {
	stream, ok := AgentConnections[j.AgentHost]
	if !ok {
		return errAgentUnavailable
	}

	if err := stream.Send(&drlm.AgentConnectionFromCore{
		MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_JOB_NEW,
		JobNew: &drlm.AgentConnectionFromCore_JobNew{
			Id:     uint32(j.ID),
			Name:   fmt.Sprintf("drlm-plugin-%s-%s-%s", j.Plugin.Repo, j.Plugin.Name, j.Plugin.Version),
			Config: j.Config,
			Target: j.BucketName,
		},
	}); err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.Unavailable {
			return errAgentUnavailable
		}

		return fmt.Errorf("error starting the job: %v", err)
	}

	return nil
}
