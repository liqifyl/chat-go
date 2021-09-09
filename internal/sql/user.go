package sql

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/liqifyl/chat-go/internal/util"
	"log"
	"time"
)

var (
	sqlUserBirthdayLayout = "2006-01-02 15:04:05"
)

type ChatUser struct {
	Id          int64  `json:"id"`       //用户id，这个是数据库生成的；表的中字段名为id
	Nick        string `json:"nick"`     //用户名, 长度[0,200]；表的中字段名为nick
	Password    string `json:"pwd"`      //密码，长度[0,30]，表中的字段名为password
	Age         uint8  `json:"age"`      //年龄，[0,255]；表中的字段名为age
	Birthday    string `json:"birthday"` //生日，datetime，YYYY-mm-dd HH::MM::SS；表中字段名为birthday
	Sign        string `json:"sign"`     //个性签名，[0,100];表中字段名为sign
	Country     string `json:"country"`  //国家，[0,20];表中字段名为country
	Sex         uint8  `json:"sex"`      //0男，1女;表中字段名为sex
	PhoneNumber string `json:"pnumber"`  //长度11，必须全部是数字；表中字段名为pnumber
}

func QueryUserById(id int64) ([]*ChatUser, error) {
	if id < 1 {
		return nil, errors.New("id must be greater than 0")
	}
	db, err := GetImDb()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("select nick,password,age,birthday,sign,country,sex,pnumber from user where id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*ChatUser
	for rows.Next() {
		var nick string
		var password string
		var age uint8
		var birthday string
		var sign string
		var country string
		var sex uint8
		var phoneNumber string
		err = rows.Scan(&nick, &password, &age, &birthday, &sign, &country, &sex, &phoneNumber)
		if err != nil {
			break
		}
		user := ChatUser{Id: id, Nick: nick, Password: password, Age: age, Birthday: birthday, Sign: sign, Country: country, Sex: sex, PhoneNumber: phoneNumber}
		results = append(results, &user)
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func verifyPhoneNumber(phoneNumber string) error {
	if len(phoneNumber) != 11 {
		return errors.New("phone number length must equal 11")
	}
	var err  error = nil
	for index, num := range phoneNumber {
		if index == 0 {
			if num <= '0' {
				err = errors.New("phone number first must be greater than 0")
				break
			}
		}
		if num >= '0' && num <= '9' {
			continue
		}
		err = errors.New("phone number all must be digit")
		break
	}
	return err
}

//添加用户
func InsertUser(user *ChatUser) (int64, error) {
	if len(user.Nick) == 0 {
		return 0, errors.New("name is empty")
	}
	if len(user.Password) == 0 {
		return 0, errors.New("password is empty")
	}
	//if len(user.Sign) == 0 {
	//	return 0, errors.New("self sign is empty")
	//}
	if len(user.Country) == 0 {
		user.Country = "China"
	}
	if user.Sex > 1 {
		user.Sex = 1
	} else {
		user.Sex = 0
	}
	err := verifyPhoneNumber(user.PhoneNumber)
	if err != nil {
		return 0, err
	}
	//查询电话号码是否存在
	if user.Birthday == "" {
		t := time.Now()
		user.Birthday = t.Format("2006-01-02 15:04:05")
	}
	rtime := util.CurrentTimeStr("2006-01-02 15:04:05")
	db, err := GetImDb()
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return 0, err
	}
	stmt, err := db.Prepare("insert into user(nick, password, age, birthday, sign, country, sex, pnumber, rtime) values(?,?,?,?,?,?,?,?,?)")
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	r, err := stmt.Exec(user.Nick, user.Password, user.Age, user.Birthday, user.Sign, user.Country, user.Sex, user.PhoneNumber, rtime)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return id, err
}

//更新用户密码
func UpdateUserPwd(user *ChatUser, newPwd string) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Id <= 0 {
		return errors.New("user is invalid")
	}
	if len(newPwd) == 0 {
		return errors.New("new password is empty")
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
	stmt, err := db.Prepare("UPDATE user SET password = ? WHERE id = ?")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	r, err := stmt.Exec(newPwd, user.Id)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("exe update user pwd error %v", err)
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("update pwd fail, because user is not exist")
	}
	return nil
}

//更新用户Nick
func UpdateUserNick(user *ChatUser, newNick string) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Id <= 0 {
		return errors.New("user is invalid")
	}
	if newNick == "" {
		return errors.New("new nick is empty")
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
	stmt, err := db.Prepare("UPDATE user SET nick = ? WHERE id = ?")
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	r, err := stmt.Exec(newNick, user.Id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("update name fail, because user is not exist")
	}
	return nil
}

//更新用户签名
func UpdateUserSign(user *ChatUser, newSign string) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Id <= 0 {
		return errors.New("user is invalid")
	}
	if len(newSign) == 0 {
		return errors.New("new self sign is empty")
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
	stmt, err := db.Prepare("UPDATE user SET sign = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	r, err := stmt.Exec(newSign, user.Id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("update sign fail, because user is not exist")
	}
	return nil
}

//更新用户生日
func UpdateUserBirthday(user *ChatUser, newBirthday string) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Id <= 0 {
		return errors.New("user is invalid")
	}
	if len(newBirthday) == 0 {
		return errors.New("new birthday is empty")
	}
	_, err := time.Parse(sqlUserBirthdayLayout, newBirthday)
	if err != nil {
		return err
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
	stmt, err := db.Prepare("UPDATE user SET birthday = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	r, err := stmt.Exec(newBirthday, user.Id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("update birthday fail, because user is not exist")
	}
	return nil
}

//查看用户签名
func GetUserSignById(id int64) (string, error) {
	db, err := GetImDb()
	if err != nil {
		return "", err
	}
	rows, err := db.Query("select sign from user where id=?", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var sign string
	for rows.Next() {
		err = rows.Scan(&sign)
		break
	}
	if err != nil {
		return "", err
	}
	return sign, nil
}

//获得用户nick
func GetUserNick(id int64) (string, error) {
	db, err := GetImDb()
	if err != nil {
		return "", err
	}
	rows, err := db.Query("select nick from user where id=?", id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var nick string
	for rows.Next() {
		err = rows.Scan(&nick)
		break
	}
	if err != nil {
		return "", err
	}
	return nick, nil
}
