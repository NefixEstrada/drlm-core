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
	"github.com/stretchr/testify/suite"
)

type TestPluginSuite struct {
	suite.Suite
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestPluginSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func (s *TestPluginSuite) AfterTest() {
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestPlugin(t *testing.T) {
	suite.Run(t, &TestPluginSuite{})
}

func (s *TestPluginSuite) TestString() {
	p := &models.Plugin{
		Repo: "default",
		Name: "tar",
	}

	s.Equal("default/tar", p.String())
}

func (s *TestPluginSuite) TestAdd() {
	s.Run("should add the plugin correctly to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "plugins" ("created_at","updated_at","deleted_at","repo","name","version","agent_host") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "plugins"."id"`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(1),
		)
		s.mock.ExpectCommit()

		p := models.Plugin{
			Repo:      "default",
			Name:      "tar",
			Version:   "v1.0.0",
			AgentHost: "laptop",
		}

		s.NoError(p.Add(s.ctx))
	})

	s.Run("should return an error if there's an error addintg the plugin", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "plugins" ("created_at","updated_at","deleted_at","repo","name","version","agent_host") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "plugins"."id"`)).WillReturnError(errors.New("testing error"))

		p := models.Plugin{
			Repo:      "default",
			Name:      "tar",
			Version:   "v1.0.0",
			AgentHost: "laptop",
		}

		s.EqualError(p.Add(s.ctx), "error adding the plugin to the DB: testing error")
	})
}
