// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/utils/tests"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TestGRPCInternalSuite struct {
	suite.Suite
	c   *CoreServer
	ctx context.Context
}

func (s *TestGRPCInternalSuite) SetupTest() {
	s.c = &CoreServer{tests.GenerateCtx()}
	s.ctx = context.Background()
}

func TestGRPCInternal(t *testing.T) {
	suite.Run(t, &TestGRPCInternalSuite{})
}

func (s *TestGRPCInternalSuite) TestUnaryInterceptor() {
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "", errors.New("failed! :0")
	}

	s.Run("should check if the user is authenticated before executing the handler", func() {
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserAdd",
		}

		rsp, err := s.c.unaryInterceptor(s.ctx, "", info, h)

		s.Nil(rsp)
		s.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
	})

	s.Run("should not check for the authentication if the method is the UserLogin method", func() {
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserLogin",
		}

		rsp, err := s.c.unaryInterceptor(s.ctx, "", info, h)

		s.Equal("", rsp)
		s.EqualError(err, "failed! :0")
	})

	s.Run("should not check for the authentication if the method is the UserTokenRenew method", func() {
		ctx := context.Background()
		info := &gRPC.UnaryServerInfo{
			FullMethod: "/drlm.DRLM/UserTokenRenew",
		}

		rsp, err := s.c.unaryInterceptor(ctx, "", info, h)

		s.Equal("", rsp)
		s.EqualError(err, "failed! :0")
	})
}

func (s *TestGRPCInternalSuite) TestCheckAuth() {
	s.Run("should return nil if the token is valid", func() {
		tests.GenerateCfg(s.T(), s.c.ctx)

		signedTkn, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &auth.TokenClaims{
			Usr: "nefix",
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(s.c.ctx.Cfg.Security.TokensLifespan).Unix(),
			},
		}).SignedString([]byte(s.c.ctx.Cfg.Security.TokensSecret))
		s.NoError(err)

		inCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", signedTkn))

		err = checkAuth(s.c.ctx, inCtx)

		s.NoError(err)
	})

	s.Run("should return an error if the token is invalid", func() {
		inCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("tkn", "invalid token"))

		err := checkAuth(s.c.ctx, inCtx)

		s.Equal(status.Error(codes.InvalidArgument, "invalid token"), err)
	})

	s.Run("should return an error if there's no token", func() {
		err := checkAuth(s.c.ctx, context.Background())

		s.Equal(status.Error(codes.Unauthenticated, "not authenticated"), err)
	})
}

func (s *TestGRPCInternalSuite) TestParseAuthType() {
	t := []struct {
		in  types.Type
		out drlm.AuthType
	}{
		{
			types.Unknown,
			drlm.AuthType_AUTH_UNKNOWN,
		},
		{
			types.Local,
			drlm.AuthType_AUTH_LOCAL,
		},
	}

	for _, tt := range t {
		s.Equal(tt.out, parseAuthType(tt.in))
	}
}
