package sql

import (
	"database/sql"
	"sync"
)

var (
	imDb     *sql.DB = nil
	imDbLock         = new(sync.Mutex)
)

func GetImDb() (*sql.DB, error) {
	defer imDbLock.Unlock()
	imDbLock.Lock()
	if imDb == nil {
		db, err := sql.Open("mysql", "root:liqifyl10051113@tcp(127.0.0.1:3306)/im")
		if err != nil {
			return nil, err
		}
		//db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		imDb = db
	}
	return imDb, nil
}
