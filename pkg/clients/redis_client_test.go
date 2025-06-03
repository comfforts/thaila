package clients_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/comfforts/thaila/pkg/clients"
)

const MESSAGE_CHANNEL = "test-message-ch"

func TestRedisClientPubSub(t *testing.T) {
	err := godotenv.Load("../../env/test.env")
	if err != nil {
		t.Fatalf("error loading environment: %v", err)
	}

	t.Parallel()

	// Initialize the Redis client
	cfg := getRedisTestClientConfig()
	rcl, err := clients.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("error initializing redis client: %v", err)
	}
	defer rcl.Close()

	ctx := t.Context()
	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	ch, err := rcl.Subscribe(ctx, MESSAGE_CHANNEL)
	if err != nil {
		t.Fatalf("error subscribing to redis channel: %v ", err)
	}

	MSG_CNT := 5
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					break
				}
				t.Log("Received message from channel:", msg.Payload)
				count++
				if count >= MSG_CNT {
					t.Logf("Received %d messages from channel, returning", MSG_CNT)
					wg.Done()
					return
				}
			}
		}
	}()

	wg.Add(1)
	i := 1
	for i <= MSG_CNT {
		go rcl.Publish(ctx, MESSAGE_CHANNEL, fmt.Sprintf("message payload %d", i))
		i++
	}
	wg.Done()

	wg.Wait()
	t.Log("Redis pub sub message processing done")
}

func TestRedisClientSetGetDelete(t *testing.T) {
	err := godotenv.Load("../../env/test.env")
	if err != nil {
		t.Fatalf("error loading environment: %v", err)
	}

	t.Parallel()

	// Initialize the Redis client
	cfg := getRedisTestClientConfig()
	rcl, err := clients.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("error initializing redis client: %v ", err)
	}
	defer rcl.Close()

	ctx := t.Context()
	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	testKey := "test-key"

	err = rcl.Set(ctx, testKey, "test-value", time.Second*5)
	if err != nil {
		t.Fatalf("error setting redis key-value pair: %v ", err)
	}
	value, err := rcl.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("error getting redis key-value pair: %v ", err)
	}
	t.Logf("Redis key - %s, value - %s", testKey, value)

	err = rcl.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("error deleting redis key-value pair: %v ", err)
	}
	if value, err = rcl.Get(ctx, testKey); err == nil {
		t.Fatalf("Unexpectedly found key - %s with value - %s after deletion", testKey, value)
	}
	t.Log("Redis set get delete operations test completed successfully")
}

func getRedisTestClientConfig() clients.RedisConfig {
	redisPort := os.Getenv("REDIS_PORT")
	redisHost := os.Getenv("REDIS_HOST")
	redisPass := os.Getenv("REDIS_PASSWORD")

	return clients.NewRedisConfig(redisHost, redisPass, redisPort)
}
