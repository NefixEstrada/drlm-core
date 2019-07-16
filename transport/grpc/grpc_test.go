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

	cmnTests "github.com/brainupdaters/drlm-common/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	assert := assert.New(t)

	t.Run("should start the GRPC server with TLS correctly", func(t *testing.T) {
		tests.GenerateCfg(t)

		port := cmnTests.GetFreePort(t)
		cfg.Config.GRPC.Port = port

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, "wg", &wg)

		go grpc.Serve(ctx)
		wg.Add(1)

		cancel()
	})

	t.Run("should exit if there's an error listening to the port", func(t *testing.T) {
		tests.GenerateCfg(t)

		port := cmnTests.GetFreePort(t)
		cfg.Config.GRPC.Port = port

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Config.GRPC.Port))
		assert.Nil(err)

		cmnTests.AssertExits(t, func() { grpc.Serve(context.Background()) })

		assert.Nil(lis.Close())
	})

	t.Run("should exit if there's an error loading the TLS credentials", func(t *testing.T) {
		tests.GenerateCfg(t)

		port := cmnTests.GetFreePort(t)
		cfg.Config.GRPC.Port = port

		cfg.Config.GRPC.CertPath = "/notacert.go"

		cmnTests.AssertExits(t, func() { grpc.Serve(context.Background()) })
	})
}
