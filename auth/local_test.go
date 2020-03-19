// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type TestLocalSuite struct {
	suite.Suite
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func TestLocal(t *testing.T) {
	suite.Run(t, new(TestLocalSuite))
}

func (s *TestLocalSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func (s *TestLocalSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func (s *TestLocalSuite) TestLogin() {
	s.Run("should return a token if the authentication succeeds", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW", types.Local),
		)

		tkn, expiresAt, err := auth.LoginLocal(s.ctx, "nefix", "f0cKt3Rf$")

		s.Nil(err)
		s.NotNil(tkn)
		s.True(expiresAt.After(time.Now()))
	})

	s.Run("should return an error if the user isn't found", func() {

	})

	s.Run("should return an error if there's an error loading the user", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		tkn, _, err := auth.LoginLocal(s.ctx, "nefix", "f0cKt3Rf$")

		s.EqualError(err, "error loading the user from the DB: testing error")
		s.Equal("", tkn.String())
	})

	s.Run("should return an error if the login type isn't local", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW", types.Unknown),
		)

		tkn, _, err := auth.LoginLocal(s.ctx, "nefix", "f0cKt3Rf$")

		s.EqualError(err, "invalid authentication method: user authentication type is unknown")
		s.Equal("", tkn.String())
	})

	s.Run("should return an error if the passwords don't match", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW", types.Local),
		)

		tkn, _, err := auth.LoginLocal(s.ctx, "nefix", "asdfzxcv")

		s.Equal(bcrypt.ErrMismatchedHashAndPassword, err)
		s.Equal("", tkn.String())
	})

	s.Run("should return an error if there's an error comparing the passwords", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "f0cKt3Rf$", types.Local),
		)

		tkn, _, err := auth.LoginLocal(s.ctx, "nefix", "asdfzxcv")

		s.EqualError(err, "password error: crypto/bcrypt: hashedSecret too short to be a bcrypted password")
		s.Equal("", tkn.String())
	})
}
