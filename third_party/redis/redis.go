// TODO: Refactor this peace of crap! @allanger
package redis

import (
	"os"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

// Client retur nredis client
func Client() *redis.Client {
	if client == nil {
		NewClient()
	}
	return client
}

// NewClient create redis client
func NewClient() error {
	redisAddr, exists := os.LookupEnv("REDIS_HOST")
	if !exists {
		redisAddr = "localhost:6379"
	}
	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return nil
}
