// SPDX-License-Identifier: AGPL-3.0-only

package scheduler

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestSchedulerInternalSuite struct {
	test.Test
}

func TestSchedulerInternal(t *testing.T) {
	suite.Run(t, &TestSchedulerInternalSuite{})
}

func (s *TestSchedulerInternalSuite) TestScheduler() {
	s.Run("should add the job to the queue correctly", func() {
		ctx, cancel := context.WithCancel()
		jobs.v = []*models.Job{}

		j := &models.Job{
			Plugin: &models.Plugin{
				Repo: "default",
				Name: "tar",
			},
			AgentHost: "laptop",
			Status:    models.JobStatusScheduled,
			Time:      time.Now(),
		}
		jobs.Add(j)

		queue := make(chan *models.Job)

		go scheduler(ctx, queue)
		ctx.WG.Add(1)

		queueJob := <-queue

		cancel()

		s.Equal(j, queueJob)
	})

	s.Run("should skip the job if the job doesn't have to be executed", func() {
		ctx, cancel := context.WithCancel()
		jobs.v = []*models.Job{}

		j := &models.Job{}
		jobs.Add(j)

		queue := make(chan *models.Job)

		go scheduler(ctx, queue)
		ctx.WG.Add(1)

		var queueJob *models.Job

		go func() { queueJob = <-queue }()

		time.Sleep(5 * time.Second)

		cancel()

		j.Mux.Lock()
		j.Mux.Unlock()

		s.Nil(queueJob)
	})

	s.Run("should mark the WaitGroup as done when getting cancellation from the context", func() {
		ctx, cancel := context.WithCancel()

		queue := make(chan *models.Job)

		go scheduler(ctx, queue)
		ctx.WG.Add(1)

		cancel()

		ctx.WG.Wait()
	})
}

func (s *TestSchedulerInternalSuite) TestWorker() {
	s.Run("should start the job correctly", func() {
		ctx, cancel := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs" SET "updated_at" = $1, "deleted_at" = $2, "plugin_id" = $3, "agent_host" = $4, "status" = $5, "time" = $6, "config" = $7, "bucket_name" = $8, "info" = $9, "reconn_attempts" = $10  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $11`)).WillReturnResult(sqlmock.NewResult(83, 1))
		mock.ExpectCommit()

		j := &models.Job{
			Model: gorm.Model{ID: 85},
			Plugin: &models.Plugin{
				Repo:    "default",
				Name:    "tar",
				Version: "v1.0.0",
			},
			AgentHost:  "127.0.0.1",
			Status:     models.JobStatusScheduled,
			Config:     "{}",
			BucketName: "drlm-agent-1-name",
		}
		j.Mux.Lock()

		agentConnMock := &tests.AgentConnectionServerMock{}
		agentConnMock.On("Send", &drlm.AgentConnectionFromCore{
			MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_JOB_NEW,
			JobNew: &drlm.AgentConnectionFromCore_JobNew{
				Id:     uint32(j.ID),
				Name:   fmt.Sprintf("drlm-plugin-%s-%s-%s", j.Plugin.Repo, j.Plugin.Name, j.Plugin.Version),
				Config: j.Config,
				Target: j.BucketName,
			},
		}).Return(nil)

		agent.Connections.Add("127.0.0.1", agentConnMock)
		defer agent.Connections.Delete("127.0.0.1")

		queue := make(chan *models.Job)

		go worker(ctx, queue)
		ctx.WG.Add(1)

		queue <- j

		cancel()
		ctx.WG.Wait()

		s.Equal(models.JobStatusRunning, j.Status)
		agentConnMock.AssertExpectations(s.T())
	})

	s.Run("should increment the reconnection attempts by one if the agent connection isn't in the connection pool", func() {
		ctx, cancel := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		j := &models.Job{AgentHost: "127.0.0.1"}
		j.Mux.Lock()

		queue := make(chan *models.Job)

		go worker(ctx, queue)
		ctx.WG.Add(1)

		queue <- j

		cancel()
		ctx.WG.Wait()

		s.Equal(1, j.ReconnAttempts)
	})

	s.Run("should increment the reconnection attempts by one if the agent returns an unavailable error when starting the job", func() {
		ctx, cancel := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs" SET "updated_at" = $1, "deleted_at" = $2, "plugin_id" = $3, "agent_host" = $4, "status" = $5, "time" = $6, "config" = $7, "bucket_name" = $8, "info" = $9, "reconn_attempts" = $10  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $11`)).WillReturnResult(sqlmock.NewResult(83, 1))
		mock.ExpectCommit()

		j := &models.Job{
			Model: gorm.Model{ID: 85},
			Plugin: &models.Plugin{
				Repo:    "default",
				Name:    "tar",
				Version: "v1.0.0",
			},
			AgentHost:  "127.0.0.1",
			Status:     models.JobStatusScheduled,
			Config:     "{}",
			BucketName: "drlm-agent-1-name",
		}
		j.Mux.Lock()

		agentConnMock := &tests.AgentConnectionServerMock{}
		agentConnMock.On("Send", &drlm.AgentConnectionFromCore{
			MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_JOB_NEW,
			JobNew: &drlm.AgentConnectionFromCore_JobNew{
				Id:     uint32(j.ID),
				Name:   fmt.Sprintf("drlm-plugin-%s-%s-%s", j.Plugin.Repo, j.Plugin.Name, j.Plugin.Version),
				Config: j.Config,
				Target: j.BucketName,
			},
		}).Return(status.Error(codes.Unavailable, "you can't do anything, sorry :("))

		agent.Connections.Add("127.0.0.1", agentConnMock)
		defer agent.Connections.Delete("127.0.0.1")

		queue := make(chan *models.Job)

		go worker(ctx, queue)
		ctx.WG.Add(1)

		queue <- j

		cancel()
		ctx.WG.Wait()

		s.Equal(1, j.ReconnAttempts)
		agentConnMock.AssertExpectations(s.T())
	})

	s.Run("should fail the job if there's an error starting the job", func() {
		ctx, cancel := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs" SET "updated_at" = $1, "deleted_at" = $2, "plugin_id" = $3, "agent_host" = $4, "status" = $5, "time" = $6, "config" = $7, "bucket_name" = $8, "info" = $9, "reconn_attempts" = $10  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $11`)).WillReturnResult(sqlmock.NewResult(83, 1))
		mock.ExpectCommit()

		j := &models.Job{
			Model: gorm.Model{ID: 85},
			Plugin: &models.Plugin{
				Repo:    "default",
				Name:    "tar",
				Version: "v1.0.0",
			},
			AgentHost:  "127.0.0.1",
			Status:     models.JobStatusScheduled,
			Config:     "{}",
			BucketName: "drlm-agent-1-name",
		}
		j.Mux.Lock()

		agentConnMock := &tests.AgentConnectionServerMock{}
		agentConnMock.On("Send", &drlm.AgentConnectionFromCore{
			MessageType: drlm.AgentConnectionFromCore_MESSAGE_TYPE_JOB_NEW,
			JobNew: &drlm.AgentConnectionFromCore_JobNew{
				Id:     uint32(j.ID),
				Name:   fmt.Sprintf("drlm-plugin-%s-%s-%s", j.Plugin.Repo, j.Plugin.Name, j.Plugin.Version),
				Config: j.Config,
				Target: j.BucketName,
			},
		}).Return(status.Error(codes.Unknown, "testing error"))

		agent.Connections.Add("127.0.0.1", agentConnMock)
		defer agent.Connections.Delete("127.0.0.1")

		queue := make(chan *models.Job)

		go worker(ctx, queue)
		ctx.WG.Add(1)

		queue <- j

		cancel()
		ctx.WG.Wait()

		s.Equal(models.JobStatusFailed, j.Status)
		s.Equal(status.Error(codes.Unknown, "testing error").Error(), j.Info)
		agentConnMock.AssertExpectations(s.T())
	})

	s.Run("should log an error if there's an error updating the job in the DB", func() {
		ctx, _ := context.WithCancel()
		mock := tests.GenerateDB(s.T(), ctx)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnError(errors.New("testing error"))

		j := &models.Job{AgentHost: "127.0.0.1"}
		j.Mux.Lock()

		queue := make(chan *models.Job)

		go worker(ctx, queue)
		ctx.WG.Add(1)

		queue <- j
	})
}

func (s *TestSchedulerInternalSuite) TestHandleJobError() {
	s.Run("increment the reconnection attempts if it's an agent unavailable error and it has less than 10 attempts", func() {
		j := &models.Job{}

		handleJobError(j, errAgentUnavailable)

		s.Equal(1, j.ReconnAttempts)
	})

	s.Run("set the status to failed if the error is unknown or has 10 or more reconnection attempts", func() {
		j := &models.Job{}

		handleJobError(j, errors.New("testing error"))

		s.Equal(models.JobStatusFailed, j.Status)
		s.Equal("testing error", j.Info)
	})
}
