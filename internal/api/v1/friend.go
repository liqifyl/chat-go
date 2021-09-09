package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/liqifyl/chat-go/internal/cache"
	"github.com/liqifyl/chat-go/internal/config"
	"github.com/liqifyl/chat-go/internal/sql"
	"log"
	"net/http"
	"strconv"
)

const (
	friendV1UidInvalid = iota + 300
	friendV1FidInvalid
	friendV1UidAndFidSame
	friendV1ConvertUidFail
	friendV1QueryFriendsFail
	friendV1ExeAddFriendFail
	friendV1NewNickEmpty
	friendV1ExeUpdateFriendNickFail
	friendV1ExeDelFriendFail
)

const (
	friendHeaderUidKey = "uid"
)

type getFriendsResponse struct {
	Friends []*sql.Friend `json:"friends"`
}

type addFriendRequest struct {
	Uid int64 `json:"uid"`
	Fid int64 `json:"fid"`
}

type addFriendResponse struct {
	Id    int64  `json:"id"`
	Uid   int64  `json:"uid"`
	Fid   int64  `json:"fid"`
	Fnick string `json:"fnick"`
	Etime string `json:"etime"`
}

type updateFriendNickRequest struct {
	Id      int64  `json:"id"`
	Uid     int64  `json:"uid"`
	Fid     int64  `json:"fid"`
	NewNick string `json:"new_nick"`
}

type updateFriendNickResponse struct {
	updateFriendNickRequest
}

type deleteFriendRequest struct {
	Id  int64 `json:"id"`
	Uid int64 `json:"uid"`
	Fid int64 `json:"fid"`
}

type deleteFriendResponse struct {
	errResponse
}

type FriendV1API struct {
	Config config.GinServerConfig
}

func NewFriendV1API(config config.GinServerConfig) *FriendV1API {
	return &FriendV1API{Config: config}
}

//注册对外输出api
func (self *FriendV1API) RegisterFriendApi(gin *gin.Engine) {
	gin.POST("/v1/friend/add", self.addFriend)
	gin.POST("/v1/friend/update/nick", self.updateFriendNick)
	gin.POST("/v1/friend/delete", self.deleteFriend)
	gin.GET("/v1/friend/query/friends", self.getFriendsByUid)
}

//通过用户id获取通讯录
func (self *FriendV1API) getFriendsByUid(c *gin.Context) {
	logTag := "friend->get->friends->"
	uidStr := c.GetHeader(friendHeaderUidKey)
	if uidStr == "" {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1UidInvalid, "uid invalid"))
		return
	}
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Printf("%sstrconv uid error %v", logTag, err)
		c.JSON(http.StatusOK, fail(friendV1ConvertUidFail, "uid invalid"))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}
	friends, err := cache.GetFriendsByUid(int64(uid))
	if err != nil {
		log.Printf("%sget friends error from redis or db, %v", logTag, err)
		c.JSON(http.StatusOK, fail(friendV1QueryFriendsFail, err.Error()))
		return
	}
	response := getFriendsResponse{Friends: friends}
	c.JSON(http.StatusOK, response)
}

//添加好友
func (self *FriendV1API) addFriend(c *gin.Context) {
	logTag := "friend->add->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		code := HttpErrorContentTypeInvalid
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(code, msg))
		return
	}
	contentLenStr := c.GetHeader("Content-Length")
	if contentLenStr == "" {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenEmpty, "content length is empty"))
		return
	}
	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is empty"))
		return
	}
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len, contentLen:%d, bodyLen:%d", logTag, contentLen, len(body))
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &addFriendRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal addFriendRequest err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1UidInvalid, "uid invalid"))
		return
	}
	if request.Fid <= 0 {
		log.Printf("%sfid invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1FidInvalid, "fid invalid"))
		return
	}
	if request.Uid == request.Fid {
		log.Printf("%suid is equal fid", logTag)
		c.JSON(http.StatusOK, fail(friendV1UidAndFidSame, "uid is equal fid"))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}
	friend := &sql.Friend{Fid: request.Fid, Uid: request.Uid}
	id, err := cache.AddFriend(friend)
	if err != nil {
		log.Printf("%sexe add friend error %v", logTag, err)
		c.JSON(http.StatusOK, fail(friendV1ExeAddFriendFail, "exe add friend fail"))
		return
	}
	response := addFriendResponse{Id: id, Uid: friend.Uid, Fid: friend.Fid, Fnick: friend.Fnick}
	if friend.Etime != "" {
		response.Etime = friend.Etime
	}
	c.JSON(http.StatusOK, response)
}

//更新朋友的昵称
func (self *FriendV1API) updateFriendNick(c *gin.Context) {
	logTag := "friend->update->friend->nick->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		code := HttpErrorContentTypeInvalid
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(code, msg))
		return
	}
	contentLenStr := c.GetHeader("Content-Length")
	if contentLenStr == "" {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenEmpty, "content length is empty"))
		return
	}
	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is empty"))
		return
	}
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len, contentLen:%d, bodyLen:%d", logTag, contentLen, len(body))
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &updateFriendNickRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal updateFriendNickRequest err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid < 1 {
		log.Printf("%suid is invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1UidInvalid, "uid is invalid"))
		return
	}
	if request.Fid < 1 {
		log.Printf("%suid is invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1FidInvalid, "uid is invalid"))
		return
	}
	if request.NewNick == "" {
		log.Printf("%snew nick is empty", logTag)
		c.JSON(http.StatusOK, fail(friendV1NewNickEmpty, "new nick is empty"))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}
	friend := sql.Friend{Id: request.Id, Fid: request.Fid, Uid: request.Uid, Fnick: request.NewNick}
	err = cache.UpdateFriendNick(&friend)
	if err != nil {
		log.Printf("%sexe update nick fail %v", logTag, err)
		c.JSON(http.StatusOK, fail(friendV1ExeUpdateFriendNickFail, "fid or uid or id is wrong"))
		return
	}
	response := updateFriendNickResponse{}
	response.Id = friend.Id
	response.Uid = friend.Uid
	response.Fid = friend.Fid
	response.NewNick = friend.Fnick
	c.JSON(http.StatusOK, response)
}

//删除好友
func (self *FriendV1API) deleteFriend(c *gin.Context) {
	logTag := "friend->delete->friend->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		code := HttpErrorContentTypeInvalid
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(code, msg))
		return
	}
	contentLenStr := c.GetHeader("Content-Length")
	if contentLenStr == "" {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenEmpty, "content length is empty"))
		return
	}
	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		log.Printf("%scontent length is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is empty"))
		return
	}
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len, contentLen:%d, bodyLen:%d", logTag, contentLen, len(body))
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &deleteFriendRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal deleteFriendRequest err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid < 1 {
		log.Printf("%suid is invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1UidInvalid, "uid is invalid"))
		return
	}
	if request.Fid < 1 {
		log.Printf("%suid is invalid", logTag)
		c.JSON(http.StatusOK, fail(friendV1FidInvalid, "uid is invalid"))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}
	friend := sql.Friend{Id: request.Id, Fid: request.Fid, Uid: request.Uid}
	err = cache.DelFriend(&friend)
	if err != nil {
		log.Printf("%sexe delete friend fail %v", logTag, err)
		c.JSON(http.StatusOK, fail(friendV1ExeDelFriendFail, "fid or uid or id is wrong"))
		return
	}
	response := deleteFriendResponse{}
	response.ErrorCode = 0
	response.Msg = ""
	c.JSON(http.StatusOK, response)
}
