package tests

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/stretchr/testify/require"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/spf13/afero"
)

// GenerateCfg creates a configuration file with the default values
func GenerateCfg(t *testing.T) {
	require := require.New(t)
	fs.FS = afero.NewMemMapFs()

	err := afero.WriteFile(fs.FS, "/etc/drlm/core.toml", []byte(`[grpc]
cert_path = "/tls/godev/godev.crt"
key_path = "/tls/godev/godev.key"

[security]
bcrypt_cost = 1
tokens_secret = "Ⓐ"`), 0644)
	require.Nil(err)

	cfg.Init("")
}
