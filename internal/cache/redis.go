package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
)

const (
	cacheRedisError = iota + 10000
)

var (
	cacheRedisCtx                               = context.Background()
	cacheRedisClient              *redis.Client = nil
	cacheRedisClientLock                        = sync.Mutex{}
	cacheRedisClientMap                         = make(map[string]*redisClient)
	CacheDefaultRedisClientConfig               = RedisClientConfig{}
)

type RedisClientConfig struct {
	Addr string
	Pwd  string
	Db   int
}

type redisClient struct {
	config RedisClientConfig
	client *redis.Client
}

func convertConfigToStr(config RedisClientConfig) string {
	str := fmt.Sprintf("%s-%s-%d", config.Addr, config.Pwd, config.Db)
	return str
}

func getRedisClient() *redis.Client {
	return getRedisClientByConfig(CacheDefaultRedisClientConfig)
}

func getRedisClientByConfig(config RedisClientConfig) *redis.Client {
	cacheRedisClientLock.Lock()
	defer cacheRedisClientLock.Unlock()
	key := convertConfigToStr(config)
	client := cacheRedisClientMap[key]
	if client == nil {
		client := &redisClient{}
		client.config = config
		client.client = redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Pwd, // no password set
			DB:       config.Db,  // use default DB
		})
	}
	return client.client
}
