package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// Ticket的获取和储存操作
var (
	ErrTicketNotExists = errors.New("ticket not exists")
)

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

//type Store interface {
//	GetTicket(ctx context.Context, ticket string) (UserInfo, error)
//	SetTicket(ctx context.Context, ticket string, info UserInfo) error
//}

func SetTicketToRedis(ctx context.Context, ticket string, r *redis.Client, info UserInfo) error {
	jsoninfo, _ := json.Marshal(info)
	data := string(jsoninfo)
	ttl := getTTL()
	result := r.Set(ctx, ticket, data, time.Duration(ttl)*time.Second)
	if result.Err() != nil {
		fmt.Println("redis set fault")
		return nil
	}
	fmt.Println("redis set success")
	return nil
}
func GetTicketFromRedis(ctx context.Context, ticket string, r *redis.Client) (UserInfo, error) {

	existe, _ := r.Exists(ctx, ticket).Result()
	var info UserInfo
	if existe == 1 {
		jsoninfo, err := r.Get(ctx, ticket).Result()
		if err != nil {
			return UserInfo{}, ErrTicketNotExists
		}
		err = json.Unmarshal([]byte(jsoninfo), &info)
		if err != nil {
			return UserInfo{}, ErrTicketNotExists
		}
		return info, nil
	}

	return UserInfo{}, ErrTicketNotExists
}

type TTL struct {
	Redis struct {
		TTL int `yaml:"ttl"`
	} `yaml:"redis"`
}

func getTTL() int {
	b, err := os.ReadFile("config/config.yaml")
	if err != nil {
		panic(err)
	}
	var cfg TTL
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}
	return cfg.Redis.TTL
}
