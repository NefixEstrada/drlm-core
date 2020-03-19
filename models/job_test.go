// SPDX-License-Identifier: AGPL-3.0-only

package models_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

type TestJobSuite struct {
	suite.Suite
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestJobSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func (s *TestJobSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}
func TestJob(t *testing.T) {
	suite.Run(t, new(TestJobSuite))
}

func (s *TestJobSuite) TestList() {
	s.Run("should reutrn the list of jobs correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"id", "plugin_id", "status", "agent_host"}).
			AddRow(1, 4, models.JobStatusRunning, "laptop").
			AddRow(2, 4, models.JobStatusFinished, "server"),
		)

		expectedJobs := []*models.Job{
			&models.Job{
				Model: gorm.Model{
					ID: 1,
				},
				PluginID:  4,
				Status:    models.JobStatusRunning,
				AgentHost: "laptop",
			},
			&models.Job{
				Model: gorm.Model{
					ID: 2,
				},
				PluginID:  4,
				Status:    models.JobStatusFinished,
				AgentHost: "server",
			},
		}

		jobs, err := models.JobList(s.ctx)

		s.Nil(err)
		s.Equal(expectedJobs, jobs)
	})

	s.Run("should return an error if there's an error listing the jobs", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		jobs, err := models.JobList(s.ctx)

		s.EqualError(err, "error getting the jobs list: testing error")
		s.Equal([]*models.Job{}, jobs)
	})
}
func (s *TestJobSuite) TestAdd() {
	s.Run("should add the job to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(1),
		)
		s.mock.ExpectCommit()

		j := models.Job{
			PluginID:   4,
			AgentHost:  "192.168.1.61",
			Status:     models.JobStatusScheduled,
			BucketName: "drlm-bn74rasu9jr587gc4fhg",
		}

		s.Nil(j.Add(s.ctx))
	})

	s.Run("should return an error if there's an error adding the job to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","plugin_id","agent_host","status","time","config","bucket_name","info","reconn_attempts") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "jobs"."id"`)).WillReturnError(errors.New("testing error"))
		s.mock.ExpectCommit()

		j := models.Job{
			PluginID:   4,
			AgentHost:  "192.168.1.61",
			Status:     models.JobStatusScheduled,
			BucketName: "drlm-bn74rasu9jr587gc4fhg",
		}

		s.EqualError(j.Add(s.ctx), "error adding the job to the DB: testing error")
	})
}

func (s *TestJobSuite) TestLoad() {
	s.Run("should load the job correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "plugin_id", "status", "agent_host"}).
			AddRow(1, 4, models.JobStatusRunning, "192.168.1.61"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins" WHERE "plugins"."deleted_at" IS NULL AND (("id" = $1)) ORDER BY "plugins"."id" ASC`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name", "version", "agent_host"}).
			AddRow(4, "default", "tar", "v1.0.0", "192.168.1.61"),
		)

		expectedJob := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			PluginID: 4,
			Plugin: &models.Plugin{
				Model: gorm.Model{
					ID: 4,
				},
				Repo:      "default",
				Name:      "tar",
				Version:   "v1.0.0",
				AgentHost: "192.168.1.61",
			},
			Status:    models.JobStatusRunning,
			AgentHost: "192.168.1.61",
		}

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.Nil(j.Load(s.ctx))
		s.Equal(expectedJob, j)
	})

	s.Run("should return an error if job isn't found", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WillReturnError(gorm.ErrRecordNotFound)

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.True(gorm.IsRecordNotFoundError(j.Load(s.ctx)))
		s.Equal(&models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			Plugin: &models.Plugin{},
		}, j)
	})

	s.Run("should return an error if there's an error getting the job", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		expectedJob := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			Plugin: &models.Plugin{},
		}

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.EqualError(j.Load(s.ctx), "error loading the job from the DB: testing error")
		s.Equal(expectedJob, j)
	})
}

func (s *TestJobSuite) TestUpdate() {
	s.Run("should update the job correctly", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs" SET "updated_at" = $1, "deleted_at" = $2, "plugin_id" = $3, "agent_host" = $4, "status" = $5, "time" = $6, "config" = $7, "bucket_name" = $8, "info" = $9, "reconn_attempts" = $10  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $11`)).WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		j := &models.Job{
			Model:     gorm.Model{ID: 1},
			AgentHost: "server",
		}

		s.NoError(j.Update(s.ctx))
	})

	s.Run("should return an error if there's an error updating the job", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "jobs" SET "updated_at" = $1, "deleted_at" = $2, "plugin_id" = $3, "agent_host" = $4, "status" = $5, "time" = $6, "config" = $7, "bucket_name" = $8, "info" = $9, "reconn_attempts" = $10  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $11`)).WillReturnError(errors.New(`testing error`))

		j := &models.Job{
			Model:     gorm.Model{ID: 1},
			AgentHost: "server",
		}

		s.EqualError(j.Update(s.ctx), "error updating the job: testing error")
	})
}
