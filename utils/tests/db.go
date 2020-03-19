// SPDX-License-Identifier: AGPL-3.0-only

package tests

import (
	"database/sql/driver"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	// import the gorm postgres dialect
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// GenerateDB generates a new Mock DB
func GenerateDB(t *testing.T, ctx *context.Context) sqlmock.Sqlmock {
	require := require.New(t)

	d, mock, err := sqlmock.New()
	require.NoError(err)

	ctx.DB, err = gorm.Open("postgres", d)
	require.NoError(err)

	ctx.DB.LogMode(false)

	return mock
}

// DBAnyTime is used to mock time.Time in the DB
type DBAnyTime struct{}

// Match is the function responsible for returning whether the mock expression matches or not the expectations
func (a DBAnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// DBAnyEncryptedPassword is used to mock encrypted passwords using bcrypt
type DBAnyEncryptedPassword struct{}

// Match is the function responsible for returning whether the mock expression matches or not the expectations
func (p DBAnyEncryptedPassword) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}

	_, err := bcrypt.Cost([]byte(s))
	if err != nil {
		return false
	}

	return true
}

// DBAnyBucketName is used to mock bucket names (since they have a UID part)
type DBAnyBucketName struct{}

// Match is the function responsible for returning whether the mock expression matches or not the expectations
func (b DBAnyBucketName) Match(v driver.Value) bool {
	bName, ok := v.(string)
	if !ok {
		return false
	}

	return strings.HasPrefix(bName, "drlm-")
}

// DBAnySecret is used to mock a secret
type DBAnySecret struct{}

// Match is the function responsible for returning whether the mock expression matches or not the expectations
func (b DBAnySecret) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}

	_, err := hex.DecodeString(s)
	if err != nil {
		return false
	}

	return len(s) == 32
}
