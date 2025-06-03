package thaila_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/comfforts/thaila"
	"github.com/comfforts/thaila/pkg/clients"
	"github.com/comfforts/thaila/pkg/modules"
)

func TestThailaRedisSetGetDelete(t *testing.T) {
	err := godotenv.Load("./env/test.env")
	if err != nil {
		t.Fatalf("error loading environment: %v", err)
	}

	t.Parallel()

	thailaCfg := getRedisTestClientConfig()
	cache, err := thaila.NewThaila(thaila.ThailaRedis, thailaCfg)
	if err != nil {
		t.Fatalf("failed to create Thaila cache: %v", err)
	}
	defer cache.Close()
	ctx := t.Context()

	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	testKey := "test-key"

	err = cache.Set(ctx, testKey, "test-value", time.Second*5)
	if err != nil {
		t.Fatalf("error setting thaila key-value pair: %v ", err)
	}
	value, err := cache.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("error getting thaila key-value pair: %v ", err)
	}
	t.Logf("thaila key - %s, value - %s\n", testKey, value)

	err = cache.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("error deleting thaila key-value pair: %v ", err)
	}
	value, err = cache.Get(ctx, testKey)
	if err == nil {
		t.Fatalf("Unexpectedly found key - %s with value - %s after deletion\n", testKey, value)
	}
	t.Log("thaila redis set get delete operations test completed successfully")
}

func TestThailaLRUSetGetDelete(t *testing.T) {
	t.Parallel()

	thailaCfg := modules.NewLRUConfig(5, 0)
	cache, err := thaila.NewThaila(thaila.ThailaLRU, thailaCfg)
	if err != nil {
		t.Fatalf("failed to create Thaila cache: %v", err)
	}
	defer cache.Clear()

	ctx := t.Context()
	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	if err := cache.Set(ctx, "key1", "value1", 0); err != nil {
		t.Fatalf("failed to set value in thaila cache: %v ", err)
	}
	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("failed to get value from thaila cache: %v ", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
	if err := cache.Delete(ctx, "key1"); err != nil {
		t.Fatalf("failed to delete key from thaila cache: %v ", err)
	}
	value, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Fatalf("expected error when getting deleted key, got value: %s ", value)
	}

	t.Log("thaila LRU get set delete operations test completed successfully")
}

func getRedisTestClientConfig() clients.RedisConfig {
	redisPort := os.Getenv("REDIS_PORT")
	redisHost := os.Getenv("REDIS_HOST")
	redisPass := os.Getenv("REDIS_PASSWORD")

	return clients.NewRedisConfig(redisHost, redisPass, redisPort)
}
