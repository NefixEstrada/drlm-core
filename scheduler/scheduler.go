package scheduler

import (
	"fmt"
	"sync"

	"github.com/brainupdaters/drlm-core/models"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// Cron is the responsible for notifying the agents the scheduled jobs
var Cron *cron.Cron

// wg is a WaitGroup to be able to cancel the running jobs
var wg *sync.WaitGroup

// Init creates the Cron scheduler
func Init() {
	Cron = cron.New()

	agents, err := models.AgentList()
	if err != nil {
		log.Fatalf("error initializing the scheduler: error getting the agents list: %v", err)
	}

	for _, agent := range agents {
		Cron.AddFunc("@daily", func() {
			if err := AddJob("sync", agent.Host); err != nil {
				log.Fatalf("error scheduling the sync job for %s: %v", agent.Host, err)
			}
		})
	}

	Cron.Start()
}

// AddJob adds a new job to the scheduler
func AddJob(job string, host string) error {
	j := &models.Job{
		Name:      job,
		Status:    models.JobStatusScheduled,
		AgentHost: host,
	}

	if err := j.Add(); err != nil {
		return fmt.Errorf("error adding the job: %v", err)
	}

	return nil
}
