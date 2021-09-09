package sql

import (
	"errors"
	"github.com/liqifyl/chat-go/internal/util"
)

const (
	sqlErrorFriendUidInvalid = iota + 100
	sqlErrorFriendFidInvalid
	sqlErrorFriendOpenDbFail
	sqlErrorFriendQFnickInvalid
	sqlErrorFriendCreatePrepareFail
	sqlErrorFriendExecInsertFail
	sqlErrorFriendGetLastInsertIdFail
)

const (
	sqlFriendETimeLayout = "2006-01-02 15:04:05"
)

type Friend struct {
	Id    int64  `json:"id"`    //朋友关系唯一id，方便索引的id；表中字段名为id
	Uid   int64  `json:"uid"`   //用户id，参考user(id)的外键; 表中字段名为uid
	Fid   int64  `json:"fid"`   //朋友id，参考user(id)的外键；表中字段名为fid
	Fnick string `json:"fnick"` //朋友昵称，建立朋友关系时此字段值为朋友的sign;表中字段名为fnick
	Etime string `json:"etime"` //朋友建立时间;表中字段名为etime
}

//添加好友
func AddFriend(friend *Friend) (int64, error) {
	if friend.Uid < 1 {
		return sqlErrorFriendUidInvalid, errors.New("uid is invalid")
	}
	if friend.Fid < 1 {
		return sqlErrorFriendFidInvalid, errors.New("fid is invalid")
	}
	if friend.Fnick == "" {
		return sqlErrorFriendQFnickInvalid, errors.New("fnick is invalid")
	}
	db, err := GetImDb()
	if err != nil {
		return sqlErrorFriendOpenDbFail, err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return 0, err
	}
	etime := util.CurrentTimeStr(sqlFriendETimeLayout)
	stmt, err := db.Prepare("insert into friend(uid, fid, fnick, etime) values(?,?,?,?)")
	if err != nil {
		_ = tx.Rollback()
		return sqlErrorFriendCreatePrepareFail, err
	}
	defer stmt.Close()
	r, err := stmt.Exec(friend.Uid, friend.Fid, friend.Fnick, etime)
	if err != nil {
		_ = tx.Rollback()
		return sqlErrorFriendExecInsertFail, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return sqlErrorFriendGetLastInsertIdFail, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	friend.Etime = etime
	return id, err
}

func DeleteFriendById(id int64) error {
	if id < 1 {
		return errors.New("id is invalid")
	}
	db, err := GetImDb()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return err
	}
	result, err := db.Exec("delete table friend where id = ?", id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("rows affected is 0")
	}
	return nil
}

//删除好友
func DeleteFriend(friend *Friend) error {
	if friend.Uid < 1 {
		return errors.New("uid is invalid")
	}
	if friend.Fid < 1 {
		return errors.New("fid is invalid")
	}
	db, err := GetImDb()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return err
	}
	result, err := db.Exec("delete table friend where uid = ? and fid = ?", friend.Uid, friend.Fid)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("rows affected is 0")
	}
	return nil
}

func UpdateFriendNickBy(id int64, newNick string) error {
	if id < 1 {
		return  errors.New("id is invalid")
	}
	db, err := GetImDb()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return err
	}
	result, err := db.Exec("update table friend set fnick = ? where id = ?", newNick, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("rows affected is 0")
	}
	return nil
 	return nil
}

//更新好友nick
func UpdateFriendNick(friend *Friend) error {
	if friend.Uid < 1 {
		return errors.New("uid is invalid")
	}
	if friend.Fid < 1 {
		return errors.New("fid is invalid")
	}
	if friend.Fnick == "" {
		return errors.New("fnick is empty")
	}
	db, err := GetImDb()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return err
	}
	result, err := db.Exec("update table friend set fnick = ? where uid = ? and fid = ?", friend.Fnick, friend.Uid, friend.Fid)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("rows affected is 0")
	}
	return nil
}

//根据用户id获取所有好友
func GetFriendsByUid(uid int64) ([]*Friend, error) {
	if uid < 1 {
		return nil, errors.New("uid is invalid")
	}
	db, err := GetImDb()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("select id, fid, fnick, etime from friend where uid = ?", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*Friend
	for rows.Next() {
		var fid int64
		var fnick string
		var id int64
		var etime string
		err = rows.Scan(&id, &fid, &fnick, &etime)
		if err != nil {
			break
		}
		tmp := &Friend{Uid: uid, Fid: fid, Fnick: fnick}
		tmp.Uid = uid
		results = append(results, tmp)
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}
