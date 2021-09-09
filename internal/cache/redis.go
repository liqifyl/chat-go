package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"sync"
)

const (
	cacheRedisError = iota + 20
)

var (
	cacheRedisCtx                      = context.Background()
	cacheRedisClient     *redis.Client = nil
	cacheRedisClientLock               = sync.Mutex{}
)

func getRedisClient() *redis.Client {
	cacheRedisClientLock.Lock()
	defer cacheRedisClientLock.Unlock()
	if cacheRedisClient == nil {
		cacheRedisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	}
	return cacheRedisClient
}
