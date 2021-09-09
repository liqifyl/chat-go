package sql

//对应im数据库中的friend_circle表
type FriendCircle struct {
	Id    int64  `json:"id"`    //id；表中字段名称id
	Uid   int64  `json:"uid"`   //用户id，user表中id;表中字段名称uid
	Ptime string `json:"ptime"` //朋友圈发布时间;表中对应字段名称ptime
	title string `json:"title"` //朋友圈对应标题；表中对应名称title
}

//发布一条朋友圈
func PublishFriendCircle(friendCircle *FriendCircle) (int64, error) {
	return 0, nil
}

//根据唯一id删除一条朋友圈
func RemoveFriendCircleById(id int64) error {
	return nil
}

//根据用户id获取自己以及朋友最新朋友圈,根据ptime降序
func GetFriendCircleByUid(uid int64, maxPublishTime string, limit int) ([]*FriendCircle, error) {
	return nil, nil
}
