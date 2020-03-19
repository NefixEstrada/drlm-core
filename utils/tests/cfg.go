// SPDX-License-Identifier: AGPL-3.0-only

package tests

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/context"

	"github.com/brainupdaters/drlm-common/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// GenerateCfg creates a configuration file with the default values
func GenerateCfg(t *testing.T, ctx *context.Context) {
	require := require.New(t)

	s := test.Test{}
	s.SetT(t)

	err := afero.WriteFile(ctx.FS, "/etc/drlm/core.toml", []byte(`[grpc]
cert_path = "/tls/godev/godev.crt"
key_path = "/tls/godev/godev.key"

[security]
bcrypt_cost = 1
tokens_secret = "Ⓐ"

[minio]
host = "127.0.0.1"
ssl = false`), 0644)
	require.Nil(err)

	cfg.Init(ctx, "")
}
