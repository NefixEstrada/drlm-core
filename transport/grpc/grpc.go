// SPDX-License-Identifier: AGPL-3.0-only

package grpc

import (
	stdContext "context"
	"fmt"
	"net"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/context"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// API is the API version of the server
const API = "v1.0.0"

// CoreServer is the implementation of the DRLM Core GRPC server
type CoreServer struct {
	ctx *context.Context
}

// NewCoreServer returns a new CoreServer struct with the context inside
func NewCoreServer(ctx *context.Context) *CoreServer {
	return &CoreServer{ctx}
}

// Serve starts the DRLM Core GRPC server
func Serve(ctx *context.Context) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", ctx.Cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("error listening at port %d: %v", ctx.Cfg.GRPC.Port, err)
	}

	c := NewCoreServer(ctx)

	var opts = []gRPC.ServerOption{
		gRPC.UnaryInterceptor(c.unaryInterceptor),
		gRPC.StreamInterceptor(c.streamInterceptor),
	}

	if ctx.Cfg.GRPC.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			ctx.Cfg.GRPC.CertPath,
			ctx.Cfg.GRPC.KeyPath,
		)
		if err != nil {
			log.Fatalf("error loading the TLS credentials: %v", err)
		}

		opts = append(opts, gRPC.Creds(creds))
	}

	grpcServer := gRPC.NewServer(opts...)
	drlm.RegisterDRLMServer(grpcServer, c)

	log.Infof("DRLM Core listenning at port :%d", ctx.Cfg.GRPC.Port)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			if err != gRPC.ErrServerStopped {
				log.Fatalf("error serving DRLM Core GRPC: %v", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		grpcServer.Stop()
		ctx.WG.Done()
	}
}

func (c *CoreServer) unaryInterceptor(inCtx stdContext.Context, req interface{}, info *gRPC.UnaryServerInfo, handler gRPC.UnaryHandler) (interface{}, error) {
	// If it's not the login method, check for the token authenticity
	if info.FullMethod != "/drlm.DRLM/UserLogin" && info.FullMethod != "/drlm.DRLM/UserTokenRenew" {
		if err := checkAuth(c.ctx, inCtx); err != nil {
			return nil, err
		}
	}

	return handler(inCtx, req)
}

func (c *CoreServer) streamInterceptor(srv interface{}, stream gRPC.ServerStream, info *gRPC.StreamServerInfo, handler gRPC.StreamHandler) error {
	// DEV NOTE: The AgentConnection does the agent authentication check because there's no other way to send the host (that gets
	// 			 returned from the auth check) to the AgentConnection function
	if info.FullMethod != "/drlm.DRLM/AgentConnection" {
		if err := checkAuth(c.ctx, stream.Context()); err != nil {
			return err
		}
	}

	return handler(srv, stream)
}

func checkAuth(ctx *context.Context, inCtx stdContext.Context) error {
	if md, ok := metadata.FromIncomingContext(inCtx); ok {
		if len(md.Get("tkn")) > 0 {
			tkn := auth.Token(md.Get("tkn")[0])
			if !tkn.Validate(ctx) {
				return status.Error(codes.InvalidArgument, "invalid token")
			}

			return nil
		}
	}

	return status.Error(codes.Unauthenticated, "not authenticated")
}

func parseAuthType(t types.Type) drlm.AuthType {
	switch t {
	case types.Local:
		return drlm.AuthType_AUTH_LOCAL

	default:
		return drlm.AuthType_AUTH_UNKNOWN
	}
}
