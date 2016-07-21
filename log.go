// Package log provides 全局日志记录，内部使用beego/log
package engine

import (
	"errors"
	"github.com/astaxie/beego/logs"
)

var Log *logs.BeeLogger

func GlobalInit(kind, path, level string, length int) error {
	if Log == nil {
		Log = logs.NewLogger(int64(length))
	}

	err := Log.SetLogger(kind, path)
	if err != nil {
		return err
	}

	switch level {
	case "debug":
		Log.SetLevel(logs.LevelDebug)
	case "info":
		Log.SetLevel(logs.LevelInfo)
	case "warn":
		Log.SetLevel(logs.LevelWarn)
	case "error":
		Log.SetLevel(logs.LevelError)
	default:
		return errors.New("未处理的日志记录等级")
	}

	return nil

}
