package models_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/os"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

type TestAgentSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
}

func (s *TestAgentSuite) SetupTest() {
	s.mock = tests.GenerateDB(s.T())
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

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, port, user, public_key_path, private_key_path, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "host", "port", "user", "public_key_path", "private_key_path", "version", "arch", "os", "os_version", "distro", "distro_version"}).
			AddRow(1, now, now, "192.168.0.10", 22, "drlm", "/etc/drlm/keys/public.key", "/etc/drlm/keys/private.key", "v0.0.1", os.ArchAmd64, os.Linux, "v5.0.2", "debian", "10.0").
			AddRow(2, now, now, "192.168.1.5", 22, "drlm", "/etc/drlm/keys/public.key", "/etc/drlm/keys/private.key", "v0.1.0", os.ArchAmd64, os.Linux, "v5.0.0", "ubuntu", "19.04"),
		)

		expectedAgents := []*models.Agent{
			&models.Agent{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Host:           "192.168.0.10",
				Port:           22,
				User:           "drlm",
				PublicKeyPath:  "/etc/drlm/keys/public.key",
				PrivateKeyPath: "/etc/drlm/keys/private.key",
				Version:        "v0.0.1",
				Arch:           os.ArchAmd64,
				OS:             os.Linux,
				OSVersion:      "v5.0.2",
				Distro:         "debian",
				DistroVersion:  "10.0",
			},
			&models.Agent{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: now,
					UpdatedAt: now,
				},
				Host:           "192.168.1.5",
				Port:           22,
				User:           "drlm",
				PublicKeyPath:  "/etc/drlm/keys/public.key",
				PrivateKeyPath: "/etc/drlm/keys/private.key",
				Version:        "v0.1.0",
				Arch:           os.ArchAmd64,
				OS:             os.Linux,
				OSVersion:      "v5.0.0",
				Distro:         "ubuntu",
				DistroVersion:  "19.04",
			},
		}

		agents, err := models.AgentList()

		s.Nil(err)
		s.Equal(expectedAgents, agents)
	})

	s.Run("should return an error if there's an error getting the list of agents", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, port, user, public_key_path, private_key_path, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		agents, err := models.AgentList()

		s.EqualError(err, "error getting the list of agents: testing error")
		s.Equal([]*models.Agent{}, agents)
	})
}

func (s *TestAgentSuite) TestAdd() {
	s.Run("should add the agent to the DB correctly", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","port","user","public_key_path","private_key_path","host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "agents"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		a := &models.Agent{
			Host: "192.168.1.61",
			Port: 1312,
			User: "drlm",
		}

		s.Nil(a.Add())
	})

	s.Run("should return an error if there's an error adding the agent to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","port","user","public_key_path","private_key_path","host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "agents"."id"`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host: "192.168.1.61",
			Port: 1312,
			User: "drlm",
		}

		s.EqualError(a.Add(), "error adding the agent to the DB: testing error")
	})
}

func (s *TestAgentSuite) TestLoad() {

	s.Run("should load the agent correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host", "port", "user"}).
			AddRow(161, "192.168.1.61", 1312, "drlm"),
		)

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.Nil(a.Load())
		s.Equal(&models.Agent{
			Model: gorm.Model{ID: 161},
			Host:  "192.168.1.61",
			Port:  1312,
			User:  "drlm",
		}, a)
	})

	s.Run("should return an error if the agent ins't found", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(gorm.ErrRecordNotFound)

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.True(gorm.IsRecordNotFoundError(a.Load()))
	})

	s.Run("should return an error if there's an error loading the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		a := &models.Agent{
			Host: "192.168.1.61",
		}

		s.EqualError(a.Load(), "error loading the agent from the DB: testing error")
	})
}

func (s *TestAgentSuite) TestDelete() {
	s.Run("should delete the agent correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host"}).
			AddRow(1, "192.168.1.61"),
		)
		s.mock.ExpectBegin()
		s.mock.ExpectExec(regexp.QuoteMeta(`UPDATE "agents" SET "deleted_at"=$1  WHERE "agents"."deleted_at" IS NULL AND "agents"."id" = $2`)).WithArgs(&tests.DBAnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		s.mock.ExpectCommit()

		a := models.Agent{
			Host: "192.168.1.61",
		}

		s.Nil(a.Delete())
	})

	s.Run("should return an error if there's an error deleting the agent", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "agents" WHERE "agents"."deleted_at" IS NULL AND ((host = $1)) ORDER BY "agents"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		a := models.Agent{
			Host: "192.168.1.61",
		}

		s.EqualError(a.Delete(), "error loading the agent from the DB: testing error")
	})
}
