package db

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/cfg"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	// Use MYSQL as the Gorm dialect
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB is the connection with the DB
var DB *gorm.DB

// Init creates the DB connection and does the migrations
func Init() {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		cfg.Config.DB.Usr,
		cfg.Config.DB.Pwd,
		cfg.Config.DB.Host,
		cfg.Config.DB.Port,
		cfg.Config.DB.DB,
	)

	var err error
	DB, err = gorm.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("error connecting to the DB: %v", err)
	}

	log.Info("successfully connected to the DB")
}
