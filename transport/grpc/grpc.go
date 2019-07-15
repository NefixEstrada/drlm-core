package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/brainupdaters/drlm-core/cfg"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func unaryInterceptor(ctx context.Context, req interface{}, info *gRPC.UnaryServerInfo, handler gRPC.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func streamInterceptor(srv interface{}, stream gRPC.ServerStream, info *gRPC.StreamServerInfo, handler gRPC.StreamHandler) error {
	return handler(srv, stream)
}

// Serve starts the DRLM Core GRPC server
func Serve() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Config.GRPC.Port))
	if err != nil {
		log.Fatalf("error listening at port %d: %v", cfg.Config.GRPC.Port, err)
	}

	var grpcServer *gRPC.Server
	if cfg.Config.GRPC.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			cfg.Config.GRPC.CertPath,
			cfg.Config.GRPC.KeyPath,
		)
		if err != nil {
			log.Fatalf("error loading the TLS credentials: %v", err)
		}

		grpcServer = gRPC.NewServer(
			gRPC.UnaryInterceptor(unaryInterceptor),
			gRPC.StreamInterceptor(streamInterceptor),
			gRPC.Creds(creds),
		)

	} else {
		grpcServer = gRPC.NewServer(
			gRPC.UnaryInterceptor(unaryInterceptor),
			gRPC.StreamInterceptor(streamInterceptor),
		)
	}

	drlm.RegisterDRLMServer(grpcServer, &CoreServer{})

	log.Infof("DRLM Core listenning at port :%d", cfg.Config.GRPC.Port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("error serving DRLM Core GRPC: %v", err)
	}
}
