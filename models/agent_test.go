// SPDX-License-Identifier: AGPL-3.0-only

package models_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

type TestAgentSuite struct {
	suite.Suite
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestAgentSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func (s *TestAgentSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func TestAgent(t *testing.T) {
	suite.Run(t, new(TestAgentSuite))
}

func (s *TestAgentSuite) TestList() {
	s.Run("should return a list of agents", func() {
		now := time.Now()

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("agents"."accepted" = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "host", "minio_key", "secret", "ssh_port", "ssh_user", "version", "arch", "os", "os_version", "distro", "distro_version"}).
			AddRow(1, now, now, "192.168.0.10", "minioKey", "f0cKt3Rf$", 22, "drlm", "v0.0.1", os.ArchAmd64, os.Linux, "v5.0.2", "debian", "10.0").
			AddRow(2, now, now, "192.168.1.5", "minioKey", "f0cKt3Rf$", 22, "root", "v0.1.0", os.ArchAmd64, os.Linux, "v5.0.0", "ubuntu", "19.04"),
		)

		expectedAgents := []*models.Agent{
			&models.Agent{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Host:          "192.168.0.10",
				MinioKey:      "minioKey",
				Secret:        "f0cKt3Rf$",
				SSHPort:       22,
				SSHUser:       "drlm",
				Version:       "v0.0.1",
				Arch:          os.ArchAmd64,
				OS:            os.Linux,
				OSVersion:     "v5.0.2",
				Distro:        "debian",
				DistroVersion: "10.0",
			},
			&models.Agent{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Host:          "192.168.1.5",
				MinioKey:      "minioKey",
				Secret:        "f0cKt3Rf$",
				SSHPort:       22,
				SSHUser:       "root",
				Version:       "v0.1.0",
				Arch:          os.ArchAmd64,
				OS:            os.Linux,
				OSVersion:     "v5.0.0",
				Distro:        "ubuntu",
				DistroVersion: "19.04",
			},
		}

		agents, err := models.AgentList(s.ctx)

		s.Nil(err)
		s.Equal(expectedAgents, agents)
	})

	s.Run("should return an error if there's an error getting the list of agents", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("agents"."accepted" = $1))`)).WillReturnError(errors.New("testing error"))

		agents, err := models.AgentList(s.ctx)

		s.EqualError(err, "error getting the list of agents: testing error")
		s.Equal([]*models.Agent{}, agents)
	})
}

func (s *TestAgentSuite) TestAdd() {
	s.Run("should add the agent to the DB correctly", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		a := &models.Agent{
			Host:    "192.168.1.61",
			SSHPort: 1312,
			SSHUser: "drlm",
		}

		s.Nil(a.Add(s.ctx))
	})

	s.Run("should return an error if there's an error adding the agent to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host:    "192.168.1.61",
			SSHPort: 1312,
			SSHUser: "root",
		}

		s.EqualError(a.Add(s.ctx), "error adding the agent to the DB: testing error")
	})
}

func (s *TestAgentSuite) TestLoad() {
	s.Run("should load the agent correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host", "ssh_port", "ssh_user"}).
			AddRow(161, "192.168.1.61", 1312, "drlm"),
		)

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.Nil(a.Load(s.ctx))
		s.Equal(&models.Agent{
			Model:   gorm.Model{ID: 161},
			Host:    "192.168.1.61",
			SSHPort: 1312,
			SSHUser: "drlm",
		}, a)
	})

	s.Run("should return an error if the agent ins't found", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(gorm.ErrRecordNotFound)

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.True(gorm.IsRecordNotFoundError(a.Load(s.ctx)))
	})

	s.Run("should return an error if there's an error loading the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.EqualError(a.Load(s.ctx), "error loading the agent from the DB: testing error")
	})
}

func (s *TestAgentSuite) TestUpdate() {
	s.Run("should update the agent correctly", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "agents" SET "updated_at" = $1, "deleted_at" = $2, "host" = $3, "accepted" = $4, "minio_key" = $5, "secret" = $6, "ssh_port" = $7, "ssh_user" = $8, "ssh_host_keys" = $9, "version" = $10, "arch" = $11, "os" = $12, "os_version" = $13, "distro" = $14, "distro_version" = $15  WHERE "agents"."deleted_at" IS NULL AND "agents"."id" = $16`)).WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		a := &models.Agent{
			Model: gorm.Model{ID: 1},
			Host:  "server",
		}

		s.NoError(a.Update(s.ctx))
	})

	s.Run("should return an error if there's an error updating the agent", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "agents" SET "updated_at" = $1, "deleted_at" = $2, "host" = $3, "accepted" = $4, "minio_key" = $5, "secret" = $6, "ssh_port" = $7, "ssh_user" = $8, "ssh_host_keys" = $9, "version" = $10, "arch" = $11, "os" = $12, "os_version" = $13, "distro" = $14, "distro_version" = $15  WHERE "agents"."deleted_at" IS NULL AND "agents"."id" = $16`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Model: gorm.Model{ID: 1},
			Host:  "server",
		}

		s.EqualError(a.Update(s.ctx), "error updating the agent: testing error")
	})
}

func (s *TestAgentSuite) TestDelete() {
	s.Run("should delete the agent correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).
			AddRow(1, "192.168.1.61"),
		)
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "agents" SET "deleted_at"=$1 WHERE "agents"."deleted_at" IS NULL AND "agents"."id" = $2`)).WithArgs(&tests.DBAnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		a := models.Agent{
			Host: "192.168.1.61",
		}

		s.Nil(a.Delete(s.ctx))
	})

	s.Run("should return an error if there's an error deleting the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		a := models.Agent{
			Host: "192.168.1.61",
		}

		s.EqualError(a.Delete(s.ctx), "error loading the agent from the DB: testing error")
	})
}

func (s *TestAgentSuite) TestLoadJobs() {
	s.Run("should return the list of jobs correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs" WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "plugin_id", "agent_host", "status"}).
			AddRow(1, 4, "192.168.1.61", models.JobStatusFinished).
			AddRow(5, 5, "192.168.1.61", models.JobStatusCancelled).
			AddRow(23, 7, "192.168.1.61", models.JobStatusScheduled),
		)

		a := models.Agent{
			Host: "192.168.1.61",
		}

		err := a.LoadJobs(s.ctx)

		s.NoError(err)
		s.Equal([]*models.Job{
			{
				Model: gorm.Model{
					ID: 1,
				},
				PluginID:  4,
				AgentHost: "192.168.1.61",
				Status:    models.JobStatusFinished,
			},
			{
				Model: gorm.Model{
					ID: 5,
				},
				PluginID:  5,
				AgentHost: "192.168.1.61",
				Status:    models.JobStatusCancelled,
			},
			{
				Model: gorm.Model{
					ID: 23,
				},
				PluginID:  7,
				AgentHost: "192.168.1.61",
				Status:    models.JobStatusScheduled,
			},
		}, a.Jobs)
	})

	s.Run("should return an empty list if there are no jobs of the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs" WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnRows(sqlmock.NewRows([]string{"id", "plugin_id", "agent_host", "status"}))

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		err := a.LoadJobs(s.ctx)

		s.Nil(err)
		s.Equal([]*models.Job{}, a.Jobs)
	})

	s.Run("should return an error if there's an error listing the jobs", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "jobs"  WHERE "jobs"."deleted_at" IS NULL AND ((agent_host = $1))`)).WithArgs("192.168.1.61").WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		err := a.LoadJobs(s.ctx)

		s.EqualError(err, "error getting the jobs list: testing error")
		s.Equal([]*models.Job(nil), a.Jobs)
	})
}

func (s *TestAgentSuite) TestLoadPlugins() {
	s.Run("should load the plugins correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins"  WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "repo", "name", "version"}).
			AddRow(1, "default", "tar", "v1.0.0").
			AddRow(2, "default", "copy", "v1.0.0").
			AddRow(3, "nefix", "borg", "v0.4.2"),
		)

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.NoError(a.LoadPlugins(s.ctx))
		s.Equal(&models.Agent{
			Host: "192.168.1.61",
			Plugins: []*models.Plugin{{
				Model:   gorm.Model{ID: 1},
				Repo:    "default",
				Name:    "tar",
				Version: "v1.0.0",
			}, {
				Model:   gorm.Model{ID: 2},
				Repo:    "default",
				Name:    "copy",
				Version: "v1.0.0",
			}, {
				Model:   gorm.Model{ID: 3},
				Repo:    "nefix",
				Name:    "borg",
				Version: "v0.4.2",
			}},
		}, a)
	})

	s.Run("should return an error if there's an error listing the plugins", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "plugins"  WHERE "plugins"."deleted_at" IS NULL AND ((agent_host = $1))`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.EqualError(a.LoadPlugins(s.ctx), "error getting the plugins list: testing error")
	})
}
