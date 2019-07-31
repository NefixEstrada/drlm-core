package models_test

import (
	"errors"
	"testing"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
)

func TestUserAdd(t *testing.T) {
	assert := assert.New(t)

	t.Run("should add the user to the DB", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES(?,?,?,?,?,?)`).WithReply([]map[string]interface{}{}).OneTime()

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
			AuthType: types.Local,
		}

		assert.Nil(u.Add())
	})

	t.Run("should return an error if the password is too weak", func(t *testing.T) {
		tests.GenerateCfg(t)

		u := models.User{
			Username: "nefix",
			Password: "",
			AuthType: types.Local,
		}

		assert.EqualError(u.Add(), "the password requires, at least, a length of 8 characters")
	})

	t.Run("should return an error if there's an error adding the user to the DB", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`INSERT  INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES (?,?,?,?,?,?)`).WithError(errors.New("testing error")).OneTime()

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
			AuthType: types.Local,
		}

		assert.EqualError(u.Add(), "error adding the user to the DB: testing error")
	})
}

func TestUserLoad(t *testing.T) {
	assert := assert.New(t)

	t.Run("should load the user from the DB correctly", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "f0cKt3Rf$",
			"auth_type": 0,
		}}).OneTime()

		u := models.User{
			Username: "nefix",
		}

		assert.Nil(u.Load())

		assert.Equal(uint(1), u.ID)
		assert.Equal("nefix", u.Username)
		assert.Equal("f0cKt3Rf$", u.Password)
		assert.Equal(types.Local, u.AuthType)
	})

	t.Run("should return an error if there's an error loading the user from the DB", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithError(errors.New("testing error")).OneTime()

		u := models.User{
			Username: "nefix",
		}

		assert.EqualError(u.Load(), "testing error")
	})
}

func TestUserDelete(t *testing.T) {
	assert := assert.New(t)

	t.Run("should delete the user correctly", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "f0cKt3Rf$",
			"auth_type": 0,
		}}).OneTime()
		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithReply([]map[string]interface{}{}).OneTime()

		u := models.User{
			Username: "nefix",
		}

		assert.Nil(u.Delete())
	})

	t.Run("should return an error if there's an error loading the user", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithError(errors.New("testing error")).OneTime()

		u := models.User{
			Username: "nefix",
		}

		assert.EqualError(u.Delete(), "testing error")
	})

	t.Run("should return an error if there's an error deleting the user", func(t *testing.T) {
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "f0cKt3Rf$",
			"auth_type": 0,
		}}).OneTime()
		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithError(errors.New("testing error")).OneTime()

		u := models.User{
			Username: "nefix",
		}

		assert.EqualError(u.Delete(), "testing error")
	})

}

func TestUserBeforeSave(t *testing.T) {
	assert := assert.New(t)

	t.Run("should encrypt the password correctly", func(t *testing.T) {
		tests.GenerateCfg(t)

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3Rf$",
		}

		assert.Nil(u.BeforeSave())
		assert.NotEqual("f0cKt3Rf$", u.Password)
		assert.NotNil(u.Password)
	})

	t.Run("should return an error if the password is too weak", func(t *testing.T) {
		tests.GenerateCfg(t)

		u := models.User{
			Username: "nefix",
			Password: "",
		}

		assert.EqualError(u.BeforeSave(), "the password requires, at least, a length of 8 characters")
	})
}
