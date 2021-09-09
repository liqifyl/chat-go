package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "github.com/liqifyl/chat-go/internal/api/v1"
	"github.com/liqifyl/chat-go/internal/config"
)

func StartGinServer(config config.GinServerConfig) {
	r := gin.Default()
	userV1Api := v1.NewUserV1API(config)
	userV1Api.RegisterUserRestfulAPI(r)
	friendV1Api := v1.NewFriendV1API(config)
	friendV1Api.RegisterFriendApi(r)
	listenAddr := fmt.Sprintf("%s:%s", config.HostName, config.Port)
	r.Run(listenAddr)
}
