package models_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/models"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type TestUserSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
}

func (s *TestUserSuite) SetupTest() {
	s.mock = tests.GenerateDB(s.T())
}

func (s *TestUserSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func TestUser(t *testing.T) {
	suite.Run(t, new(TestUserSuite))
}

func (s *TestUserSuite) TestList() {
	s.Run("should return a list of the users in the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT created_at, updated_at, username, auth_type FROM "users"  WHERE "users"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "auth_type"}).
			AddRow(1, "nefix", types.Local).
			AddRow(2, "admin", types.Local).
			AddRow(3, "notnefix", types.Local),
		)

		expectedUsers := []*models.User{
			&models.User{
				Model: gorm.Model{
					ID: 1,
				},
				Username: "nefix",
				AuthType: types.Local,
			},
			&models.User{
				Model: gorm.Model{
					ID: 2,
				},
				Username: "admin",
				AuthType: types.Local,
			},
			&models.User{
				Model: gorm.Model{
					ID: 3,
				},
				Username: "notnefix",
				AuthType: types.Local,
			},
		}

		users, err := models.UserList()

		s.Nil(err)
		s.Equal(expectedUsers, users)
	})

	s.Run("should return an error if there's an error listing the users in the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT created_at, updated_at, username, auth_type FROM "users"  WHERE "users"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		users, err := models.UserList()

		s.EqualError(err, "error getting the list of users: testing error")
		s.Equal([]*models.User{}, users)
	})
}

func (s *TestUserSuite) TestAdd() {
	tests.GenerateCfg(s.T())

	s.Run("should add the user to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT  INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "users"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "nefix", tests.DBAnyEncryptedPassword{}, types.Local).WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(1),
		)
		s.mock.ExpectCommit()

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
			AuthType: types.Local,
		}

		s.Nil(u.Add())
	})

	s.Run("should return an error if the password is too weak", func() {
		s.mock.ExpectBegin()

		u := models.User{
			Username: "nefix",
			Password: "",
			AuthType: types.Local,
		}

		s.EqualError(u.Add(), "the password requires, at least, a length of 8 characters")
	})

	s.Run("should return an error if there's an error adding the user to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "users"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "nefix", tests.DBAnyEncryptedPassword{}, types.Local).WillReturnError(errors.New("testing error"))

		u := models.User{
			Username: "nefix",
			Password: "f0cKt3rF$",
			AuthType: types.Local,
		}

		s.EqualError(u.Add(), "error adding the user to the DB: testing error")
	})
}

func (s *TestUserSuite) TestLoad() {

	s.Run("should load the user from the DB correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "f0cKt3Rf$", types.Local),
		)

		expectedUser := models.User{
			Model: gorm.Model{
				ID: 1,
			},
			Username: "nefix",
			Password: "f0cKt3Rf$",
			AuthType: types.Local,
		}

		u := models.User{
			Username: "nefix",
		}

		s.Nil(u.Load())
		s.Equal(expectedUser, u)
	})

	s.Run("should return an error if there's an error loading the user from the DB", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnError(errors.New("testing error"))

		u := models.User{
			Username: "nefix",
		}

		s.EqualError(u.Load(), "error loading the user from the DB: testing error")
	})
}

func (s *TestUserSuite) TestDelete() {
	s.Run("should delete the user correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "f0cKt3Rf$", types.Local),
		)
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1  WHERE "users"."deleted_at" IS NULL AND "users"."id" = $2`)).WithArgs(&tests.DBAnyTime{}, 1)

		u := models.User{
			Username: "nefix",
		}

		s.Nil(u.Delete())
	})

	// s.Run("should return an error if there's an error loading the user", func() {
	// 	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnError(errors.New("testing error"))

	// 	u := models.User{
	// 		Username: "nefix",
	// 	}

	// 	s.EqualError(u.Delete(), "error deleting the user from the DB: error loading the user from the DB: testing error")
	// })
}

// func TestUserDelete(t *testing.T) {
// 	assert := assert.New(t)

// 	t.Run("should delete the user correctly", func(t *testing.T) {
// 		tests.GenerateDB(t)

// 		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
// 			"id":        1,
// 			"username":  "nefix",
// 			"password":  "f0cKt3Rf$",
// 			"auth_type": 1,
// 		}}).OneTime()
// 		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithReply([]map[string]interface{}{}).OneTime()

// 		u := models.User{
// 			Username: "nefix",
// 		}

// 		assert.Nil(u.Delete())
// 	})

// 	t.Run("should return an error if there's an error deleting the user", func(t *testing.T) {
// 		tests.GenerateDB(t)

// 		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
// 			"id":        1,
// 			"username":  "nefix",
// 			"password":  "f0cKt3Rf$",
// 			"auth_type": 1,
// 		}}).OneTime()
// 		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithError(errors.New("testing error")).OneTime()

// 		u := models.User{
// 			Username: "nefix",
// 		}

// 		assert.EqualError(u.Delete(), "testing error")
// 	})

// }

func (s *TestUserSuite) TestBeforeSave() {
	tests.GenerateCfg(s.T())

	s.Run("should encrypt the password correctly", func() {
		u := models.User{
			Username: "nefix",
			Password: "f0cKt3Rf$",
		}

		s.Nil(u.BeforeSave())
		_, err := bcrypt.Cost([]byte(u.Password))
		s.Nil(err)
	})

	s.Run("should return an error if the password is too weak", func() {
		u := models.User{
			Username: "nefix",
			Password: "",
		}

		s.EqualError(u.BeforeSave(), "the password requires, at least, a length of 8 characters")
	})
}
