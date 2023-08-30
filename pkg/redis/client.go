package redis

import (
	"encoding/json"
	"fmt"
	"main/pkg/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Cache interface {
	Exists(ctx context.Context, keys ...any) (int, error)
	Exist(ctx context.Context, key any) (bool, error)
	Get(ctx context.Context, key any, val any) error
	Set(ctx context.Context, key any, val any) error
	Flush(ctx context.Context) error
	HSet(ctx context.Context, key any, value any) error
	HGet(ctx context.Context, key any, value any) error
}

type client struct {
	ctx    context.Context
	client *redis.Client
}

func NewRedis(ctx context.Context, config *Config) (Cache, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("got error for redis ping command: %v", err)
	}
	return &client{
		ctx:    ctx,
		client: c,
	}, nil
}

func (c *client) HGet(ctx context.Context, key any, value any) error {
	if err := c.client.HGetAll(ctx, redisKey(key)).Scan(value); err != nil {
		return fmt.Errorf("redis hget command error: %v", err)
	}
	return nil
}

func (c *client) HSet(ctx context.Context, key any, value any) error {
	if _, err := c.client.HSet(ctx, redisKey(key), value).Result(); err != nil {
		return fmt.Errorf("redis hset command error: %v", err)
	}
	return nil
}

func (c *client) Flush(ctx context.Context) error {
	if _, err := c.client.FlushDB(ctx).Result(); err != nil {
		return fmt.Errorf("redis flush command error: %v", err)
	}
	return nil
}

func (c *client) Exist(ctx context.Context, key any) (bool, error) {
	count, err := c.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (c *client) Exists(ctx context.Context, keys ...any) (int, error) {
	str := make([]string, len(keys))

	utils.ForEach(func(key any) {
		str = append(str, redisKey(key))
	}, keys...)

	count, err := c.client.Exists(ctx, str...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists command error: %v", err)
	}
	return int(count), nil
}

func (c *client) Get(ctx context.Context, key any, val any) error {
	buf, err := c.client.Get(ctx, redisKey(key)).Bytes()
	if err != nil {
		return fmt.Errorf("redis get command error: %v", err)
	}
	if err := json.Unmarshal(buf, val); err != nil {
		return fmt.Errorf("json unmarshal error: %v", err)
	}
	return nil
}

func (c *client) Set(ctx context.Context, key any, val any) error {
	buf, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}
	if err := c.client.Set(ctx, fmt.Sprint(key), buf, 0).Err(); err != nil {
		return fmt.Errorf("redis set key value error: %v", err)
	}
	return nil
}

func redisKey(key any) string {
	return fmt.Sprint(key)
}
