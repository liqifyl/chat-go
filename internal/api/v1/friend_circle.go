package v1

import "github.com/gin-gonic/gin"

//注册对出输出Api
func RegisterFriendCircleApi(gin *gin.Engine) {
}

//查看用户自己以朋友发布的朋友圈，按时间排序
func getFriendsCircleByUid(id int64, maxDateTime string, limit int) {
	
}

//用户发布朋友圈
func publishFriendCircle() {
}
