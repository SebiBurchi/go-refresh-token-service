package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const KeepTTL = redis.KeepTTL

type RedisProxy interface {
	SetObject(ctx context.Context, key string, object interface{}, expiration time.Duration) error
	GetObject(ctx context.Context, key string) ([]byte, error)
	DeleteObject(ctx context.Context, key string) error
	Close() error
}

type RedisWrapper struct {
	client *redis.Client
}

func NewRedisClient(host, password string, port uint32) (proxy RedisProxy) {
	addr := fmt.Sprintf("%s:%d", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	proxy = &RedisWrapper{
		client: client,
	}

	return proxy
}

func (proxy *RedisWrapper) SetObject(ctx context.Context, key string, object interface{}, expiration time.Duration) error {
	err := proxy.client.Set(ctx, key, object, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (proxy *RedisWrapper) GetObject(ctx context.Context, key string) ([]byte, error) {
	value, err := proxy.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return []byte(value), nil
}

func (proxy *RedisWrapper) DeleteObject(ctx context.Context, key string) error {
	count, err := proxy.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("no entry found")
	}

	return err
}

func (proxy *RedisWrapper) Close() error {
	return proxy.client.Close()
}
