// SPDX-License-Identifier: AGPL-3.0-only

package cfg_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func assertCfg(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(50051, cfg.Config.GRPC.Port)
	assert.Equal(true, cfg.Config.GRPC.TLS)
	assert.Equal("cert/server.crt", cfg.Config.GRPC.CertPath)
	assert.Equal("cert/server.key", cfg.Config.GRPC.KeyPath)

	assert.Equal(14, cfg.Config.Security.BcryptCost)
	assert.Equal("", cfg.Config.Security.TokensSecret)
	assert.Equal(5*time.Minute, cfg.Config.Security.TokensLifespan)
	assert.Equal(240*time.Hour, cfg.Config.Security.LoginLifespan)
	assert.Equal("./ssh", cfg.Config.Security.SSHKeysPath)

	assert.Equal("mariadb", cfg.Config.DB.Host)
	assert.Equal(3306, cfg.Config.DB.Port)
	assert.Equal("drlm3", cfg.Config.DB.Usr)
	assert.Equal("drlm3db", cfg.Config.DB.Pwd)
	assert.Equal("drlm3", cfg.Config.DB.DB)

	assert.Equal("minio", cfg.Config.Minio.Host)
	assert.Equal(9443, cfg.Config.Minio.Port)
	assert.Equal(true, cfg.Config.Minio.SSL)
	assert.Equal("cert/minio.crt", cfg.Config.Minio.CertPath)
	assert.Equal("drlm3minio", cfg.Config.Minio.AccessKey)
	assert.Equal("drlm3minio", cfg.Config.Minio.SecretKey)
	assert.Equal("eu-west-3", cfg.Config.Minio.Location)

	assert.Equal("info", cfg.Config.Log.Level)
	assert.Equal("/var/log/drlm/core.log", cfg.Config.Log.File)
}

type TestCfgSuite struct {
	test.Test
}

func TestCfg(t *testing.T) {
	suite.Run(t, new(TestCfgSuite))
}

func (s *TestCfgSuite) TestInit() {
	s.Run("should work as expected", func() {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/etc/drlm/core.toml", nil, 0644)
		s.Nil(err)

		cfg.Init("")

		assertCfg(s.T())
	})

	s.Run("should work as expected with a specified configuration file", func() {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/core.toml", nil, 0644)
		s.Nil(err)

		cfg.Init("/core.toml")

		assertCfg(s.T())
	})

	s.Run("should reload the configuration correctly", func() {
		fs.FS = afero.NewOsFs()

		d, err := afero.TempDir(fs.FS, "", "drlm-core-cfg-reload")
		s.Nil(err)

		defer fs.FS.RemoveAll(d)

		cfgFile := filepath.Join(d, "core.toml")

		err = afero.WriteFile(fs.FS, cfgFile, nil, 0644)
		s.Nil(err)

		cfg.Init(cfgFile)

		assertCfg(s.T())

		err = afero.WriteFile(fs.FS, cfgFile, []byte(`[grpc]
port = 1312`), 0644)
		s.Nil(err)

		time.Sleep(1 * time.Second)

		s.Equal(1312, cfg.Config.GRPC.Port)
	})

	s.Run("should exit if there's an error decoding the configuration", func() {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/etc/drlm/core.json", []byte("invalid config"), 0644)
		s.Nil(err)

		s.Exits(func() { cfg.Init("") })
	})
}
