package auth

import (
	"fmt"
	"time"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/models"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// LoginLocal authenticates the user against the DB
func LoginLocal(usr, pwd string) (Token, time.Time, error) {
	u := models.User{Username: usr}
	if err := u.Load(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", time.Time{}, err
		}

		return "", time.Time{}, fmt.Errorf("error loading the user from the DB: %v", err)
	}

	if u.AuthType != types.Local {
		return "", time.Time{}, fmt.Errorf("invalid authentication method: user authentication type is %s", u.AuthType.String())
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pwd))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return "", time.Time{}, err
		}

		return "", time.Time{}, fmt.Errorf("password error: %v", err)
	}

	tkn, expiresAt, err := NewToken(usr)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error generating the login token: %v", err)
	}

	return tkn, expiresAt, nil
}
