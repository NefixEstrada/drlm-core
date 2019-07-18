package types_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/auth/types"

	"github.com/stretchr/testify/assert"
)

func TestStirng(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("local", types.Local.String())
	assert.Equal("unknown", types.Type(-1).String())
}
