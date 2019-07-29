package migrations

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/db"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// Migrate runs all the DB migrations
func Migrate() {
	// 2019/07/16 10:02 - Add the User model
	db.DB.AutoMigrate(&models.User{})

	// Create the admin user if it doesn't exist
	u := models.User{Username: "admin"}
	if err := u.Load(); err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("error creating the admin user: %v", err)
		}

		fmt.Print("Please, set the admin password: ")
		bPwd, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("error creating the admin user: error reading the password: %v", err)
		}
		pwd := strings.TrimSpace(string(bPwd))

		fmt.Print("\n")

		u = models.User{
			Username: "admin",
			Password: pwd,
			AuthType: types.Local,
		}

		if err := u.Add(); err != nil {
			if models.IsErrUsrPwdStrength(err) {
				log.Fatalf(err.Error())
			}
		}
	}

}
