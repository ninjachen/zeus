package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Addr               string `default:"127.0.0.1:6379" validate:"required"`
	DB                 int
	Username           string
	Password           string
	MaxRetries         int
	MinRetryBackoff    time.Duration
	MaxRetryBackoff    time.Duration
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
}

//创建redis连接
func NewRedis(conf *Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	err := client.Ping(context.Background()).Err()
	return client, err
}
