package database

import (
	"context"
	"time"

	"git.blauwelle.com/go/crate/log"
	"github.com/redis/go-redis/v9"

	"git.blauwelle.com/go/crate/cmd/sso/config"
)

func NewRedis(cfg config.Config) (*redis.Client, error) {
	log.Info(context.TODO(), "New Redis Client...")
	db := redis.NewClient(&redis.Options{
		Addr:            cfg.Redis.DSN,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		ConnMaxIdleTime: time.Duration(cfg.Redis.ConnMaxIdleTime) * 60000,
		ConnMaxLifetime: time.Duration(cfg.Redis.ConnMaxLifetime) * 60000,
	})
	return db, nil
}
