// SPDX-License-Identifier: AGPL-3.0-only

package tests_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCfg(t *testing.T) {
	assert := assert.New(t)

	tests.GenerateCfg(t)

	exists, err := afero.Exists(fs.FS, "/etc/drlm/core.toml")
	assert.Nil(err)
	assert.True(exists)

	assert.NotNil(cfg.Config)
}
