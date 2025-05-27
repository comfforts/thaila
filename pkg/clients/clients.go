package clients

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Publisher interface {
	Publish(ctx context.Context, ch, msg string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, ch string) (<-chan *redis.Message, error)
}

type PubSuber interface {
	Publisher
	Subscriber
}

type Cacher interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
