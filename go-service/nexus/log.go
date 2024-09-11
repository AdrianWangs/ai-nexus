package main

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path"
	"runtime"
)

type MyFormatter struct{}

// 颜色
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

func (m *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	//根据不同的level去展示颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//自定义日期格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	if entry.HasCaller() {

		// 取出栈中的第二个调用者信息
		pc, callerFile, callerLine, ok := runtime.Caller(8)

		var callerFuncName string
		if !ok {
			callerFile = "unknown"
			callerLine = 0
			callerFuncName = "unknown"
		} else {
			// 处理文件名
			callerFile = path.Base(callerFile)

			// 获取函数名
			funcObj := runtime.FuncForPC(pc)
			if funcObj != nil {
				callerFuncName = funcObj.Name()
			} else {
				callerFuncName = "unknown"
			}
		}

		//自定义文件路径
		fileVal := fmt.Sprintf("%s:%d", path.Base(callerFile), callerLine)
		//自定义输出格式
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s %s %s\n", timestamp, levelColor, entry.Level, fileVal, callerFuncName, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m  %s\n", timestamp, levelColor, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}
