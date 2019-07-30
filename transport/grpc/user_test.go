package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/dgrijalva/jwt-go"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

func TestUserTokenRenew(t *testing.T) {
	assert := assert.New(t)

	t.Run("should renew the token correctly", func(t *testing.T) {
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
			Usr:         "nefix",
			FirstIssued: originalExpirationTime,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: originalExpirationTime.Unix(),
				IssuedAt:  originalExpirationTime.Add(-1 * time.Minute).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", signedTkn))
		req := &drlm.UserTokenRenewRequest{}

		c := grpc.CoreServer{}
		rsp, err := c.UserTokenRenew(ctx, req)

		assert.Nil(err)
		assert.NotEqual("", rsp.Tkn)
		assert.True(time.Unix(rsp.TknExpiration.Seconds, 0).After(originalExpirationTime))
	})

	t.Run("should return an error if there's an error renewing the token", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", "invalid tkn"))
		req := &drlm.UserTokenRenewRequest{}

		c := grpc.CoreServer{}
		rsp, err := c.UserTokenRenew(ctx, req)

		assert.Equal(status.Error(codes.Unknown, "error renewing the token: the token is invalid or can't be renewed"), err)
		assert.Equal(&drlm.UserTokenRenewResponse{}, rsp)
	})

	t.Run("should return an error if no token was provided", func(t *testing.T) {
		ctx := context.Background()
		req := &drlm.UserTokenRenewRequest{}

		c := grpc.CoreServer{}
		rsp, err := c.UserTokenRenew(ctx, req)

		assert.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
		assert.Equal(&drlm.UserTokenRenewResponse{}, rsp)
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
