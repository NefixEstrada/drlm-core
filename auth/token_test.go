package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/dgrijalva/jwt-go"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
	assert := assert.New(t)

	t.Run("should return a new token", func(t *testing.T) {
		tests.GenerateCfg(t)

		tkn, expiresAt, err := auth.NewToken("nefix")
		assert.Nil(err)
		assert.NotNil(tkn)
		assert.True(expiresAt.After(time.Now()))
	})
}

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	t.Run("should return true if the token is valid", func(t *testing.T) {
		tests.GenerateCfg(t)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.Config.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := auth.Token(signedTkn)
		assert.True(tkn.Validate())
	})

	t.Run("should return false if there's an error parsing the token", func(t *testing.T) {
		tests.GenerateCfg(t)

		tkn := auth.Token("invalid token!")
		assert.False(tkn.Validate())
	})

	t.Run("should return false if the token is invalid", func(t *testing.T) {
		tests.GenerateCfg(t)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-cfg.Config.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := auth.Token(signedTkn)
		assert.False(tkn.Validate())
	})
}

func TestRenew(t *testing.T) {
	assert := assert.New(t)

	t.Run("should renew the token correctly", func(t *testing.T) {
		tests.GenerateCfg(t)

		originalExpirationTime := time.Now().Add(1 * time.Minute)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew()

		assert.Nil(err)
		assert.True(expiresAt.After(originalExpirationTime))
	})

	t.Run("should renew the token if it has expired but the user hasn't been modified since the moment when it was issued", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":         1,
			"username":   "nefix",
			"updated_at": time.Now().Add(-10 * time.Minute),
			"created_at": time.Now().Add(-10 * time.Minute),
		}}).OneTime()

		originalExpirationTime := time.Now().Add(-cfg.Config.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := auth.Token(signedTkn)
		expiresAt, err := tkn.Renew()

		assert.Nil(err)
		assert.True(expiresAt.After(originalExpirationTime))
	})

	t.Run("should return an error if there's an error parsing the token", func(t *testing.T) {
		tkn := auth.Token("invalid token!")

		_, err := tkn.Renew()
		assert.EqualError(err, "error renewing the token: the token is invalid or can't be renewed")
	})

	t.Run("should return an error if the token has expired and there's an error loading the DB user", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithError(errors.New("testing error")).OneTime()

		originalExpirationTime := time.Now().Add(-cfg.Config.Security.TokensLifespan)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		tkn := auth.Token(signedTkn)
		_, err = tkn.Renew()
		assert.EqualError(err, "error renewing the token: error loading the user from the DB: testing error")
	})
}

func TestString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("imaginethisisatoken", auth.Token("imaginethisisatoken").String())
}
