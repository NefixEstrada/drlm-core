package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/jinzhu/gorm"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginLocal(t *testing.T) {
	assert := assert.New(t)

	t.Run("should return a token if the authentication succeeds", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW",
			"auth_type": 1,
		}}).OneTime()

		tkn, expiresAt, err := auth.LoginLocal("nefix", "f0cKt3Rf$")
		assert.Nil(err)
		assert.NotNil(tkn)
		assert.True(expiresAt.After(time.Now()))
	})

	t.Run("should return an error if the user isn't found", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply(nil).OneTime()

		tkn, _, err := auth.LoginLocal("nefix", "f0cKt3Rf$")
		assert.Equal(gorm.ErrRecordNotFound, err)
		assert.Equal("", tkn.String())
	})

	t.Run("should return an error if there's an error loading the user", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithError(errors.New("testing error")).OneTime()

		tkn, _, err := auth.LoginLocal("nefix", "f0cKt3Rf$")
		assert.EqualError(err, "error loading the user from the DB: testing error")
		assert.Equal("", tkn.String())
	})

	t.Run("should return an error if the login type isn't local", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW",
			"auth_type": -1,
		}}).OneTime()

		tkn, _, err := auth.LoginLocal("nefix", "f0cKt3Rf$")
		assert.EqualError(err, "invalid authentication method: user authentication type is unknown")
		assert.Equal("", tkn.String())
	})

	t.Run("should return an error if the passwords don't match", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "$2y$12$U9o2EJDhZiwCkcP2sk3tSOHtEajHjcw0/izc8WfvqeX2M2YwQLhgW",
			"auth_type": 1,
		}}).OneTime()

		tkn, _, err := auth.LoginLocal("nefix", "asdfzxcv")
		assert.Equal(bcrypt.ErrMismatchedHashAndPassword, err)
		assert.Equal("", tkn.String())
	})

	t.Run("should return an error if there's an error comparing the passwords", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "f0cKt3Rf$",
			"auth_type": 1,
		}}).OneTime()

		tkn, _, err := auth.LoginLocal("nefix", "f0cKt3Rf$")
		assert.EqualError(err, "password error: crypto/bcrypt: hashedSecret too short to be a bcrypted password")
		assert.Equal("", tkn.String())
	})
}
