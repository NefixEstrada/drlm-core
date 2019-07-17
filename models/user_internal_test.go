package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCheckPwdStrength(t *testing.T) {
	assert := assert.New(t)

	t.Run("should validate correct passwords", func(t *testing.T) {
		u := User{Password: "P4$$w0rd!"}
		assert.Nil(u.checkPwdStrength())
	})

	t.Run("should return an error if the password is too short", func(t *testing.T) {
		u := User{Password: "a"}
		assert.EqualError(u.checkPwdStrength(), "the password requires, at least, a length of 8 characters")
	})

	t.Run("should return an error if the password hasn't an uppercase character", func(t *testing.T) {
		u := User{Password: "asdfzxcv"}
		assert.EqualError(u.checkPwdStrength(), "the password requires, at least, an uppercase character")
	})

	t.Run("should return an error if the password hasn't a number", func(t *testing.T) {
		u := User{Password: "asdfzxÇv"}
		assert.EqualError(u.checkPwdStrength(), "the password requires, at least, a number")
	})

	t.Run("should return an error if the password hasn't a special character", func(t *testing.T) {
		u := User{Password: "4sdfzxÇv"}
		assert.EqualError(u.checkPwdStrength(), "the password requires, at least, an special character")
	})
}
