package auth

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// LoginLocal authenticates the user against the DB
func LoginLocal(usr, pwd string) (Token, error) {
	u := models.User{Username: usr}
	if err := u.Load(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", err
		}

		return "", fmt.Errorf("error loading the user from the DB: %v", err)
	}

	if u.AuthType != types.Local {
		return "", fmt.Errorf("invalid authentication method: user authentication type is %s", u.AuthType.String())
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pwd))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return "", err
		}

		return "", fmt.Errorf("password error: %v", err)
	}

	tkn, err := NewToken(usr)
	if err != nil {
		return "", fmt.Errorf("error generating the login token: %v", err)
	}

	return tkn, nil
}
