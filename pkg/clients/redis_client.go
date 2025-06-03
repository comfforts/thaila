package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ERR_REDIS_SUBSCRIBE = "error subscribing to redis PubSub"
	ERR_REDIS_HOST      = "error missing redis host"
	ERR_REDIS_PASS      = "error missing redis pass"
	ERR_REDIS_PORT      = "error missing redis port"
	ERR_NOT_FOUND       = "error not found"
	ERR_REDIS_CLOSE     = "error closing redis"
	ERR_REDIS_PUBLISH   = "error publishing to redis channel"
)

var (
	ErrRedisPass      = errors.New(ERR_REDIS_PASS)
	ErrRedisHost      = errors.New(ERR_REDIS_HOST)
	ErrRedisPort      = errors.New(ERR_REDIS_PORT)
	ErrRedisClose     = errors.New(ERR_REDIS_CLOSE)
	ErrRedisSubscribe = errors.New(ERR_REDIS_SUBSCRIBE)
	ErrRedisPublish   = errors.New(ERR_REDIS_PUBLISH)
	ErrNotFound       = errors.New(ERR_NOT_FOUND)
)

type redisClient struct {
	client *redis.Client
}

// RedisConfig holds configuration for connecting to a Redis server.
type RedisConfig struct {
	Host     string // Redis server host
	Password string // Redis server password
	Port     string // Redis server port
}

// NewRedisConfig returns a new RedisConfig with the given host, password, and port.
func NewRedisConfig(host, pass, port string) RedisConfig {
	return RedisConfig{
		Host:     host,
		Password: pass,
		Port:     port,
	}
}

// NewRedisClient creates a new Redis client using the provided configuration.
// Returns an error if any required configuration is missing.
func NewRedisClient(cfg RedisConfig) (*redisClient, error) {
	if cfg.Password == "" {
		return nil, ErrRedisPass
	}

	if cfg.Host == "" {
		return nil, ErrRedisHost
	}

	if cfg.Port == "" {
		return nil, ErrRedisPort
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
	})

	// TODO setup sentinel support if needed
	// rdb := redis.NewFailoverClient(&redis.FailoverOptions{
	// 	MasterName:    redisMasterName,
	// 	SentinelAddrs: []string{fmt.Sprintf("%s:%s", redisHost, redisSentinelPort)},
	// 	Password:      redisPass,
	// })

	return &redisClient{
		client: rdb,
	}, nil
}

// Get retrieves the value for the given key from Redis.
// Returns ErrNotFound if the key does not exist.
func (rc *redisClient) Get(ctx context.Context, key string) (any, error) {
	result, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

// Set sets the value for the given key in Redis with the specified TTL.
func (rc *redisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := rc.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

// Delete removes the given key from Redis.
func (rc *redisClient) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

// Subscribe subscribes to the specified Redis channel and returns a channel for receiving messages.
// Returns an error if the subscription fails.
func (rc *redisClient) Subscribe(ctx context.Context, ch string) (<-chan *redis.Message, error) {
	ps := rc.client.Subscribe(ctx, ch)
	if val, err := ps.Receive(ctx); err != nil {
		return nil, ErrRedisSubscribe
	} else {
		switch val.(type) {
		case *redis.Subscription:
			fmt.Println("redis subscribe succeeded")
		case *redis.Message:
			fmt.Println("redis received first message")
		case *redis.Pong:
			fmt.Println("redis pong received")
		default:
			return nil, ErrRedisSubscribe
		}
	}

	return ps.Channel(), nil
}

// Publish sends a message to the specified Redis channel.
// Returns an error if the publish fails.
func (rc *redisClient) Publish(ctx context.Context, ch, msg string) error {
	res := rc.client.Publish(ctx, ch, msg)
	_, err := res.Result()
	if err != nil {
		return ErrRedisPublish
	}
	return nil
}

// Close closes the Redis client connection.
func (rc *redisClient) Close() error {
	err := rc.client.Close()
	if err != nil {
		return err
	}

	return nil
}

// Clear removes all entries from the cache.
func (rc *redisClient) Clear() {
	// Redis does not have a direct way to clear all keys in a single command.
	// You can use the FLUSHALL command to remove all keys from all databases.
	// However, this is dangerous in production environments.
	// Use with caution!
	ctx := context.Background()
	err := rc.client.FlushAll(ctx).Err()
	if err != nil {
		fmt.Println("error clearing Redis cache - ", err)
	}
	// Note: This is a destructive operation and should be used with caution
}
