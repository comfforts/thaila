package clients_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/comfforts/thaila/pkg/clients"
)

func TestRedisClientPubSub(t *testing.T) {
	t.Parallel()

	// Initialize the Redis client
	rcl, err := clients.NewRedisClient()
	if err != nil {
		fmt.Println("error initializing redis client")
		return
	}
	defer rcl.Close()

	ctx := t.Context()
	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	ch, err := rcl.Subscribe(ctx, clients.MESSAGE_CHANNEL)
	if err != nil {
		fmt.Println("error subscribing to redis channel")
		return
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
				fmt.Println("Received message from channel:", msg.Payload)
				count++
				if count >= MSG_CNT {
					fmt.Printf("Received %d messages from channel, returning\n", MSG_CNT)
					wg.Done()
					return
				}
			}
		}
	}()

	wg.Add(1)
	i := 1
	for i <= MSG_CNT {
		go rcl.Publish(ctx, clients.MESSAGE_CHANNEL, fmt.Sprintf("message payload %d", i))
		i++
	}
	wg.Done()

	wg.Wait()
	fmt.Println("message processing done")
}

func TestRedisClientSetGetDelete(t *testing.T) {
	t.Parallel()

	// Initialize the Redis client
	rcl, err := clients.NewRedisClient()
	if err != nil {
		fmt.Println("error initializing redis client")
		return
	}
	defer rcl.Close()

	ctx := t.Context()
	d := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(ctx, d)
	defer cancel()

	testKey := "test-key"

	err = rcl.Set(ctx, testKey, "test-value", time.Second*5)
	if err != nil {
		fmt.Println("error setting redis key-value pair - ", err)
		return
	}
	value, err := rcl.Get(ctx, testKey)
	if err != nil {
		fmt.Println("error getting redis key-value pair - ", err)
		return
	}
	fmt.Printf("Redis key - %s, value - %s\n", testKey, value)

	err = rcl.Delete(ctx, testKey)
	if err != nil {
		fmt.Println("error deleting redis key-value pair - ", err)
		return
	}
	value, err = rcl.Get(ctx, testKey)
	if err != nil {
		fmt.Println("error getting redis key-value pair after deletion - ", err)
	} else {
		fmt.Printf("Unexpectedly found key - %s with value - %s after deletion\n", testKey, value)
	}
	fmt.Println("Redis key-value operations test completed successfully")
}
