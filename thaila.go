package thaila

import (
	"context"
	"errors"
	"time"

	"github.com/comfforts/thaila/pkg/clients"
	"github.com/comfforts/thaila/pkg/modules"
	"github.com/redis/go-redis/v9"
)

type ThailaStrategy string

const (
	ThailaRedis ThailaStrategy = "REDIS_THAILA"
	ThailaLRU   ThailaStrategy = "LRU_THAILA"
)

const (
	ERR_INVALID_STRATEGY  = "error invalid caching strategy"
	ERR_INVALID_REDIS_CFG = "error invalid redis config"
	ERR_INVALID_LRU_CFG   = "error invalid LRU config"
)

var (
	ErrInvalidStrategy = errors.New(ERR_INVALID_STRATEGY)
	ErrInvalidRedisCfg = errors.New(ERR_INVALID_REDIS_CFG)
	ErrInvalidLRUCfg   = errors.New(ERR_INVALID_LRU_CFG)
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

type Thaila[C ThailaConfig] interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
	Clear()
}

type ThailaConfig interface {
	clients.RedisConfig | modules.LRUConfig
}

func NewThaila[C ThailaConfig](stgy ThailaStrategy, cfg C) (Thaila[C], error) {
	if stgy != ThailaRedis && stgy != ThailaLRU {
		return nil, ErrInvalidStrategy
	}

	if stgy == ThailaRedis {
		redisCfg, ok := any(cfg).(clients.RedisConfig)
		if !ok {
			return nil, ErrInvalidRedisCfg
		}

		return clients.NewRedisClient(redisCfg)
	} else if stgy == ThailaLRU {
		lruCfg, ok := any(cfg).(modules.LRUConfig)
		if !ok {
			return nil, ErrInvalidRedisCfg
		}

		return modules.NewLRUCache[string, any](lruCfg)
	}
	return nil, ErrInvalidStrategy
}
