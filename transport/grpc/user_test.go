package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserLogin(t *testing.T) {
	assert := assert.New(t)

	t.Run("should return the token correctly", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "$2y$12$JGfbXRGMBgDxMVhR9tT6B.C3xmAFM1BxkHD6.F0eUS5ugGXcZ5mUq",
			"auth_type": 0,
		}}).OneTime()

		ctx := context.Background()
		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserLogin(ctx, req)

		assert.Nil(err)
		assert.NotNil(rsp.Tkn)
	})

	t.Run("should return an error if the user is not found", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply(nil).OneTime()

		ctx := context.Background()
		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserLogin(ctx, req)

		assert.Equal(status.Error(codes.NotFound, `error logging in: user "nefix" not found`), err)
		assert.Equal(&drlm.UserLoginResponse{}, rsp)
	})

	t.Run("should return an error if the password isn't correct", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "$2y$12$JGfbXRGMBgDxMVhR9tT6B.C3xmAFM1BxkHD6.F0eUS5ugGXcZ5mUq",
			"auth_type": 0,
		}}).OneTime()

		ctx := context.Background()
		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0CKt3Rf$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserLogin(ctx, req)

		assert.Equal(status.Error(codes.Unauthenticated, "error logging in: incorrect password"), err)
		assert.Equal(&drlm.UserLoginResponse{}, rsp)
	})

	t.Run("should return an error if there's an error logging in", func(t *testing.T) {
		tests.GenerateCfg(t)
		tests.GenerateDB(t)

		mocket.Catcher.NewMock().WithQuery(`SELECT * FROM "users"  WHERE`).WithReply([]map[string]interface{}{{
			"id":        1,
			"username":  "nefix",
			"password":  "f0cKt3Rf$",
			"auth_type": 0,
		}}).OneTime()

		ctx := context.Background()
		req := &drlm.UserLoginRequest{
			Usr: "nefix",
			Pwd: "f0cKt3Rf$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserLogin(ctx, req)

		assert.Equal(status.Error(codes.Unknown, "error logging in: password error: crypto/bcrypt: hashedSecret too short to be a bcrypted password"), err)
		assert.Equal(&drlm.UserLoginResponse{}, rsp)
	})
}

func TestUserAdd(t *testing.T) {
	assert := assert.New(t)

	t.Run("should add the new user", func(t *testing.T) {
		tests.GenerateDB(t)
		mocket.Catcher.NewMock().WithQuery(`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","password") VALUES(?,?,?,?,?)`).WithReply([]map[string]interface{}{}).OneTime()

		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKT3rF$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Nil(err)
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})

	t.Run("should return an error if the password is too weak", func(t *testing.T) {
		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Equal(err, status.Error(codes.InvalidArgument, "the password requires, at least, a length of 8 characters"))
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})

	t.Run("should return an error if there's an error adding the user to the DB", func(t *testing.T) {
		tests.GenerateDB(t)
		mocket.Catcher.NewMock().WithQuery(`INSERT  INTO "users" ("created_at","updated_at","deleted_at","username","password","auth_type") VALUES (?,?,?,?,?,?)`).WithError(errors.New("testing error")).OneTime()

		ctx := context.Background()
		req := &drlm.UserAddRequest{
			Usr: "nefix",
			Pwd: "f0cKT3rF$",
		}

		c := grpc.CoreServer{}
		rsp, err := c.UserAdd(ctx, req)

		assert.Equal(status.Error(codes.Unknown, "error adding the user to the DB: testing error"), err)
		assert.Equal(&drlm.UserAddResponse{}, rsp)
	})
}
