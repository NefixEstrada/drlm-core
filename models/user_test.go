package models_test

import (
	"errors"
	"testing"

	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"
	mocket "github.com/selvatico/go-mocket"

	"github.com/stretchr/testify/assert"
)

func TestUserAdd(t *testing.T) {
	assert := assert.New(t)

	t.Run("should add the user to the DB", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password") VALUES(?,?,?,?,?)`).WithReply([]map[string]interface{}{})

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
		}

		assert.Nil(u.Add())
	})

	t.Run("should return an error if there's an error adding the user to the DB", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`INSERT  INTO "users" ("created_at","updated_at","deleted_at","username","password") VALUES (?,?,?,?,?)`).WithError(errors.New("testing error"))

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
		}

		assert.EqualError(u.Add(), "error adding the user to the DB: testing error")
	})
}

func TestUserBeforeSave(t *testing.T) {
	assert := assert.New(t)

	t.Run("should encrypt the password correctly", func(t *testing.T) {
		u := models.User{
			Username: "nefix",
			Password: "f0cKt3Rf$",
		}

		assert.Nil(u.BeforeSave())
		assert.NotEqual(u.Password, "f0cKt3Rf$")
		assert.NotNil(u.Password)
	})

	t.Run("should return an error if the password is too weak", func(t *testing.T) {
		u := models.User{
			Username: "nefix",
			Password: "",
		}

		assert.EqualError(u.BeforeSave(), "password too weak: the password requires, at least, a length of 8 characters")
	})
}
