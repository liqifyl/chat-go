package v1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/liqifyl/chat-go/internal/cache"
	"github.com/liqifyl/chat-go/internal/config"
	"github.com/liqifyl/chat-go/internal/sql"
	"github.com/liqifyl/chat-go/internal/token"
	"github.com/liqifyl/chat-go/internal/util"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	userErrSqlExeErr = iota + 200
	userErrUidInvalid
	userErrPwdInvalid
	userErrLoginFail
	userErrImagePathInvalid
	userErrDecodeImagePathErr
	userErrStatImageErr
	userErrImageFileTypeMismatch
	userErrImageFileOpenFail
	userErrParseMultipartFormFail
	userErrImageFileLenInvalid
	userErrImageFileFormatMismatch
	userErrSaveImageFileFail
	userErrImageFileSizeInvalid
	userErrUpdatePwdFail
	userErrUpdateNameFail
	userErrUpdateSignFail
	userErrUpdateBirthdayFail
	userErrNewNickInvalid
	userErrNewSignInvalid
	userErrNewBirthDayInvalid
)

type registerSuccessResponse struct {
	Uid int64 `json:"id"`
}

type userLoginRequest struct {
	Uid int64  `json:"id"`
	Pwd string `json:"pwd"`
}

type userLoginSuccessResponse struct {
	Id       int64  `json:"id"`
	Nick     string `json:"nick"`
	Sign     string `json:"sign"`
	Birthday string `json:"birthday"`
	Age      uint8  `json:"age"`
	Sex      string `json:"sex"`
	Country  string `json:"country"`
	ImageUrl string `json:"image_url"`
	Token    string `json:"token"`
}

type userUpdatePwdRequest struct {
	userLoginRequest
	NewPwd string `json:"new_pwd"`
}

type userUpdateNickRequest struct {
	userLoginRequest
	NewNick string `json:"new_nick"`
}

type userUpdateSignRequest struct {
	userLoginRequest
	NewSign string `json:"new_sign"`
}

type userUpdateBirthdayRequest struct {
	userLoginRequest
	NewBirthday string `json:"new_birthday"`
}

type userUpdateImageResponse struct {
	NewImageUrl string `json:"new_image_url"`
}

type UserV1API struct {
	Config config.GinServerConfig
}

func NewUserV1API(config config.GinServerConfig) *UserV1API {
	return &UserV1API{Config: config}
}

//注册所有对外输出接口
func (self *UserV1API) RegisterUserRestfulAPI(gin *gin.Engine) {
	gin.POST("/v1/user/register", self.register)
	gin.POST("/v1/user/login", self.login)
	gin.POST("/v1/user/update/pwd", self.updatePwd)
	gin.POST("/v1/user/update/image", self.updateImage)
	gin.POST("/v1/user/update/nick", self.updateNick)
	gin.POST("/v1/user/update/sign", self.updateSign)
	gin.POST("/v1/user/update/birthday", self.updateBirthDay)
	gin.GET("/v1/user/image/:id", self.getUserImage)
}

func (self *UserV1API) generateUserImageUrl(id int64) string {
	if self.Config.HostName == "" || self.Config.Port == "" {
		return ""
	}
	url := fmt.Sprintf("%s:%sv1/user/image/", self.Config.HostName, self.Config.Port)
	idStr := fmt.Sprintf("%d", id)
	fileName := fmt.Sprintf("%s.png", idStr)
	absolutePath := self.Config.UserImageSaveDir + "/image/" + idStr + "/" + fileName
	if !util.FileIsExist(absolutePath) {
		return ""
	}
	base64FileName := base64.StdEncoding.EncodeToString([]byte(fileName))
	url = fmt.Sprintf("%s%d?url=%s", url, id, base64FileName)
	return url
}

//用户注册
func (self *UserV1API) register(c *gin.Context) {
	logTag := "user->register->"
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
	user := &sql.ChatUser{}
	err = json.Unmarshal(body, user)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	uid, err := sql.InsertUser(user)
	if err != nil {
		log.Printf("%sinsert user info err %v, code:%d", logTag, err, uid)
		c.JSON(http.StatusOK, fail(userErrSqlExeErr, err.Error()))
		return
	}
	c.JSON(http.StatusOK, registerSuccessResponse{Uid: uid})
}

//登录
func (self *UserV1API) login(c *gin.Context) {
	logTag := "user->login->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		log.Printf("%s%s", logTag, msg)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
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
	log.Printf("%s%s", logTag, string(body))
	request := &userLoginRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	log.Printf("%s(%d,%s)", logTag, request.Uid, request.Pwd)
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "uid invalid"))
		return
	}
	if request.Pwd == "" {
		log.Printf("%spwd is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrPwdInvalid, "pwd is empty"))
		return
	}
	user := &sql.ChatUser{Id: request.Uid, Password: request.Pwd}
	code, err := cache.UserLogin(user)
	if err != nil {
		log.Printf("%sexe login fail, (%d, %v)", logTag, code, err)
		c.JSON(http.StatusOK, fail(userErrLoginFail, err.Error()))
		return
	}
	//返回用户哪些数据
	response := userLoginSuccessResponse{}
	response.Id = user.Id
	if user.Sex == 0 {
		response.Sex = "男"
	} else {
		response.Sex = "女"
	}
	response.Country = user.Country
	response.Sign = user.Sign
	response.Birthday = user.Birthday
	response.Nick = user.Nick
	response.Age = user.Age
	response.ImageUrl = self.generateUserImageUrl(request.Uid)
	token, err := token.GenerateToken(user.Id)
	if err != nil {
		log.Printf("%sgenerate token fail, %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorGenerateTokenFail, "generate token fail"))
		return
	}
	response.Token = token
	c.JSON(http.StatusOK, response)
}

//更新密码
func (self *UserV1API) updatePwd(c *gin.Context) {
	logTag := "user->updatePwd->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
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
		log.Printf("%scontent length is not equal body len", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}

	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	request := &userUpdatePwdRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "uid is le 0"))
		return
	}

	user := &sql.ChatUser{Id: request.Uid, Password: request.Pwd}
	code, err := cache.UpdateUserPwd(user, request.NewPwd)
	if err != nil {
		log.Printf("%supdate user pwd failed, (%d, %v)", logTag, code, err)
		c.JSON(http.StatusOK, fail(userErrUpdatePwdFail, err.Error()))
		return
	}
	c.JSON(http.StatusOK, ok())
}

func (self *UserV1API) saveImageToDisk(httpFile *multipart.FileHeader, id string) error {
	imageSaveDir := self.Config.UserImageSaveDir + "/image/" + id
	saveDirInfo, err := os.Stat(imageSaveDir)
	if err != nil {
		//不存在
		err := os.MkdirAll(imageSaveDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		//存在
		if !saveDirInfo.IsDir() {
			return errors.New("save path is directory")
		}
	}
	//创建文件
	saveFilePath := imageSaveDir + "/" + id + ".png"
	saveFileTmpPath := saveFilePath + ".tmp"
	saveFileTmp, err := os.Create(saveFileTmpPath)
	if err != nil {
		return err
	} else {
		defer saveFileTmp.Close()
		reader, err := httpFile.Open()
		if err != nil {
			return err
		} else {
			defer reader.Close()
			contents, err := ioutil.ReadAll(reader)
			if err != nil {
				return err
			}
			if len(contents) == 0 {
				return errors.New("upload file size is zero")
			}
			n, err := saveFileTmp.Write(contents)
			if err != nil {
				return err
			} else {
				log.Printf("write %d to %s", n, saveFilePath)
			}
		}
	}
	err = os.Rename(saveFileTmpPath, saveFilePath)
	if err != nil {
		removeErr := os.Remove(saveFileTmpPath)
		if removeErr != nil {
			log.Printf("remove %s error %v", saveFileTmpPath, removeErr)
		}
		return err
	}
	return nil
}

//更新用户图像
func (self *UserV1API) updateImage(c *gin.Context) {
	logTag := "user->updateImage->"
	contentType := c.ContentType()
	if !strings.HasPrefix(contentType, HttpMultipartFormData) {
		msg := "content type must be multipart/form-data"
		log.Printf("%s%s,%s", logTag, msg, contentType)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
		return
	}
	userIdStr := c.GetHeader(UserIdKey)
	if userIdStr == "" {
		log.Printf("%suserId is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "userId is empty"))
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		log.Printf("%sconvert uid error %v", logTag, err)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "convert uid fail"))
		return
	}
	if userId < 1 {
		log.Printf("%suser id is le to 0", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "user id is le to 0"))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	multipartFrom, err := c.MultipartForm()
	if err != nil {
		log.Printf("%sParseMultipartForm error %v", logTag, err)
		c.JSON(http.StatusOK, fail(userErrParseMultipartFormFail, err.Error()))
		return
	}

	imageFiles := multipartFrom.File["image"]
	if len(imageFiles) == 0 {
		log.Printf("%simage file count must greater than 0", logTag)
		c.JSON(http.StatusOK, fail(userErrImageFileLenInvalid, "image file count must greater than 0"))
		return
	}
	imageFile := imageFiles[0]
	if imageFile.Header.Get(HttpContentTypeKey) != HttpImagePng {
		log.Printf("%simage file format must be image/png", logTag)
		c.JSON(http.StatusOK, fail(userErrImageFileFormatMismatch, "image format must be image/png"))
		return
	}
	if imageFile.Size == 0 {
		log.Printf("%simage file size must be greater than 0", logTag)
		c.JSON(http.StatusOK, fail(userErrImageFileSizeInvalid, "image file size must be greater than 0"))
		return
	}
	err = self.saveImageToDisk(imageFile, userIdStr)
	if err != nil {
		log.Printf("%ssave image error %v", logTag, err)
		c.JSON(http.StatusOK, fail(userErrSaveImageFileFail, err.Error()))
		return
	}
	//生成新的用户图像url
	newImageUrl := self.generateUserImageUrl(int64(userId))
	response := userUpdateImageResponse{NewImageUrl: newImageUrl}
	c.JSON(http.StatusOK, response)
}

//更新用户名
func (self *UserV1API) updateNick(c *gin.Context) {
	logTag := "user->updateNick->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
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
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &userUpdateNickRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "uid is le 0"))
		return
	}

	if request.NewNick == "" {
		log.Printf("%snew nick is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrNewNickInvalid, "new nick is empty"))
		return
	}
	user := &sql.ChatUser{Id: request.Uid, Password: request.Pwd}
	code, err := cache.UpdateUserNick(user, request.NewNick)
	if err != nil {
		log.Printf("%supdate user name fail, (%d, %v)", logTag, code, err)
		c.JSON(http.StatusOK, fail(userErrUpdateNameFail, err.Error()))
		return
	}
	c.JSON(http.StatusOK, ok())
}

//更新用户签名
func (self *UserV1API) updateSign(c *gin.Context) {
	logTag := "user->updateSign->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
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
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &userUpdateSignRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "uid is le 0"))
		return
	}
	if request.NewSign == "" {
		log.Printf("%snew sign is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrNewSignInvalid, "new self sign is empty"))
		return
	}
	user := &sql.ChatUser{Id: request.Uid, Password: request.Pwd}
	code, err := cache.UpdateUserSign(user, request.NewSign)
	if err != nil {
		log.Printf("%supdate user self sign fail, (%d, %v)", logTag, code, err)
		c.JSON(http.StatusOK, fail(userErrUpdateSignFail, err.Error()))
		return
	}
	c.JSON(http.StatusOK, ok())
}

//更新用户生日
func (self *UserV1API) updateBirthDay(c *gin.Context) {
	logTag := "user->updateBirthDay->"
	contentType := c.ContentType()
	if contentType != HttpApplicationJson {
		msg := "content type must application/json"
		log.Printf("%scontent type must application/json", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentTypeInvalid, msg))
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
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		log.Printf("%sread body err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorReadBodyFail, err.Error()))
		return
	}
	if len(body) != contentLen {
		log.Printf("%scontent length is not equal body len", logTag)
		c.JSON(http.StatusOK, fail(HttpErrorContentLenInvalid, "content length is not equal body len"))
		return
	}
	request := &userUpdateBirthdayRequest{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("%smarshal user info err %v", logTag, err)
		c.JSON(http.StatusOK, fail(HttpErrorMarshalJsonFail, err.Error()))
		return
	}
	if request.Uid <= 0 {
		log.Printf("%suid invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "uid is le 0"))
		return
	}
	if request.NewBirthday == "" {
		log.Printf("%snew birthday is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrNewBirthDayInvalid, "new birthday is empty"))
		return
	}
	user := &sql.ChatUser{Id: request.Uid, Password: request.Pwd}
	code, err := cache.UpdateUserBirthday(user, request.NewBirthday)
	if err != nil {
		log.Printf("%supdate user birthday fail, (%d, %v)", logTag, code, err)
		c.JSON(http.StatusOK, fail(userErrUpdateBirthdayFail, err.Error()))
		return
	}
	c.JSON(http.StatusOK, ok())
}

//获取用户图像
func (self *UserV1API) getUserImage(c *gin.Context) {
	logTag := "user->getUserImage->"
	id := c.Param("id")
	if id == "" {
		log.Printf("%suser id is empty", logTag)
		c.JSON(http.StatusOK, fail(userErrUidInvalid, "user id is empty"))
		return
	}
	//查询用户是否存在
	path := c.Query("url")
	if path == "" {
		log.Printf("%surl is invalid", logTag)
		c.JSON(http.StatusOK, fail(userErrImagePathInvalid, "url is invalid"))
		return
	}
	decodePathBytes, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		log.Printf("%s%s", logTag, err.Error())
		c.JSON(http.StatusOK, fail(userErrDecodeImagePathErr, err.Error()))
		return
	}
	if !verifyToken(c, logTag, self.Config.TestUid) {
		return
	}

	decodePath := string(decodePathBytes)
	imageAbsPath := self.Config.UserImageSaveDir + "/image/" + id + "/" + decodePath
	fileInfo, err := os.Stat(imageAbsPath)
	if err != nil {
		log.Printf("%s%s", logTag, err.Error())
		c.JSON(http.StatusOK, fail(userErrStatImageErr, err.Error()))
		return
	}
	if fileInfo.IsDir() {
		log.Printf("%s%s is dir", logTag, imageAbsPath)
		c.JSON(http.StatusOK, fail(userErrImageFileTypeMismatch, "image is directory"))
		return
	}
	file, err := os.Open(imageAbsPath)
	if err != nil {
		log.Printf("%sopen %s error %v", logTag, imageAbsPath, err)
		c.JSON(http.StatusOK, fail(userErrImageFileOpenFail, err.Error()))
		return
	}
	defer file.Close()
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("%sread %s error %v", logTag, imageAbsPath, err)
		return
	}
	c.Header(HttpContentTypeKey, HttpImagePng)
	c.Header(HttpContentLengthKey, fmt.Sprintf("%d", len(contents)))
	c.Header(HttpResponseServerKey, "com.liqi")
	c.Status(http.StatusOK)
	n, err := c.Writer.Write(contents)
	if err != nil {
		log.Printf("%sfail to write : %v", logTag, err)
	} else {
		log.Printf("%swrite : %d", logTag, n)
	}
}
