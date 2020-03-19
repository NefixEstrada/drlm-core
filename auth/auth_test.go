// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestAuthSuite struct {
	test.Test
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func (s *TestAuthSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(TestAuthSuite))
}

func (s *TestAuthSuite) TestInit() {
	s.Run("should not create the admin user if it exists", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "auth_type"}).
			AddRow(1, "admin", types.Local),
		)

		s.ctx.Cfg.Security.TokensSecret = "Q~-Z{P|ahLL/7L~&UJnU~x+!t+7x-n^V~M&b$O.5[sZ+lS4zfG8Mz.:'#B,Iwr]5"

		auth.Init(s.ctx)
	})

	// s.Run("should create the admin user if it doesn't exist", func() {
	// 	tests.GenerateCfg(s.T(), s.ctx)

	// 	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = ?)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnError(gorm.ErrRecordNotFound)

	// 	s.ctx.Cfg.Security.TokensSecret = "Q~-Z{P|ahLL/7L~&UJnU~x+!t+7x-n^V~M&b$O.5[sZ+lS4zfG8Mz.:'#B,Iwr]5"

	// 	auth.Init(s.ctx)
	// })

	s.Run("should exit if there's no token secret", func() {
		tests.GenerateCfg(s.T(), s.ctx)
		s.Exits(func() { auth.Init(s.ctx) })
	})

	s.Run("should exit if the secret has less than 32 characters", func() {
		tests.GenerateCfg(s.T(), s.ctx)
		s.ctx.Cfg.Security.TokensSecret = "lorem ipsum dolor"

		s.Exits(func() { auth.Init(s.ctx) })
	})

	s.Run("should exit if there's an error querying for an admin user", func() {

	})
}
