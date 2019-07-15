package cfg_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/utils/tests"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)

	t.Run("should work as expected", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/etc/drlm/core.toml", nil, 0644)
		assert.Nil(err)

		cfg.Init("")

		assert.Equal(50051, cfg.Config.GRPC.Port)
		assert.Equal(true, cfg.Config.GRPC.TLS)
		assert.Equal("cert/server.crt", cfg.Config.GRPC.CertPath)
		assert.Equal("cert/server.key", cfg.Config.GRPC.KeyPath)

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
	})

	t.Run("should work as expected with a specified configuration file", func(t *testing.T) {
		fs.FS = afero.NewMemMapFs()

		err := afero.WriteFile(fs.FS, "/core.toml", nil, 0644)
		assert.Nil(err)

		cfg.Init("/core.toml")

		assert.Equal(50051, cfg.Config.GRPC.Port)
		assert.Equal(true, cfg.Config.GRPC.TLS)
		assert.Equal("cert/server.crt", cfg.Config.GRPC.CertPath)
		assert.Equal("cert/server.key", cfg.Config.GRPC.KeyPath)

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
