// SPDX-License-Identifier: AGPL-3.0-only

package grpc_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/transport/grpc"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
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
		ctx, cancel := context.WithCancel()
		ctx.FS = afero.NewMemMapFs()
		tests.GenerateCfg(s.T(), ctx)

		ctx.Cfg.GRPC.Port = s.FreePort()

		go grpc.Serve(ctx)
		ctx.WG.Add(1)

		cancel()
	})

	s.Run("should exit if there's an error listening to the port", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)

		ctx.Cfg.GRPC.Port = s.FreePort()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", ctx.Cfg.GRPC.Port))
		s.Require().Nil(err)

		s.Exits(func() { grpc.Serve(ctx) })

		s.Require().NoError(lis.Close())
	})

	s.Run("should exit if there's an error loading the TLS credentials", func() {
		ctx := tests.GenerateCtx()
		tests.GenerateCfg(s.T(), ctx)

		ctx.Cfg.GRPC.Port = s.FreePort()
		ctx.Cfg.GRPC.CertPath = "/notacert.go"

		s.Exits(func() { grpc.Serve(ctx) })
	})
}
