package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/brainupdaters/drlm-core/cfg"

	drlm "github.com/brainupdaters/drlm-common/pkg/proto"
	log "github.com/sirupsen/logrus"
	gRPC "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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
	return handler(ctx, req)
}

func streamInterceptor(srv interface{}, stream gRPC.ServerStream, info *gRPC.StreamServerInfo, handler gRPC.StreamHandler) error {
	return handler(srv, stream)
}
