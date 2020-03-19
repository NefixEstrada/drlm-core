// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type TestTokenSuite struct {
	test.Test
	ctx  *context.Context
	mock sqlmock.Sqlmock
}

func TestToken(t *testing.T) {
	suite.Run(t, new(TestTokenSuite))
}

func (s *TestTokenSuite) SetupTest() {
	s.ctx = tests.GenerateCtx()
	s.mock = tests.GenerateDB(s.T(), s.ctx)
}

func (s *TestTokenSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func (s *TestTokenSuite) TestNew() {
	tests.GenerateCfg(s.T(), s.ctx)

	tkn, expiresAt, err := auth.NewToken(s.ctx, "nefix")
	s.Nil(err)
	s.NotNil(tkn)
	s.True(expiresAt.After(time.Now()))
}

func (s *TestTokenSuite) TestValidate() {
	s.Run("should return true if the token is valid", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{

			Usr:         "nefix",
			FirstIssued: time.Now(),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(s.ctx.Cfg.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		s.True(tkn.Validate(s.ctx))
	})

	s.Run("should return false if there's an error parsing the token", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		tkn := auth.Token("invalid token!")
		s.False(tkn.Validate(s.ctx))
	})

	s.Run("should return false if the token is invalid", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: time.Now(),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-s.ctx.Cfg.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		s.False(tkn.Validate(s.ctx))
	})
}

func (s *TestTokenSuite) TestValidateAgent() {
	s.Run("should return true if the secret is valid", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("agents"."accepted" = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host", "secret"}).
			AddRow(1, "server", "supersecret").
			AddRow(2, "laptop", "secret"),
		)

		tkn := auth.Token("secret")

		host, ok := tkn.ValidateAgent(s.ctx)

		s.True(ok)
		s.Equal("laptop", host)
	})

	s.Run("should return false if there's an error getting the agent list", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("agents"."accepted" = $1))`)).WillReturnError(errors.New("testing error"))

		tkn := auth.Token("secret")

		host, ok := tkn.ValidateAgent(s.ctx)

		s.False(ok)
		s.Equal("", host)
	})

	s.Run("should return false if the secret isn't found in the agents list", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, created_at, updated_at, host, accepted, minio_key, secret, ssh_port, ssh_user, ssh_host_keys, version, arch, os, os_version, distro, distro_version FROM "agents" WHERE "agents"."deleted_at" IS NULL AND (("agents"."accepted" = $1))`)).WillReturnRows(sqlmock.NewRows([]string{"id", "host", "secret"}).
			AddRow(1, "server", "supersecret").
			AddRow(2, "laptop", "secret"),
		)

		tkn := auth.Token("h4ck3r")

		host, ok := tkn.ValidateAgent(s.ctx)

		s.False(ok)
		s.Equal("", host)
	})
}

func (s *TestTokenSuite) TestRenew() {
	s.Run("should renew the token correctly", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		originalExpirationTime := time.Now().Add(1 * time.Minute)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew(s.ctx)

		s.Nil(err)
		s.True(expiresAt.After(originalExpirationTime))
	})

	s.Run("should renew the token if it has expired but the user hasn't been modified since the token was issued", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(1, "nefix", time.Now().Add(-10*time.Minute), time.Now().Add(-10*time.Minute)),
		)

		originalExpirationTime := time.Now().Add(-s.ctx.Cfg.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew(s.ctx)

		s.Nil(err)
		s.True(expiresAt.After(originalExpirationTime))
	})

	s.Run("should return an error if there's an error parsing the token", func() {
		tkn := auth.Token("invalid token!")

		_, err := tkn.Renew(s.ctx)
		s.EqualError(err, "error renewing the token: the token is invalid or can't be renewed")
	})

	s.Run("should return an error if the token has expired and the login lifespan has been reached", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		originalExpirationTime := time.Now().Add(-s.ctx.Cfg.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime.Add(-250 * time.Hour),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		_, err = tkn.Renew(s.ctx)
		s.EqualError(err, "error renewing the token: login lifespan exceeded, login again")
	})

	s.Run("should return an error if the token has expired and there's an error loading the DB user", func() {
		tests.GenerateCfg(s.T(), s.ctx)

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		originalExpirationTime := time.Now().Add(-s.ctx.Cfg.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(s.ctx.Cfg.Security.TokensSecret))
		s.Nil(err)

		tkn := auth.Token(signedTkn)
		_, err = tkn.Renew(s.ctx)
		s.EqualError(err, "error renewing the token: error loading the user from the DB: testing error")
	})
}

func (s *TestTokenSuite) TestString() {
	s.Equal("imaginethisisatoken", auth.Token("imaginethisisatoken").String())
}
