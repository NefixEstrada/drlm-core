package tests

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// GenerateCfg creates a configuration file with the default values
func GenerateCfg(t *testing.T) {
	assert := assert.New(t)
	fs.FS = afero.NewMemMapFs()

	err := afero.WriteFile(fs.FS, "/etc/drlm/core.toml", []byte(`[grpc]
cert_path = "/tls/godev/godev.crt"
key_path = "/tls/godev/godev.key"`), 0644)
	assert.Nil(err)

	cfg.Init("")
}
