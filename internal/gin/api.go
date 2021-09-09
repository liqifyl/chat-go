package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "github.com/liqifyl/chat-go/internal/api/v1"
	"github.com/liqifyl/chat-go/internal/cache"
	"github.com/liqifyl/chat-go/internal/config"
	"github.com/liqifyl/chat-go/internal/sql"
)

func StartGinServer(config config.GinServerConfig) {
	r := gin.Default()
	cache.CacheDefaultRedisClientConfig.Addr = config.RedisServerAddress
	cache.CacheDefaultRedisClientConfig.Pwd = config.RedisServerPwd
	cache.CacheDefaultRedisClientConfig.Db = config.RedisSelectDB
	sql.DefaultDbConfig.DriveName = "mysql"
	sql.DefaultDbConfig.DataSourceName = config.MysqlChatDataSourceName
	userV1Api := v1.NewUserV1API(config)
	userV1Api.RegisterUserRestfulAPI(r)
	friendV1Api := v1.NewFriendV1API(config)
	friendV1Api.RegisterFriendApi(r)
	listenAddr := fmt.Sprintf("%s:%s", config.HostName, config.Port)
	r.Run(listenAddr)
}
