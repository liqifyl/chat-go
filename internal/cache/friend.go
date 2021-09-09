package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/liqifyl/chat-go/internal/sql"
	"log"
	"time"
)

const (
	cacheFriendPrefix = "friend"
)

func generateFriendCacheKey(id int64, suffix string) string {
	if suffix == "" {
		return fmt.Sprintf("%s-%d", cacheFriendPrefix, id)
	}
	return fmt.Sprintf("%s-%d-%s", cacheFriendPrefix, id, suffix)
}

func generateFriendCacheKeyByUid(id int64) string {
	return generateFriendCacheKey(id, "")
}

func delFriendsFromCacheByUid(uid int64) {
	client := getRedisClient()
	keyId := generateFriendCacheKeyByUid(uid)
	_, err := client.Del(cacheRedisCtx, keyId).Result()
	if err != nil {
		log.Printf("delFriendsFromCacheByUid->del %s error %v", keyId, err)
	}
}

//添加好友
func AddFriend(friend *sql.Friend) (int64, error) {
	logTag := "AddFriend->"
	nick, err := GetUserNickById(friend.Fid)
	if err != nil {
		return 0, err
	}
	if nick == "" {
		errMsg := fmt.Sprintf("%d is not exist", friend.Fid)
		log.Printf("%s%s", logTag, errMsg)
		return 0, errors.New(errMsg)
	}
	friend.Fnick = nick
	id, err := sql.AddFriend(friend)
	if err != nil {
		return 0, err
	}
	delFriendsFromCacheByUid(friend.Uid)
	return id, nil
}

//删除好友
func DelFriend(friend *sql.Friend) error {
	_ = "DelFriend->"
	if friend.Id > 0 {
		err := sql.DeleteFriendById(friend.Id)
		if err == nil {
			if friend.Uid > 0 {
				delFriendsFromCacheByUid(friend.Uid)
			}
			return nil
		}
	}
	err := sql.DeleteFriend(friend)
	if err != nil {
		return err
	}
	delFriendsFromCacheByUid(friend.Uid)
	return nil
}

//更新好友昵称
func UpdateFriendNick(friend *sql.Friend) error {
	_ = "UpdateFriendNick->"
	if friend.Id >= 1 {
		err := sql.UpdateFriendNickBy(friend.Id, friend.Fnick)
		if err == nil {
			if friend.Uid > 0 {
				delFriendsFromCacheByUid(friend.Uid)
			}
			return nil
		}
	}
	err := sql.UpdateFriendNick(friend)
	if err != nil {
		return err
	}
	delFriendsFromCacheByUid(friend.Uid)
	return nil
}

//将friends序列化成json字符串
func marshalFriends(friends []*sql.Friend) (string, error) {
	jsonBytes, err := json.Marshal(friends)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)
	return jsonStr, nil
}

//获取用户所有好友
func GetFriendsByUid(uid int64) ([]*sql.Friend, error) {
	logTag := "GetFriends->"
	client := getRedisClient()
	keyId := generateFriendCacheKey(uid, "friends")
	friendsJsonStr, err := client.Get(cacheRedisCtx, keyId).Result()
	if err != nil || friendsJsonStr == "" {
		if err != nil {
			log.Printf("%sget friends from cache error %v", logTag, err)
		} else {
			log.Printf("%sfriends is empty from cache", logTag)
		}
		friends, err := sql.GetFriendsByUid(uid)
		if err != nil {
			return nil, err
		}
		//保存到redis中
		friendsJsonStr, saveToCacheErr := marshalFriends(friends)
		if saveToCacheErr != nil {
			log.Printf("%smarshal friends error %v", logTag, saveToCacheErr)
		} else {
			str, saveToCacheErr := client.Set(cacheRedisCtx, keyId, friendsJsonStr, time.Second*5).Result()
			log.Printf("%sexe save %s to redis result(%s,%v)", logTag, keyId, str, saveToCacheErr)
		}
		return friends, err
	}
	var friends []*sql.Friend
	err = json.Unmarshal([]byte(friendsJsonStr), &friends)
	if err != nil {
		return nil, err
	}
	return friends, nil
}
