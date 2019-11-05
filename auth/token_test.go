package auth_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type TestTokenSuite struct {
	test.Test
	mock sqlmock.Sqlmock
}

func TestToken(t *testing.T) {
	suite.Run(t, new(TestTokenSuite))
}

func (s *TestTokenSuite) SetupTest() {
	s.mock = tests.GenerateDB(s.T())
}

func (s *TestTokenSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func (s *TestTokenSuite) TestNew() {
	tests.GenerateCfg(s.T())

	tkn, expiresAt, err := auth.NewToken("nefix")
	s.Nil(err)
	s.NotNil(tkn)
	s.True(expiresAt.After(time.Now()))
}

func (s *TestTokenSuite) TestValidate() {
	s.Run("should return true if the token is valid", func() {
		tests.GenerateCfg(s.T())

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: time.Now(),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.Config.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		s.True(tkn.Validate())
	})

	s.Run("should return false if there's an error parsing the token", func() {
		tests.GenerateCfg(s.T())

		tkn := auth.Token("invalid token!")
		s.False(tkn.Validate())
	})

	s.Run("should return false if the token is invalid", func() {
		tests.GenerateCfg(s.T())

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: time.Now(),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-cfg.Config.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		s.False(tkn.Validate())
	})
}

func (s *TestTokenSuite) TestRenew() {
	s.Run("should renew the token correctly", func() {
		tests.GenerateCfg(s.T())

		originalExpirationTime := time.Now().Add(1 * time.Minute)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew()

		s.Nil(err)
		s.True(expiresAt.After(originalExpirationTime))
	})

	s.Run("should renew the token if it has expired but the user hasn't been modified since the token was issued", func() {
		tests.GenerateCfg(s.T())

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(1, "nefix", time.Now().Add(-10*time.Minute), time.Now().Add(-10*time.Minute)),
		)

		originalExpirationTime := time.Now().Add(-cfg.Config.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew()

		s.Nil(err)
		s.True(expiresAt.After(originalExpirationTime))
	})

	s.Run("should return an error if there's an error parsing the token", func() {
		tkn := auth.Token("invalid token!")

		_, err := tkn.Renew()
		s.EqualError(err, "error renewing the token: the token is invalid or can't be renewed")
	})

	s.Run("should return an error if the token has expired and the login lifespan has been reached", func() {
		tests.GenerateCfg(s.T())

		originalExpirationTime := time.Now().Add(-cfg.Config.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime.Add(-250 * time.Hour),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Require().Nil(err)

		tkn := auth.Token(signedTkn)
		_, err = tkn.Renew()
		s.EqualError(err, "error renewing the token: login lifespan exceeded, login again")
	})

	s.Run("should return an error if the token has expired and there's an error loading the DB user", func() {
		tests.GenerateCfg(s.T())

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnError(errors.New("testing error"))

		originalExpirationTime := time.Now().Add(-cfg.Config.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		s.Nil(err)

		tkn := auth.Token(signedTkn)
		_, err = tkn.Renew()
		s.EqualError(err, "error renewing the token: error loading the user from the DB: testing error")
	})
}

func (s *TestTokenSuite) TestString() {
	s.Equal("imaginethisisatoken", auth.Token("imaginethisisatoken").String())
}
