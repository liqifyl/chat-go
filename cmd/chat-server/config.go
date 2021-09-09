package main

import "net/url"

var (
	LogToStdout         = true
	LogDevelopment      = false
	LogToGraylog        = false
	LogLevelStdout      = "debug"
	LogLevelGraylog     = "info"
	LogToGraylogAddress = &url.URL{}
	ServerListenAddress = &url.URL{}
	UserImageSaveDir    = "/Users/apple/chat/user/image"
)

const (
	TestUid = -2000
)
