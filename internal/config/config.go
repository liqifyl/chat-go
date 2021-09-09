package config

type GinServerConfig struct {
	UserImageSaveDir        string
	HostName                string
	Port                    string
	TestUid                 int64
	MysqlChatDataSourceName string
	RedisServerAddress      string
	RedisServerPwd          string
	RedisSelectDB           int
}