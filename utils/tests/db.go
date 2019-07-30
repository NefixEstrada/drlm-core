package tests

import (
	"testing"

	dB "github.com/brainupdaters/drlm-core/db"

	"github.com/jinzhu/gorm"
	mocket "github.com/selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
)

// GenerateDB generates a new Mock DB
func GenerateDB(t *testing.T) {
	assert := assert.New(t)

	mocket.Catcher.Register()
	mocket.Catcher.Logging = true

	var err error
	dB.DB, err = gorm.Open(mocket.DriverName, "")
	assert.Nil(err)
}
