package migrations

import (
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/models"
)

// Migrate runs all the DB migrations
func Migrate() {
	// 2019/07/16 10:02 - Add the User model
	db.DB.AutoMigrate(&models.User{})
}
