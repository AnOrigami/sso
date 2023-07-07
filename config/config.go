package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ListenConfig struct {
	Port int `yaml:"port"`
}

type MysqlConfig struct {
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	ConnMaxIdleTime int    `yaml:"connMaxIdleTime"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime"`
}
type RedisConfig struct {
	DSN             string `yaml:"dsn"`
	DB              int    `yaml:"db"`
	PoolSize        int    `yaml:"poolSize"`
	MinIdleConns    int    `yaml:"minIdleConns"`
	ConnMaxIdleTime int    `yaml:"connMaxIdleTime"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	TTL             int    `yaml:"ttl"`
}

type Config struct {
	Listen ListenConfig `yaml:"listen"`
	Mysql  MysqlConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
}

func GetConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
