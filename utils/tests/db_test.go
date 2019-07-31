package tests_test

import (
	"testing"

	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDB(t *testing.T) {
	assert := assert.New(t)

	tests.GenerateDB(t)

	assert.Equal("common", db.DB.Dialect().GetName())
}
