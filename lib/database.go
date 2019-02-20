package lib

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

type DatabaseConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Database string
}

var DBConn *gorm.DB

func InitDatabase(cfg DatabaseConfig) {

	if DBConn != nil {
		return
	}

	connectionString := (cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Server + ":" + cfg.Port + ")/" + cfg.Database)

	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		log.Panic("failed to connect database")
	}
	log.Info("Connected to database!")

	DBConn = db

	InitUser()
}

func closeDatabase() {
	DBConn.Close()
}
