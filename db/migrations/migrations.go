// SPDX-License-Identifier: AGPL-3.0-only

package migrations

import (
	"bytes"
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
	if err := db.DB.AutoMigrate(&models.User{}).Error; err != nil {
		log.Fatalf("error migrating the User model: %v", err)
	}
	// 2019/10/16 11:52 - Add the Agent and Job models
	if err := db.DB.AutoMigrate(&models.Agent{}).Error; err != nil {
		log.Fatalf("error migrating the Agent model: %v", err)
	}

	if err := db.DB.AutoMigrate(&models.Job{}).Error; err != nil {
		log.Fatalf("error migrating the Agent model: %v", err)
	}

	// Create the admin user if it doesn't exist
	u := models.User{Username: "admin"}
	if err := u.Load(); err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("error creating the admin user: %v", err)
		}

		fmt.Print("Please, set the admin password: ")
		bPwd, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		if err != nil {
			log.Fatalf("error creating the admin user: error reading the password: %v", err)
		}

		fmt.Print("Please, repeat admin password: ")
		bPwd2, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		if err != nil {
			log.Fatalf("error creating the admin user: error reading the password: %v", err)
		}

		if !bytes.Equal(bPwd, bPwd2) {
			log.Fatalf("error creating the admin user: passwords don't match")
		}

		pwd := strings.TrimSpace(string(bPwd))

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
