package sql

import (
	"database/sql"
	"fmt"
	"sync"
)

var (
	imDbLock        = new(sync.Mutex)
	imDbMap         = make(map[string]*sql.DB)
	DefaultDbConfig = DbConfig{}
)

type DbConfig struct {
	DriveName      string
	DataSourceName string
}

func convertDbConfigToStr(config DbConfig) string {
	str := fmt.Sprintf("%s-%s", config.DriveName, config.DataSourceName)
	return str
}

func getImDb() (*sql.DB, error) {
	return getImDbByConfig(DefaultDbConfig)
}

func getImDbByConfig(config DbConfig) (*sql.DB, error) {
	defer imDbLock.Unlock()
	imDbLock.Lock()
	key := convertDbConfigToStr(config)
	db := imDbMap[key]
	if db == nil {
		db, err := sql.Open(config.DriveName, config.DataSourceName)
		if err != nil {
			return nil, err
		}
		//db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
	}
	return db, nil
}
