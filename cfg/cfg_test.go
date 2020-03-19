// SPDX-License-Identifier: AGPL-3.0-only

package cfg_test

import (
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func assertCfg(t *testing.T, ctx *context.Context) {
	assert := assert.New(t)

	assert.Equal(50051, ctx.Cfg.GRPC.Port)
	assert.Equal(true, ctx.Cfg.GRPC.TLS)
	assert.Equal("cert/server.crt", ctx.Cfg.GRPC.CertPath)
	assert.Equal("cert/server.key", ctx.Cfg.GRPC.KeyPath)

	assert.Equal(14, ctx.Cfg.Security.BcryptCost)
	assert.Equal("", ctx.Cfg.Security.TokensSecret)
	assert.Equal(5*time.Minute, ctx.Cfg.Security.TokensLifespan)
	assert.Equal(240*time.Hour, ctx.Cfg.Security.LoginLifespan)
	assert.Equal("./ssh", ctx.Cfg.Security.SSHKeysPath)

	assert.Equal("mariadb", ctx.Cfg.DB.Host)
	assert.Equal(3306, ctx.Cfg.DB.Port)
	assert.Equal("drlm3", ctx.Cfg.DB.Usr)
	assert.Equal("drlm3db", ctx.Cfg.DB.Pwd)
	assert.Equal("drlm3", ctx.Cfg.DB.DB)

	assert.Equal("minio", ctx.Cfg.Minio.Host)
	assert.Equal(9443, ctx.Cfg.Minio.Port)
	assert.Equal(true, ctx.Cfg.Minio.SSL)
	assert.Equal("cert/minio.crt", ctx.Cfg.Minio.CertPath)
	assert.Equal("drlm3minio", ctx.Cfg.Minio.AccessKey)
	assert.Equal("drlm3minio", ctx.Cfg.Minio.SecretKey)
	assert.Equal("eu-west-3", ctx.Cfg.Minio.Location)

	assert.Equal("info", ctx.Cfg.Log.Level)
	assert.Equal("/var/log/drlm/core.log", ctx.Cfg.Log.File)
}

type TestCfgSuite struct {
	test.Test
}

func TestCfg(t *testing.T) {
	suite.Run(t, new(TestCfgSuite))
}

func (s *TestCfgSuite) TestInit() {
	s.Run("should work as expected", func() {
		ctx := tests.GenerateCtx()

		err := afero.WriteFile(ctx.FS, "/etc/drlm/core.toml", nil, 0644)
		s.Nil(err)

		cfg.Init(ctx, "")

		assertCfg(s.T(), ctx)
	})

	s.Run("should work as expected with a specified configuration file", func() {
		ctx := tests.GenerateCtx()

		err := afero.WriteFile(ctx.FS, "/core.toml", nil, 0644)
		s.Nil(err)

		cfg.Init(ctx, "/core.toml")

		assertCfg(s.T(), ctx)
	})

	s.Run("should work as expected without configuration file", func() {
		ctx := tests.GenerateCtx()

		cfg.Init(ctx, "")

		assertCfg(s.T(), ctx)
	})

	s.Run("should exit if there's an error decoding the configuration", func() {
		ctx := tests.GenerateCtx()

		err := afero.WriteFile(ctx.FS, "/etc/drlm/core.json", []byte("invalid config"), 0644)
		s.Nil(err)

		s.Exits(func() { cfg.Init(ctx, "") })
	})
}
