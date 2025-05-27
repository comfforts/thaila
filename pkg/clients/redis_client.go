package clients

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	// "mercury/pkg/log"
)

const (
	ERR_REDIS_SUBSCRIBE = "error subscribing to redis PubSub"
	ERR_REDIS_HOST      = "error missing redis host"
	ERR_REDIS_PASS      = "error missing redis pass"
)

var (
	ErrRedisSubscribe = errors.New(ERR_REDIS_SUBSCRIBE)
	ErrRedisPass      = errors.New(ERR_REDIS_PASS)
	ErrRedisHost      = errors.New(ERR_REDIS_HOST)
)

const MESSAGE_CHANNEL = "test-message-ch"

type redisClient struct {
	client *redis.Client
}

func NewRedisClient() (*redisClient, error) {
	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		return nil, ErrRedisPass
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisPass == "" {
		return nil, ErrRedisHost
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	// log.Debugf(nil, "redis client host: %s, port: %s", redisHost, redisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPass,
	})

	return &redisClient{
		client: rdb,
	}, nil
}

func (rc *redisClient) Subscribe(ctx context.Context, ch string) (<-chan *redis.Message, error) {
	ps := rc.client.Subscribe(ctx, ch)
	if val, err := ps.Receive(ctx); err != nil {
		// log.Errorf(nil, "redis subscribe error: %v", err)
		return nil, ErrRedisSubscribe
	} else {
		switch val.(type) {
		case *redis.Subscription:
			// subscribe succeeded
			// log.Info(nil, "redis subscribe succeeded")
		case *redis.Message:
			// received first message
			// log.Info(nil, "redis received first message")
		case *redis.Pong:
			// pong received
			// log.Info(nil, "redis pong received")
		default:
			// handle error
			// log.Error(nil, "redis unknown response")
			return nil, ErrRedisSubscribe
		}
	}

	return ps.Channel(), nil
}

func (rc *redisClient) Publish(ctx context.Context, ch, msg string) {
	res := rc.client.Publish(ctx, ch, msg)
	_, err := res.Result()
	if err != nil {
		// log.Errorf(nil, "redis publish error: %v", err)
	}
}

func (rc *redisClient) Close() {
	// log.Info(nil, "closing redis connection")
	err := rc.client.Close()
	if err != nil {
		// log.Errorf(nil, "error closing redis connection: %v", err)
	}
}
