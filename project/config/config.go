package config

import (
	"os"
	"sync"
)

type Config struct {
	GatewayAddr string
	RedisAddr   string
	PostgresUrl string
}

var instance *Config
var once sync.Once

func Get() *Config {
	once.Do(func() {
		instance = &Config{
			GatewayAddr: os.Getenv("GATEWAY_ADDR"),
			RedisAddr:   os.Getenv("REDIS_ADDR"),
			PostgresUrl: os.Getenv("POSTGRES_URL"),
		}
	})

	return instance
}
