package cfg_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/tests"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func assertCfg(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(50051, cfg.Config.GRPC.Port)
	assert.Equal(true, cfg.Config.GRPC.TLS)
	assert.Equal("cert/server.crt", cfg.Config.GRPC.CertPath)
	assert.Equal("cert/server.key", cfg.Config.GRPC.KeyPath)

	assert.Equal(14, cfg.Config.Security.BcryptCost)
	assert.Equal(5, cfg.Config.Security.TokensLifespan)
	assert.Equal("", cfg.Config.Security.TokensSecret)

	assert.Equal("mariadb", cfg.Config.DB.Host)
	assert.Equal(3306, cfg.Config.DB.Port)
	assert.Equal("drlm3", cfg.Config.DB.Usr)
	assert.Equal("drlm3db", cfg.Config.DB.Pwd)
	assert.Equal("drlm3", cfg.Config.DB.DB)

	assert.Equal("minio", cfg.Config.Minio.Host)
	assert.Equal(9000, cfg.Config.Minio.Port)
	assert.Equal(true, cfg.Config.Minio.SSL)
	assert.Equal("drlm3minio", cfg.Config.Minio.AccessKey)
	assert.Equal("drlm3minio", cfg.Config.Minio.SecretKey)

	assert.Equal("info", cfg.Config.Log.Level)
	assert.Equal("/var/log/drlm/core.log", cfg.Config.Log.File)
}

func TestInit(t *testing.T) {
	assert := assert.New(t)

	t.Run("should work as expected", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/etc/drlm/core.toml", nil, 0644)
		assert.Nil(err)

		cfg.Init("")

		assertCfg(t)
	})

	t.Run("should work as expected with a specified configuration file", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/core.toml", nil, 0644)
		assert.Nil(err)

		cfg.Init("/core.toml")

		assertCfg(t)
	})

	t.Run("should reload the configuration correctly", func(t *testing.T) {
		fs.FS = afero.NewOsFs()

		d, err := afero.TempDir(fs.FS, "", "drlm-core-cfg-reload")
		assert.Nil(err)

		defer fs.FS.RemoveAll(d)

		cfgFile := filepath.Join(d, "core.toml")

		err = afero.WriteFile(fs.FS, cfgFile, nil, 0644)
		assert.Nil(err)

		cfg.Init(cfgFile)

		assertCfg(t)

		err = afero.WriteFile(fs.FS, cfgFile, []byte(`[grpc]
port = 1312`), 0644)
		assert.Nil(err)

		time.Sleep(1 * time.Second)

		assert.Equal(1312, cfg.Config.GRPC.Port)
	})

	t.Run("should fail and exit if there's an error reading the configuration", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		tests.AssertExits(t, func() { cfg.Init("") })
	})

	t.Run("should fail and exit if there's an error decoding the configuration", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/etc/drlm/core.json", []byte("invalid config"), 0644)
		assert.Nil(err)

		tests.AssertExits(t, func() { cfg.Init("") })
	})
}
