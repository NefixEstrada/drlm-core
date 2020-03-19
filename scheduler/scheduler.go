// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"errors"
	"fmt"
	"time"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errAgentUnavailable = errors.New("agent unavailable")

// Init starts the scheduler
func Init(ctx *context.Context) {
	// TODO: Think if we need all the jobs or just the ones that need to be executed
	j, err := models.JobList(ctx)
	if err != nil {
		log.Fatalf("error initializating the scheduler: %v", err)
	}

	jobs.Add(j...)

	queue := make(chan *models.Job)
	go scheduler(ctx, queue)
	go worker(ctx, queue)

	// Since both the scheduler and the worker use the waitgroup it has to be incremented by 1
	ctx.WG.Add(1)
}

// scheduler is the main function of the scheduler. It has timers that execute the respective functions when needed
func scheduler(ctx *context.Context, queue chan *models.Job) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			for _, j := range jobs.List() {
				j.Mux.Lock()
				if time.Now().After(j.Time) && j.Status == models.JobStatusScheduled {
					queue <- j
				} else {
					j.Mux.Unlock()
				}
			}

		// case <-daily.C:
		// TODO: Implement Cron for "Repetitive Jobs"
		// TODO: Make sync a plugin and don't
		// agents, err := models.AgentList(ctx)
		// if err != nil {
		// 	log.Fatalf("scheduler: error listing the agents: %v", err)
		// }

		// for _, a := range agents {
		// 	if err := agent.Sync(ctx, a); err != nil {
		// 		log.Errorf("error syncing the agent '%s': %v", a.Host, err)
		// 	}
		// }

		case <-ctx.Done():
			// TODO: This gets executed twice!
			ctx.WG.Done()
			return
		}
	}
}

func worker(ctx *context.Context, queue chan *models.Job) {
	for {
		select {
		case j := <-queue:
			stream, ok := AgentConnections.Get(j.AgentHost)
			if !ok {
				handleJobError(j, errAgentUnavailable)

			} else {
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
						err = errAgentUnavailable
					}

					handleJobError(j, err)

				} else {
					j.Status = models.JobStatusRunning
				}
			}

			if err := j.Update(ctx); err != nil {
				log.Error(err.Error())
			}

			j.Mux.Unlock()

		case <-ctx.Done():
			// TODO: Wait for all the jobs to stop!
			ctx.WG.Done()
			return
		}
	}
}

func handleJobError(j *models.Job, err error) {
	if err == errAgentUnavailable && j.ReconnAttempts < 10 {
		j.ReconnAttempts++

	} else {
		j.Status = models.JobStatusFailed
		j.Info = err.Error()

		log.Errorf("error starting the job: %v", err)
	}
}
