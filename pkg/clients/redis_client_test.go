package clients_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/comfforts/thaila/pkg/clients"
)

func TestRedisClient(t *testing.T) {
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
