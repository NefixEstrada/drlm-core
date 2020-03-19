// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"fmt"

	"github.com/brainupdaters/drlm-core/context"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	// Use MYSQL as the Gorm dialect
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Init creates the DB connection and does the migrations
func Init(ctx *context.Context) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		ctx.Cfg.DB.Usr,
		ctx.Cfg.DB.Pwd,
		ctx.Cfg.DB.Host,
		ctx.Cfg.DB.Port,
		ctx.Cfg.DB.DB,
	)

	var err error
	ctx.DB, err = gorm.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("error connecting to the DB: %v", err)
	}
	ctx.DB.LogMode(false)

	log.Info("successfully connected to the DB")
}
