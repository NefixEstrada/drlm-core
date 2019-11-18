package models_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

type TestJobSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
}

func (s *TestJobSuite) SetupTest() {
	s.mock = tests.GenerateDB(s.T())
}

func (s *TestJobSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}
func TestJob(t *testing.T) {
	suite.Run(t, new(TestJobSuite))
}

func (s *TestJobSuite) TestList() {
	s.Run("should reutrn the list of jobs correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "agent_host"}).
			AddRow(1, "sync", models.JobStatusRunning, "laptop").
			AddRow(2, "sync", models.JobStatusFinished, "server"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("laptop").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "os", "distro"}).
			AddRow(161, "laptop", 22, os.Linux, "nixos"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("server").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "os", "distro"}).
			AddRow(1312, "server", 10022, os.Linux, "centos"),
		)

		expectedJobs := []*models.Job{
			&models.Job{
				Model: gorm.Model{
					ID: 1,
				},
				Name:      "sync",
				Status:    models.JobStatusRunning,
				AgentHost: "laptop",
				Agent: &models.Agent{
					Model: gorm.Model{
						ID: 161,
					},
					Host:   "laptop",
					Port:   22,
					OS:     os.Linux,
					Distro: "nixos",
				},
			},
			&models.Job{
				Model: gorm.Model{
					ID: 2,
				},
				Name:      "sync",
				Status:    models.JobStatusFinished,
				AgentHost: "server",
				Agent: &models.Agent{
					Model: gorm.Model{
						ID: 1312,
					},
					Host:   "server",
					Port:   10022,
					OS:     os.Linux,
					Distro: "centos",
				},
			},
		}

		jobs, err := models.JobList()

		s.Nil(err)
		s.Equal(expectedJobs, jobs)
	})

	s.Run("should return an error if there's an error listing the jobs", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		jobs, err := models.JobList()

		s.EqualError(err, "error getting the jobs list: testing error")
		s.Equal([]*models.Job{}, jobs)
	})

	s.Run("should return an error if there's an error getting a job agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "agent_host"}).
			AddRow(1, "sync", models.JobStatusRunning, "192.168.1.61"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnError(errors.New("testing error"))

		jobs, err := models.JobList()

		s.EqualError(err, "error getting the agent for the job #1: error loading the agent from the DB: testing error")
		s.Equal([]*models.Job{}, jobs)
	})
}

func (s *TestJobSuite) TestAgentJobList() {
	s.Run("should return the list of jobs correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents"  WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "user"}).
			AddRow(161, "192.168.1.61", 1312, "drlm"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "agent_host", "status"}).
			AddRow(1, "sync", "192.168.1.61", models.JobStatusFinished).
			AddRow(5, "rear_backup", "192.168.1.61", models.JobStatusCancelled).
			AddRow(23, "borg_backup", "192.168.1.61", models.JobStatusScheduled),
		)

		jobs, err := models.AgentJobList("192.168.1.61")

		s.Nil(err)
		s.Equal([]*models.Job{
			{
				Model: gorm.Model{
					ID: 1,
				},
				Name:      "sync",
				AgentHost: "192.168.1.61",
				Agent: &models.Agent{
					Model: gorm.Model{
						ID: 161,
					},
					Host: "192.168.1.61",
					Port: 1312,
					User: "drlm",
				},
				Status: models.JobStatusFinished,
			},
			{
				Model: gorm.Model{
					ID: 5,
				},
				Name:      "rear_backup",
				AgentHost: "192.168.1.61",
				Agent: &models.Agent{
					Model: gorm.Model{
						ID: 161,
					},
					Host: "192.168.1.61",
					Port: 1312,
					User: "drlm",
				},
				Status: models.JobStatusCancelled,
			},
			{
				Model: gorm.Model{
					ID: 23,
				},
				Name:      "borg_backup",
				AgentHost: "192.168.1.61",
				Agent: &models.Agent{
					Model: gorm.Model{
						ID: 161,
					},
					Host: "192.168.1.61",
					Port: 1312,
					User: "drlm",
				},
				Status: models.JobStatusScheduled,
			},
		}, jobs)
	})

	s.Run("should return an empty list if there are no jobs of the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents"  WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "user"}).
			AddRow(161, "192.168.1.61", 1312, "drlm"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "agent_host", "status"}))

		jobs, err := models.AgentJobList("192.168.1.61")

		s.Nil(err)
		s.Equal([]*models.Job{}, jobs)
	})

	s.Run("should return an error if there's an error loading the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents"  WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnError(errors.New("testing error"))

		jobs, err := models.AgentJobList("192.168.1.61")

		s.EqualError(err, "error loading the agent from the DB: testing error")
		s.Equal([]*models.Job{}, jobs)
	})

	s.Run("should return an error if there's an error listing the jobs", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents"  WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "user"}).
			AddRow(161, "192.168.1.61", 1312, "drlm"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnError(errors.New("testing error"))

		jobs, err := models.AgentJobList("192.168.1.61")

		s.EqualError(err, "error getting the jobs list: testing error")
		s.Equal([]*models.Job{}, jobs)
	})
}

func (s *TestJobSuite) TestAdd() {
	s.Run("should add the job to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","name","agent_host","status","bucket_name") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "jobs"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "sync", "192.168.1.61", models.JobStatusScheduled, tests.DBAnyBucketName{}).WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(1),
		)
		s.mock.ExpectCommit()

		j := models.Job{
			Name:       "sync",
			AgentHost:  "192.168.1.61",
			Status:     models.JobStatusScheduled,
			BucketName: "drlm-bn74rasu9jr587gc4fhg",
		}

		s.Nil(j.Add())
	})

	s.Run("should return an error if there's an error adding the job to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "jobs" ("created_at","updated_at","deleted_at","name","agent_host","status","bucket_name") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "jobs"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "sync", "192.168.1.61", models.JobStatusScheduled, tests.DBAnyBucketName{}).WillReturnError(errors.New("testing error"))
		s.mock.ExpectCommit()

		j := models.Job{
			Name:       "sync",
			AgentHost:  "192.168.1.61",
			Status:     models.JobStatusScheduled,
			BucketName: "drlm-bn74rasu9jr587gc4fhg",
		}

		s.EqualError(j.Add(), "error adding the job to the DB: testing error")
	})
}

func (s *TestJobSuite) TestLoad() {
	s.Run("should load the job correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "agent_host"}).
			AddRow(1, "sync", models.JobStatusRunning, "192.168.1.61"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("host" = $1)) ORDER BY "agents"."id" ASC`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "os", "distro"}).
			AddRow(161, "192.168.1.61", 22, os.Linux, "nixos"),
		)

		expectedJob := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			Name:      "sync",
			Status:    models.JobStatusRunning,
			AgentHost: "192.168.1.61",
			Agent: &models.Agent{
				Model: gorm.Model{
					ID: 161,
				},
				Host:   "192.168.1.61",
				Port:   22,
				OS:     os.Linux,
				Distro: "nixos",
			},
		}

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.Nil(j.Load())
		s.Equal(expectedJob, j)
	})

	s.Run("should return an error if job isn't found", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WithArgs(1).WillReturnError(gorm.ErrRecordNotFound)

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.True(gorm.IsRecordNotFoundError(j.Load()))
		s.Equal(&models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			Agent: &models.Agent{},
		}, j)
	})

	s.Run("should return an error if there's an error getting the job", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WithArgs(1).WillReturnError(errors.New("testing error"))

		expectedJob := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
			Agent: &models.Agent{},
		}

		j := &models.Job{
			Model: gorm.Model{
				ID: 1,
			},
		}

		s.EqualError(j.Load(), "error loading the job from the DB: testing error")
		s.Equal(expectedJob, j)
	})
}
