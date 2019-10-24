package grpc_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/stretchr/testify/suite"
)

type TestGRPCSuite struct {
	test.Test
}

func TestGRPC(t *testing.T) {
	suite.Run(t, new(TestGRPCSuite))
}

func (s *TestGRPCSuite) TestServe() {
	s.Run("should start the GRPC server with TLS correctly", func() {
		tests.GenerateCfg(s.T())

		port := s.FreePort()
		cfg.Config.GRPC.Port = port

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, "wg", &wg)

		go grpc.Serve(ctx)
		wg.Add(1)

		cancel()
	})

	s.Run("should exit if there's an error listening to the port", func() {
		tests.GenerateCfg(s.T())

		port := s.FreePort()
		cfg.Config.GRPC.Port = port

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Config.GRPC.Port))
		s.Require().Nil(err)

		s.Exits(func() { grpc.Serve(context.Background()) })

		s.Require().Nil(lis.Close())
	})

	s.Run("should exit if there's an error loading the TLS credentials", func() {
		tests.GenerateCfg(s.T())

		port := s.FreePort()
		cfg.Config.GRPC.Port = port

		cfg.Config.GRPC.CertPath = "/notacert.go"

		s.Exits(func() { grpc.Serve(context.Background()) })
	})
}
