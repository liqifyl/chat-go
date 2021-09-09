package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liqifyl/chat-go/internal/cache"
	token2 "github.com/liqifyl/chat-go/internal/token"
	"log"
	"net/http"
	"strings"
)

type errResponse struct {
	ErrorCode int    `json:"code"`
	Msg       string `json:"msg"`
}

func fail(code int, msg string) gin.H {
	response := errResponse{ErrorCode: code, Msg: msg}
	return gin.H{"err": response}
}

func ok() gin.H {
	response := errResponse{ErrorCode: 0, Msg: ""}
	return gin.H{"err": response}
}

func exeVerifyToken(token string, testUid int64) error {
	if token == "" {
		return errors.New("token is empty")
	}
	if !strings.HasPrefix(token, HttpTokenPrefix) {
		msg := fmt.Sprintf("token prefix must be %s", HttpTokenPrefix)
		return errors.New(msg)
	}
	realToken := token[len(HttpTokenPrefix):]
	claims, err := token2.ParseToken(realToken)
	if err != nil {
		return err
	}
	if claims.Issuer != token2.TokenIssuer {
		return errors.New("issuer invalid")
	}
	//测试token直接返回
	if claims.Uid == testUid {
		return nil
	}
	_, err = cache.IsExistOfUser(claims.Uid)
	if err != nil {
		return err
	}
	return nil
}

func verifyToken(c *gin.Context, logTag string, testUid int64) bool {
	token := c.GetHeader(HttpTokenKey)
	if token == "" {
		log.Printf("%stoken is empty", logTag)
		c.JSON(http.StatusOK, fail(HttpTokenEmpty, "token is empty"))
		return false
	}
	err := exeVerifyToken(token, testUid)
	if err != nil {
		log.Printf("%scheck token err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpTokenEmpty, err.Error()))
		return false
	}
	return true
}
