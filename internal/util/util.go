package util

import (
	"os"
	"time"
)

func FileIsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func CurrentTimeStr(layout string) string {
	now := time.Now()
	return now.Format(layout)
}

func ChangeWd(newWd string) {
}
