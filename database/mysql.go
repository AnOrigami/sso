package database

import (
	"context"
	"time"

	"git.blauwelle.com/go/crate/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"git.blauwelle.com/go/crate/cmd/sso/config"
)

func NewMysql(cfg config.Config) (*gorm.DB, error) {
	log.Info(context.TODO(), "New mysqlDB...")

	db, err := gorm.Open(
		mysql.Open(cfg.Mysql.DSN),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 配置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Mysql.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Mysql.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Mysql.ConnMaxIdleTime) * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Mysql.ConnMaxLifetime) * time.Minute)

	return db, nil
}
