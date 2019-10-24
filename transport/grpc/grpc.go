package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/brainupdaters/drlm-core/auth"
	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/cfg"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// API is the API version of the server
var API = "v1.0.0"

// CoreServer is the implementation of the DRLM Core GRPC server
type CoreServer struct{}

// Serve starts the DRLM Core GRPC server
func Serve(ctx context.Context) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Config.GRPC.Port))
	if err != nil {
		log.Fatalf("error listening at port %d: %v", cfg.Config.GRPC.Port, err)
	}

	var opts = []gRPC.ServerOption{
		gRPC.UnaryInterceptor(unaryInterceptor),
		gRPC.StreamInterceptor(streamInterceptor),
	}

	if cfg.Config.GRPC.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			cfg.Config.GRPC.CertPath,
			cfg.Config.GRPC.KeyPath,
		)
		if err != nil {
			log.Fatalf("error loading the TLS credentials: %v", err)
		}

		opts = append(opts, gRPC.Creds(creds))
	}

	grpcServer := gRPC.NewServer(opts...)
	drlm.RegisterDRLMServer(grpcServer, &CoreServer{})

	log.Infof("DRLM Core listenning at port :%d", cfg.Config.GRPC.Port)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			if err != gRPC.ErrServerStopped {
				log.Fatalf("error serving DRLM Core GRPC: %v", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		ctx.Value("wg").(*sync.WaitGroup).Done()
	}
}

func unaryInterceptor(ctx context.Context, req interface{}, info *gRPC.UnaryServerInfo, handler gRPC.UnaryHandler) (interface{}, error) {
	// If it's not the login method, check for the token authenticity
	if info.FullMethod != "/drlm.DRLM/UserLogin" && info.FullMethod != "/drlm.DRLM/UserTokenRenew" {
		if err := checkAuth(ctx); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}

func streamInterceptor(srv interface{}, stream gRPC.ServerStream, info *gRPC.StreamServerInfo, handler gRPC.StreamHandler) error {
	if err := checkAuth(stream.Context()); err != nil {
		return err
	}

	return handler(srv, stream)
}

func checkAuth(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md.Get("tkn")) > 0 {
			tkn := auth.Token(md.Get("tkn")[0])
			if !tkn.Validate() {
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
