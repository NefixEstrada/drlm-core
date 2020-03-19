// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/context"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// Init checks whether the tokens secret is set and secure or not. It also creates the admin user if it doesn't exist
func Init(ctx *context.Context) {
	if ctx.Cfg.Security.TokensSecret == "" {
		log.Fatal(`you need to set a tokens secret in the configuration. You can generate one with "< /dev/urandom tr -dc 'A-Za-z0-9!#%&()*+,-./:;<=>?@[\\]^_{|}~' | head -c${1:-64};echo;"`)
	}

	if len(ctx.Cfg.Security.TokensSecret) < 32 {
		log.Fatal("the tokens secret needs to be at least 32 characters long")
	}

	// Create the admin user if it doesn't exist
	u := models.User{Username: "admin"}
	if err := u.Load(ctx); err != nil {
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

		if err := u.Add(ctx); err != nil {
			if models.IsErrUsrPwdStrength(err) {
				log.Fatalf(err.Error())
			}
		}
	}
}
