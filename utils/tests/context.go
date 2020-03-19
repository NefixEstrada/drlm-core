// SPDX-License-Identifier: AGPL-3.0-only

package tests

import (
	"github.com/brainupdaters/drlm-core/context"

	"github.com/spf13/afero"
)

// GenerateCtx generates a
func GenerateCtx() *context.Context {
	return &context.Context{
		FS: afero.NewMemMapFs(),
	}
}
