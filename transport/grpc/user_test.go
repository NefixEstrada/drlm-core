package grpc_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/DATA-DOG/go-sqlmock"
	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TestUserSuite struct {
	suite.Suite
	c    *grpc.CoreServer
	ctx  context.Context
	mock sqlmock.Sqlmock
}

func (s *TestUserSuite) SetupTest() {
	s.c = &grpc.CoreServer{}
	s.ctx = context.Background()
	s.mock = tests.GenerateDB(s.T())
	tests.GenerateCfg(s.T())
}

func (s *TestUserSuite) AfterTest() {
	s.Nil(s.mock.ExpectationsWereMet())
}

func TestUser(t *testing.T) {
	suite.Run(t, new(TestUserSuite))
}

func (s *TestUserSuite) TestLogin() {
	s.Run("should return the token correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "$2y$12$JGfbXRGMBgDxMVhR9tT6B.C3xmAFM1BxkHD6.F0eUS5ugGXcZ5mUq", types.Local),
		)

		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		rsp, err := s.c.UserLogin(s.ctx, req)

		s.Nil(err)
		s.NotNil(rsp.Tkn)
	})

	// s.Run("should return an error if the user is not found", func(t *testing.T) {
	// 	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{}))

	// 	req := &drlm.UserLoginRequest{
	// 		Usr: "nefix",
	// 		Pwd: "f0cKt3Rf$",
	// 	}

	// 	rsp, err := s.c.UserLogin(s.ctx, req)

	// 	s.Equal(status.Error(codes.NotFound, `error logging in: error loading the user from the DB: user "nefix" not found`), err)
	// 	s.Equal(&drlm.UserLoginResponse{}, rsp)
	// })

	s.Run("should return an error if the password isn't correct", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "$2y$12$JGfbXRGMBgDxMVhR9tT6B.C3xmAFM1BxkHD6.F0eUS5ugGXcZ5mUq", types.Local),
		)

		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0CKt3Rf$",
		}

		rsp, err := s.c.UserLogin(s.ctx, req)

		s.Equal(status.Error(codes.Unauthenticated, "error logging in: incorrect password"), err)
		s.Equal(&drlm.UserLoginResponse{}, rsp)
	})

	s.Run("should return an error if there's an error logging in", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
			AddRow(1, "nefix", "f0cKt3Rf$", types.Local),
		)

		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		rsp, err := s.c.UserLogin(s.ctx, req)

		s.Equal(status.Error(codes.Unknown, "error logging in: password error: crypto/bcrypt: hashedSecret too short to be a bcrypted password"), err)
		s.Equal(&drlm.UserLoginResponse{}, rsp)
	})
}

// func TestUserLogin(t *testing.T) {
// 	assert := assert.New(t)

// 	t.Run("should return an error if the user is not found", func(t *testing.T) {
// 		tests.GenerateCfg(t)
// 		tests.GenerateDB(t)

// 		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply(nil).OneTime()

// 		ctx := context.Background()
// 		req := &drlm.UserLoginRequest{
// 			Usr: "nefix",
// 			Pwd: "f0cKt3Rf$",
// 		}

// 		c := grpc.CoreServer{}
// 		rsp, err := c.UserLogin(ctx, req)

// 		assert.Equal(status.Error(codes.NotFound, `error logging in: user "nefix" not found`), err)
// 		assert.Equal(&drlm.UserLoginResponse{}, rsp)
// 	})
// }

func (s *TestUserSuite) TestTokenRenew() {
	s.Run("should renew the token correctly", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WithArgs("nefix").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "updated_at", "created_at"}).
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

		ctx := metadata.NewIncomingContext(s.ctx, metadata.Pairs("tkn", signedTkn))

		rsp, err := s.c.UserTokenRenew(ctx, &drlm.UserTokenRenewRequest{})

		s.Nil(err)
		s.NotEqual("", rsp.Tkn)
		s.True(time.Unix(rsp.TknExpiration.Seconds, 0).After(originalExpirationTime))
	})

	s.Run("should return an error if there's an error renewing the token", func() {
		ctx := metadata.NewIncomingContext(s.ctx, metadata.Pairs("tkn", "invalid token"))

		rsp, err := s.c.UserTokenRenew(ctx, &drlm.UserTokenRenewRequest{})

		s.Equal(status.Error(codes.Unknown, "error renewing the token: the token is invalid or can't be renewed"), err)
		s.Equal(&drlm.UserTokenRenewResponse{}, rsp)
	})

	s.Run("should return an error if no token was provided", func() {
		rsp, err := s.c.UserTokenRenew(s.ctx, &drlm.UserTokenRenewRequest{})

		s.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
		s.Equal(&drlm.UserTokenRenewResponse{}, rsp)
	})
}

func (s *TestUserSuite) TestAdd() {
	s.Run("should add the new user", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "users"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "nefix", tests.DBAnyEncryptedPassword{}, types.Local).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		s.mock.ExpectCommit()

		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		rsp, err := s.c.UserAdd(s.ctx, req)

		s.Nil(err)
		s.Equal(&drlm.UserAddResponse{}, rsp)
	})

	s.Run("should return an error if the password is too weak", func() {
		s.mock.ExpectBegin()

		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "",
		}

		rsp, err := s.c.UserAdd(s.ctx, req)

		s.Equal(status.Error(codes.InvalidArgument, "the password requires, at least, a length of 8 characters"), err)
		s.Equal(&drlm.UserAddResponse{}, rsp)
	})

	s.Run("should return an error if there's an error adding the user to the DB", func() {
		s.mock.ExpectBegin()
		s.mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "users"."id"`)).WithArgs(tests.DBAnyTime{}, tests.DBAnyTime{}, nil, "nefix", tests.DBAnyEncryptedPassword{}, types.Local).WillReturnError(errors.New("testing error"))

		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		rsp, err := s.c.UserAdd(s.ctx, req)

		s.Equal(status.Error(codes.Unknown, "error adding the user to the DB: testing error"), err)
		s.Equal(&drlm.UserAddResponse{}, rsp)
	})
}

func (s *TestUserSuite) TestDelete() {
	// s.Run("should delete the user from the DB correctly", func() {
	// 	s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL AND ((username = $1)) ORDER BY "users"."id" ASC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "auth_type"}).
	// 		AddRow(1, "nefix", "f0cKt3Rf$", types.Local),
	// 	)
	// 	s.mock.ExpectBegin()
	// 	s.mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1  WHERE "users"."deleted_at" IS NULL AND "users"."id" = $2`)).WithArgs(tests.DBAnyTime{}, 1)

	// 	req := &drlm.UserDeleteRequest{
	// 		Usr: "nefix",
	// 	}

	// 	rsp, err := s.c.UserDelete(s.ctx, req)

	// 	s.Nil(err)
	// 	s.Equal(&drlm.UserDeleteResponse{}, rsp)
	// })

	s.Run("should return a not found error if the user isn't found in the DB", func() {

	})

	s.Run("should return an error if there's an error deleting the user", func() {

	})
}

// func TestUserDelete(t *testing.T) {
// 	assert := assert.New(t)

// 	t.Run("should delete the user from the DB correctly", func(t *testing.T) {
// 		tests.GenerateDB(t)

// 		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
// 			"id":        1,
// 			"username":  "nefix",
// 			"password":  "f0cKt3Rf$",
// 			"auth_type": 1,
// 		}}).OneTime()
// 		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithReply([]map[string]interface{}{}).OneTime()

// 		ctx := context.Background()
// 		req := &drlm.UserDeleteRequest{
// 			Usr: "nefix",
// 		}

// 		c := grpc.CoreServer{}
// 		rsp, err := c.UserDelete(ctx, req)

// 		assert.Nil(err)
// 		assert.Equal(&drlm.UserDeleteResponse{}, rsp)
// 	})

// t.Run("should return a not found error if the user isn't in the DB", func(t *testing.T) {
// 	tests.GenerateDB(t)

// 	mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply(nil).OneTime()

// 	ctx := context.Background()
// 	req := &drlm.UserDeleteRequest{
// 		Usr: "nefix",
// 	}

// 	c := grpc.CoreServer{}
// 	rsp, err := c.UserDelete(ctx, req)

// 	assert.Equal(status.Error(codes.NotFound, `error deleting the user "nefix": not found`), err)
// 	assert.Equal(&drlm.UserDeleteResponse{}, rsp)
// })

// 	t.Run("should return an error if there's an error deleting the user", func(t *testing.T) {
// 		tests.GenerateDB(t)

// 		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE "users"."deleted_at" IS NULL AND ((username = nefix)) ORDER BY "users"."id" ASC LIMIT 1`).WithReply([]map[string]interface{}{{
// 			"id":        1,
// 			"username":  "nefix",
// 			"password":  "f0cKt3Rf$",
// 			"auth_type": 1,
// 		}}).OneTime()
// 		mocket.Catcher.NewMock().WithQuery(`UPDATE "users" SET "deleted_at"=?  WHERE "users"."deleted_at" IS NULL AND "users"."id" = ?`).WithError(errors.New("testing error")).OneTime()

// 		ctx := context.Background()
// 		req := &drlm.UserDeleteRequest{
// 			Usr: "nefix",
// 		}

// 		c := grpc.CoreServer{}
// 		rsp, err := c.UserDelete(ctx, req)

// 		assert.Equal(status.Error(codes.Unknown, `error deleting the user "nefix": testing error`), err)
// 		assert.Equal(&drlm.UserDeleteResponse{}, rsp)
// 	})
// }

func (s *TestUserSuite) TestList() {
	s.Run("should return the list of users correctly", func() {
		now := time.Now()

		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT created_at, updated_at, username, auth_type FROM "users" WHERE "users"."deleted_at" IS NULL`)).WillReturnRows(sqlmock.NewRows([]string{"username", "auth_type", "created_at", "updated_at"}).
			AddRow("nefix", types.Local, now, now).
			AddRow("admin", types.Local, now, now).
			AddRow("notnefix", types.Unknown, now, now),
		)

		rsp, err := s.c.UserList(s.ctx, &drlm.UserListRequest{})

		s.Nil(err)
		s.Equal(&drlm.UserListResponse{
			Users: []*drlm.UserListResponse_User{
				&drlm.UserListResponse_User{
					Usr:       "nefix",
					AuthType:  drlm.AuthType_AUTH_LOCAL,
					CreatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
					UpdatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
				},
				&drlm.UserListResponse_User{
					Usr:       "admin",
					AuthType:  drlm.AuthType_AUTH_LOCAL,
					CreatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
					UpdatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
				},
				&drlm.UserListResponse_User{
					Usr:       "notnefix",
					AuthType:  drlm.AuthType_AUTH_UNKNOWN,
					CreatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
					UpdatedAt: &timestamp.Timestamp{Seconds: now.Unix()},
				},
			},
		}, rsp)
	})

	s.Run("should return an error if there's an error listing the users", func() {
		s.mock.ExpectQuery(regexp.QuoteMeta(`SELECT created_at, updated_at, username, auth_type FROM "users" WHERE "users"."deleted_at" IS NULL`)).WillReturnError(errors.New("testing error"))

		rsp, err := s.c.UserList(s.ctx, &drlm.UserListRequest{})

		s.Equal(status.Error(codes.Unknown, "error getting the list of users: testing error"), err)
		s.Equal(&drlm.UserListResponse{}, rsp)
	})
}
