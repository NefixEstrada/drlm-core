package tests_test

import (
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/utils/tests"

	"github.com/stretchr/testify/assert"
)

func TestGenerateDB(t *testing.T) {
	assert := assert.New(t)

	tests.GenerateDB(t)

	assert.Equal("postgres", db.DB.Dialect().GetName())
}

func TestDBAnyTimeMatch(t *testing.T) {
	assert := assert.New(t)

	dbTime := tests.DBAnyTime{}

	t.Run("should return true if the value is a time.Time", func(t *testing.T) {
		assert.True(dbTime.Match(time.Now()))
	})

	t.Run("should return false if the value is not time.Time", func(t *testing.T) {
		assert.False(dbTime.Match("time"))
	})
}

func TestDBAnyPassword(t *testing.T) {
	assert := assert.New(t)

	dbPwd := tests.DBAnyEncryptedPassword{}

	t.Run("should return true if the value is an encrypted bcrypt password", func(t *testing.T) {
		assert.True(dbPwd.Match("$2y$12$YpoiZx9WreZ05J3AygHDf.Veuag2kxSgSeh2nvl5WuOSzL/P2II9O"))
	})

	t.Run("should return false if the value is not an encrypted bcrypt password", func(t *testing.T) {
		assert.False(dbPwd.Match("p4$$w0rd!"))
	})

	t.Run("should return false if the value is not an encrypted bcrypt password", func(t *testing.T) {
		assert.False(dbPwd.Match(1))
	})
}
