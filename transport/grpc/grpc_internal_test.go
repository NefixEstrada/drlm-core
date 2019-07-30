package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryInterceptor(t *testing.T) {
	assert := assert.New(t)

	t.Run("should check if the user is authenticated before executing the handler", func(t *testing.T) {
		ctx := context.Background()
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserAdd",
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "", errors.New("add failed! :0")
		}

		rsp, err := unaryInterceptor(ctx, "", info, h)

		assert.Nil(rsp)
		assert.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
	})

	t.Run("should not check for the authentication if the method is the UserLogin method", func(t *testing.T) {
		ctx := context.Background()
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserLogin",
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "", errors.New("login failed! :0")
		}

		rsp, err := unaryInterceptor(ctx, "", info, h)

		assert.Equal("", rsp)
		assert.EqualError(err, "login failed! :0")
	})

	t.Run("should not check for the authentication if the method is the UserTokenRenew method", func(t *testing.T) {
		ctx := context.Background()
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserTokenRenew",
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "", errors.New("renew failed! :0")
		}

		rsp, err := unaryInterceptor(ctx, "", info, h)

		assert.Equal("", rsp)
		assert.EqualError(err, "renew failed! :0")
	})
}

func TestCheckAuth(t *testing.T) {
	assert := assert.New(t)

	t.Run("should return nil if the token is valid", func(t *testing.T) {
		tests.GenerateCfg(t)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.Config.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(cfg.Config.Security.TokensSecret))
		assert.Nil(err)

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", signedTkn))

		err = checkAuth(ctx)

		assert.Nil(err)
	})

	t.Run("should return an error if the token is invalid", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", "invalid token"))

		err := checkAuth(ctx)

		assert.Equal(status.Error(codes.InvalidArgument, "invalid token"), err)
	})

	t.Run("should return an error if there's no token", func(t *testing.T) {
		err := checkAuth(context.Background())

		assert.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
	})
}
