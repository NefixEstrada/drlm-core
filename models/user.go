package models

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/db"

	"github.com/jinzhu/gorm"
)

// User is an individual user of DRLM Core
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"unique;not null"`
}

// Add creates a new user in the DB
func (u *User) Add() error {
	if err := db.DB.Create(u).Error; err != nil {
		return fmt.Errorf("error adding the user to the DB: %v", err)
	}

	return nil
}
