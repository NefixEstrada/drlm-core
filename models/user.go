package models

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/brainupdaters/drlm-core/db"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User is an individual user of DRLM Core
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
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

// BeforeSave is a GORM hook that gets executed before saving the user. It's used to encrypt the password
func (u *User) BeforeSave() error {
	if err := u.checkPwdStrength(); err != nil {
		return err
	}

	b, err := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
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

// errUsrPwdNoSpecialChar indicates that the password hasn't the required special character
var errUsrPwdNoSpecialChar = errors.New("the password requires, at least, an special character")

// IsErrUsrPwdStrength checks if an error is a password strength error
func IsErrUsrPwdStrength(err error) bool {
	return err == errUsrPwdLength || err == errUsrPwdNoUpperChar || err == errUsrPwdNoNumber || err == errUsrPwdNoSpecialChar
}

// checkPwdStrength validates that the password is strong enough
func (u *User) checkPwdStrength() error {
	// Has, at least, 8 characters long
	if len(u.Password) < 8 {
		return errUsrPwdLength
	}

	hasCapitalLetter := false
	hasNumber := false
	hasSpecialCharacter := false
	for _, l := range u.Password {
		// Has, at least, one capital letter
		if !hasCapitalLetter && unicode.IsUpper(l) {
			hasCapitalLetter = true
		} else if _, err := strconv.Atoi(string(l)); !hasNumber && err == nil {
			// Has, at least, one number
			hasNumber = true

		} else if !hasSpecialCharacter && unicode.IsSymbol(l) {
			// Has, at least one special character
			hasSpecialCharacter = true
		}
	}

	if !hasCapitalLetter {
		return errUsrPwdNoUpperChar
	}
	if !hasNumber {
		return errUsrPwdNoNumber
	}
	if !hasSpecialCharacter {
		return errUsrPwdNoSpecialChar
	}

	return nil
}
