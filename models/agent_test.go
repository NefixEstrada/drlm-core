package models_test

import (
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

	})
}
