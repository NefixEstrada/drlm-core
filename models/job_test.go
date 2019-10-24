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

		s.EqualError(err, "error getting the agent for the job #1: testing error")
		s.Equal([]*models.Job{}, jobs)
	})
}

func (s *TestJobSuite) TestLoad() {
	s.Run("should load the job correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND "jobs"."id" = $1 ORDER BY "jobs"."id" ASC LIMIT 1`)).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "agent_host"}).
			AddRow(1, "sync", models.JobStatusRunning, "192.168.1.61"),
		)
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "os", "distro"}).
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
				Host:   "laptop",
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
