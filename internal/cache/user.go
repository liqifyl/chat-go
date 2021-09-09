package cache

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/liqifyl/chat-go/internal/sql"
	"log"
	"time"
)

const (
	cacheUserOK                   = 0
	cacheUserQueryUserErrorFromDb = 101
	cacheUserIsEmptyFromDb        = 102
	cacheUserPasswordWrong        = 103
)

const (
	cacheUserPrefix = "user"
)

//从数据库中查询用户
func queryUserFromDbById(id int64) (*sql.ChatUser, error) {
	users, err := sql.QueryUserById(id)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("query user is empty")
	}
	return users[0], nil
}

//将user序列化成json字符串
func marshalUser(user *sql.ChatUser) (string, error) {
	userJson, err := json2.Marshal(user)
	if err != nil {
		return "", err
	}
	userStr := string(userJson)
	return userStr, nil
}

//从数据库中查询用户并将user信息序列化成json字符串
func queryUserFromDbWithStr(id int64) (string, error) {
	user, err := queryUserFromDbById(id)
	if err != nil {
		return "", err
	}
	return marshalUser(user)
}

func generateUserCacheKeyById(id int64) string {
	return generateUserCacheKey(id, "")
}

func generateUserCacheKey(id int64, suffix string) string {
	if suffix == "" {
		return fmt.Sprintf("%s-%d", cacheUserPrefix, id)
	}
	return fmt.Sprintf("%s-%d-%s", cacheUserPrefix, id, suffix)
}

/***
根据id查询用户是否存在
返回error为空存在，不为空不存在
*/
func IsExistOfUser(id int64) (int, error) {
	client := getRedisClient()
	keyId := generateUserCacheKeyById(id)
	userExist, err := client.Exists(cacheRedisCtx, keyId).Result()
	if err != nil {
		log.Printf("cache exe exists error %v", err)
		//从数据库中查询
		ret, err := queryUserFromDbWithStr(id)
		if err != nil {
			return cacheUserQueryUserErrorFromDb, err
		}
		if ret == "" {
			return cacheUserIsEmptyFromDb, errors.New("query user from db, but user info is empty")
		} else {
			return cacheUserOK, nil
		}
	}
	if userExist == 0 {
		//如果不存在redis中，从数据库查询
		ret, err := queryUserFromDbWithStr(id)
		if err != nil {
			return cacheUserQueryUserErrorFromDb, err
		}
		if ret == "" {
			return cacheUserIsEmptyFromDb, errors.New("query user from db, but user info is empty")
		}
		//将数据库中查询到数据保存到redis中
		cmdRes, err := client.SetNX(cacheRedisCtx, keyId, ret, time.Second*5).Result()
		log.Printf("save user info to cache (%d, %v)", cmdRes, err)
		return cacheUserOK, nil
	}
	log.Printf("user info is exist in cache")
	return cacheUserOK, nil
}

func copyUser(dst *sql.ChatUser, src *sql.ChatUser) {
	dst.Birthday = src.Birthday
	dst.Nick = src.Nick
	dst.Age = src.Age
	dst.Sign = src.Sign
	dst.Country = src.Country
	dst.Sex = src.Sex
	dst.PhoneNumber = src.PhoneNumber
}

//用户登录
func UserLogin(user *sql.ChatUser) (int, error) {
	client := getRedisClient()
	keyId := generateUserCacheKeyById(user.Id)
	userJsonStr, err := client.Get(cacheRedisCtx, keyId).Result()
	if err == nil && userJsonStr != "" {
		cacheUser := &sql.ChatUser{}
		err = json2.Unmarshal([]byte(userJsonStr), cacheUser)
		if err != nil {
			log.Printf("unmarshal user string error %v", err)
		} else {
			if cacheUser.Id == user.Id {
				if user.Password != cacheUser.Password {
					err = errors.New("password is wrong")
				} else {
					copyUser(user, cacheUser)
					return cacheUserOK, nil
				}
			} else {
				err = errors.New("cacheUser id is not equal request user id")
			}
		}
	}

	log.Printf("get user info from cache error %v", err)
	//从数据库中查询
	ret, err := queryUserFromDbById(user.Id)
	if err != nil {
		return cacheUserQueryUserErrorFromDb, err
	}
	if ret == nil {
		return cacheUserIsEmptyFromDb, errors.New("query user from db, but user info is empty")
	} else {
		if user.Password != ret.Password {
			return cacheUserPasswordWrong, errors.New("password is wrong")
		}
		copyUser(user, ret)
		//保存用户信息到redis
		userMarshalStr, err := marshalUser(ret)
		if err != nil {
			log.Printf("marshal user info error %v", err)
		} else {
			cmdRes, err := client.SetNX(cacheRedisCtx, keyId, userMarshalStr, time.Second * 5).Result()
			log.Printf("save user info to cache (%d, %v)", cmdRes, err)
		}
		return 0, nil
	}
}

//从redis中删除用户
func DeleteUserFromRedis(user *sql.ChatUser) (int, error) {
	client := getRedisClient()
	keyId := generateUserCacheKeyById(user.Id)
	_, err := client.Del(cacheRedisCtx, keyId).Result()
	if err != nil {
		return -1, err
	}
	return 0, nil
}

//更新用户密码，对于redis缓存和数据库同步问题使用策略是双删策略
func UpdateUserPwd(user *sql.ChatUser, newPwd string) (int, error) {
	logTag := "UpdateUserPwd->"
	err := sql.UpdateUserPwd(user, newPwd)
	if err != nil {
		return -1, err
	}
	code, err := DeleteUserFromRedis(user)
	if err != nil {
		log.Printf("%s again delete user fail from redis, (%d,%v)", logTag, code, err)
	}
	return 0, nil
}

//更新用户nick
func UpdateUserNick(user *sql.ChatUser, newNick string) (int, error) {
	logTag := "UpdateUserNick->"
	err := sql.UpdateUserNick(user, newNick)
	if err != nil {
		return -1, err
	}
	code, err := DeleteUserFromRedis(user)
	if err != nil {
		log.Printf("%s again delete user fail from redis, (%d,%v)", logTag, code, err)
	}
	return 0, nil
}

//更新用户签名
func UpdateUserSign(user *sql.ChatUser, newSign string) (int, error) {
	logTag := "UpdateUserSign->"
	err := sql.UpdateUserSign(user, newSign)
	if err != nil {
		return -1, err
	}
	code, err := DeleteUserFromRedis(user)
	if err != nil {
		log.Printf("%s again delete user fail from redis, (%d,%v)", logTag, code, err)
	}
	return 0, nil
}

//更新用户生日
func UpdateUserBirthday(user *sql.ChatUser, newBirthday string) (int, error) {
	logTag := "UpdateUserBirthday->"
	err := sql.UpdateUserBirthday(user, newBirthday)
	if err != nil {
		return -1, err
	}
	code, err := DeleteUserFromRedis(user)
	if err != nil {
		log.Printf("%s again delete user fail from redis, (%d,%v)", logTag, code, err)
	}
	return 0, nil
}

//通过id获取sign
func GetUserSignById(id int64) (string, error) {
	logTag := "GetUserSignById->"
	client := getRedisClient()
	keyId := generateUserCacheKey(id, "sign")
	cacheSign, err := client.Get(cacheRedisCtx, keyId).Result()
	if err != nil || cacheSign == "" {
		if err != nil {
			log.Printf("%sget sign from redis error %v", logTag, err)
		} else {
			log.Printf("%ssign is empty from  redis%v", logTag, err)
		}
		//从数据库中查询
		ret, err := sql.GetUserSignById(id)
		if err != nil {
			log.Printf("%sget sign from mysql error %v", logTag, err)
			return "", err
		}
		status, err := client.Set(cacheRedisCtx, keyId, ret, time.Second*5).Result()
		log.Printf("%ssave %v to redis result(%s,%v)", logTag, keyId, status, err)
		return ret, nil
	}
	return cacheSign, nil
}

//通过id获取nick
func GetUserNickById(id int64) (string, error) {
	logTag := "GetUserNickById->"
	client := getRedisClient()
	keyId := generateUserCacheKey(id, "nick")
	cacheSign, err := client.Get(cacheRedisCtx, keyId).Result()
	if err != nil || cacheSign == "" {
		if err != nil {
			log.Printf("%sget nick from redis error %v", logTag, err)
		} else {
			log.Printf("%snick is empty from  redis%v", logTag, err)
		}
		//从数据库中查询
		ret, err := sql.GetUserSignById(id)
		if err != nil {
			log.Printf("%sget nick from mysql error %v", logTag, err)
			return "", err
		}
		status, err := client.Set(cacheRedisCtx, keyId, ret, time.Second*5).Result()
		log.Printf("%ssave %v to redis result(%s,%v)", logTag, keyId, status, err)
		return ret, nil
	}
	return cacheSign, nil
}
