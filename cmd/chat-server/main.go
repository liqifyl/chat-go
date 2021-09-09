package main

import (
	"github.com/liqifyl/chat-go/internal/config"
	"github.com/liqifyl/chat-go/internal/gin"
	"go.uber.org/zap"
	"os"
)

func main() {
	ServerListenAddress.Host = "127.0.0.1:9092"
	LogToGraylogAddress.Host = "47.107.231.119:22000"
	RedisServerAddress.Host = "localhost:6379"
	MysqlChatDataSourceName = "root:liqifyl10051113@tcp(127.0.0.1:3306)/im"
	ginConfig := config.GinServerConfig{}
	ginConfig.TestUid = TestUid
	ginConfig.UserImageSaveDir = UserImageSaveDir
	ginConfig.HostName = ServerListenAddress.Hostname()
	ginConfig.Port = ServerListenAddress.Port()
	InitLog()
	if ginConfig.HostName == "" || ginConfig.Port == "" {
		zap.L().Error("hostName or port is empty")
		os.Exit(-1)
	}
	zap.L().Debug("starting")
	gin.StartGinServer(ginConfig)
	zap.L().Debug("exited")
	TermLog()
}
