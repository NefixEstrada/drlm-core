package models

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/brainupdaters/drlm-core/auth/types"
	"github.com/brainupdaters/drlm-core/cfg"
	"github.com/brainupdaters/drlm-core/db"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User is an individual user of DRLM Core
type User struct {
	gorm.Model
	Username string     `gorm:"unique;not null"`
	Password string     `gorm:"not null"`
	AuthType types.Type `gorm:"not null"`
}

// UserList returns a list with all the users
func UserList() ([]*User, error) {
	users := []*User{}

	if err := db.DB.Select("created_at, updated_at, username, auth_type").Find(&users).Error; err != nil {
		return []*User{}, fmt.Errorf("error getting the list of users: %v", err)
	}

	return users, nil
}

// Add creates a new user in the DB
func (u *User) Add() error {
	if err := db.DB.Create(u).Error; err != nil {
		if IsErrUsrPwdStrength(err) {
			return err
		}

		return fmt.Errorf("error adding the user to the DB: %v", err)
	}

	return nil
}

// Load loads the user from the DB using the username
func (u *User) Load() error {
	if err := db.DB.Where("username = ?", u.Username).First(u).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return err
		}

		return fmt.Errorf("error loading the user from the DB: %v", err)
	}

	return nil
}

// Delete removes an user from the DB using the username
func (u *User) Delete() error {
	if err := u.Load(); err != nil {
		return err
	}

	return db.DB.Delete(u).Error
}

// BeforeSave is a GORM hook that gets executed before saving the user. It's used to encrypt the password
func (u *User) BeforeSave() error {
	if err := u.checkPwdStrength(); err != nil {
		return err
	}

	b, err := bcrypt.GenerateFromPassword([]byte(u.Password), cfg.Config.Security.BcryptCost)
	if err != nil {
		return fmt.Errorf("error encrypting the user password: %v", err)
	}

	u.Password = string(b)
	return nil
}

// errUsrPwdLength indicates that the password is too short
var errUsrPwdLength = errors.New("the password requires, at least, a length of 8 characters")

// errUsrPwdNoUpperChar indicates that the password hasn't the required uppercase character
var errUsrPwdNoUpperChar = errors.New("the password requires, at least, an uppercase character")

// errUsrPwdNoNumber indicates that the password hasn't the requied number
var errUsrPwdNoNumber = errors.New("the password requires, at least, a number")

// IsErrUsrPwdStrength checks if an error is a password strength error
func IsErrUsrPwdStrength(err error) bool {
	return err == errUsrPwdLength || err == errUsrPwdNoUpperChar || err == errUsrPwdNoNumber
}

// checkPwdStrength validates that the password is strong enough
func (u *User) checkPwdStrength() error {
	// Has, at least, 8 characters long
	if len(u.Password) < 8 {
		return errUsrPwdLength
	}

	hasCapitalLetter := false
	hasNumber := false
	for _, l := range u.Password {
		// Has, at least, one capital letter
		if !hasCapitalLetter && unicode.IsUpper(l) {
			hasCapitalLetter = true
		} else if _, err := strconv.Atoi(string(l)); !hasNumber && err == nil {
			// Has, at least, one number
			hasNumber = true

		}
	}

	if !hasCapitalLetter {
		return errUsrPwdNoUpperChar
	}
	if !hasNumber {
		return errUsrPwdNoNumber
	}

	return nil
}
