package example_commands_test

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func ExampleClient_connect_basic() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis-10751.c334.asia-southeast2-1.gce.redns.redis-cloud.com:10751",
		Username: "default",
		Password: "LpE41yUZ5w0BEo6kaEHmzIH5gdjD9uI5",
		DB:       0,
	})

	rdb.Set(ctx, "foo", "bar", 0)
	result, err := rdb.Get(ctx, "foo").Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(result) // >>> bar

}

