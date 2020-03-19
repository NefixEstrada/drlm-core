// SPDX-License-Identifier: AGPL-3.0-only

package agent_test

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/agent"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
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
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestAgent(t *testing.T) {
	suite.Run(t, &TestAgentSuite{})
}

func (s *TestAgentSuite) TestAdd() {
	tests.GenerateCfg(s.T(), s.ctx)

	s.Run("should set the default values, create the minio user and save the agent in the DB", func() {
		ts := tests.GenerateMinio(s.ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		a := &models.Agent{Host: "192.168.1.61"}

		s.NoError(agent.Add(s.ctx, a))
	})

	s.Run("should return an error if there's an error creating the minio user", func() {
		ts := tests.GenerateMinio(s.ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		a := &models.Agent{Host: "192.168.1.61"}

		s.EqualError(agent.Add(s.ctx, a), "error creating the agent minio user: error creating the minio user: Failed to parse server response.")
	})

	s.Run("should return an error if there's an error adding the agent in the DB", func() {
		ts := tests.GenerateMinio(s.ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnError(errors.New("testing error!"))
		s.mock.ExpectCommit()

		a := &models.Agent{Host: "192.168.1.61"}

		s.EqualError(agent.Add(s.ctx, a), "error adding the agent to the DB: testing error!")
	})
}

func (s *TestAgentSuite) TestAddRequest() {
	s.Run("should add the agent add request correctly", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		a := &models.Agent{Host: "192.168.1.61"}

		s.NoError(agent.AddRequest(s.ctx, a))
	})

	s.Run("should return an error if there's an error adding the agent add request", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "agents" ("created_at","updated_at","deleted_at","host","accepted","minio_key","secret","ssh_port","ssh_user","ssh_host_keys","version","arch","os","os_version","distro","distro_version") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) RETURNING "agents"."id"`)).WillReturnError(errors.New("testing error!"))
		s.mock.ExpectCommit()

		a := &models.Agent{Host: "192.168.1.61"}

		s.EqualError(agent.AddRequest(s.ctx, a), "error adding the agent to the DB: testing error!")
	})
}
