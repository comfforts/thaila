package modules_test

import (
	"fmt"
	"testing"

	"github.com/comfforts/thaila/pkg/modules"
	"github.com/stretchr/testify/require"
)

func TestLRUSetGetDelete(t *testing.T) {
	t.Parallel()
	cfg := modules.NewLRUConfig(5, 0)
	cache, err := modules.NewLRUCache[string, string](cfg)
	if err != nil {
		t.Fatalf("failed to create LRU cache: %v ", err)
	}
	defer cache.Clear()

	ctx := t.Context()
	if err := cache.Set(ctx, "key1", "value1", 0); err != nil {
		t.Fatalf("failed to set value in LRU cache: %v", err)
	}
	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("failed to get value from LRU cache: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
	if err := cache.Delete(ctx, "key1"); err != nil {
		t.Fatalf("failed to delete key from LRU cache: %v", err)
	}
	if value, err = cache.Get(ctx, "key1"); err == nil {
		t.Fatalf("expected error when getting deleted key, got value: %s", value)
	}

	t.Log("LRU get set delete operations test completed successfully")
}

func TestLRUCapacityAssignment(t *testing.T) {
	t.Parallel()
	cfg := modules.NewLRUConfig(3, 0)
	cache, err := modules.NewLRUCache[string, string](cfg)
	if err != nil {
		t.Fatalf("failed to create LRU cache: %v", err)
	}
	defer cache.Clear()

	ctx := t.Context()
	for i := 1; i <= 4; i++ {
		if err := cache.Set(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i), 0); err != nil {
			t.Fatalf("failed to set value in LRU cache - %v", err)
		}
	}

	if value, err := cache.Get(ctx, "key1"); err == nil {
		t.Fatalf("expected error when getting oldest key, got value: %s", value)
	} else {
		require.Equal(t, modules.ERR_NOT_FOUND, err.Error(), "Expected key1 to be evicted")
	}

	t.Log("LRU capacity assignment test completed successfully")
}

func TestLRULimitCheck(t *testing.T) {
	t.Parallel()
	cfg := modules.NewLRUConfig(3, 2)
	cache, err := modules.NewLRUCache[string, string](cfg)
	if err != nil {
		t.Fatalf("failed to create LRU cache: %v", err)
	}
	defer cache.Clear()

	ctx := t.Context()

	if err = cache.Set(ctx, "key1", "value1", 0); err == nil {
		t.Fatalf("failed to set limit constraint LRU cache: %v", err)
	} else {
		require.Equal(t, modules.ERR_VAL_TOO_LARGE, err.Error(), "Expected error for value exceeding limit")
	}

	t.Log("LRU limit check test completed successfully")
}
