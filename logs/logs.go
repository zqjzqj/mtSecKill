package logs

import (
	"github.com/gookit/color"
	"log"
	"os"
)

var logger *log.Logger

func AllowFileLogs() {
	logsFile, err := os.OpenFile("mtSecKill.log",  os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		logger = log.New(logsFile, "", log.Lshortfile|log.LstdFlags)
	}
}

//这个包用来统一的日志输出处理
//目前只做简单两个方法 后续根据具体需要在这里增加日志操作
func Println(v ...interface{}) {
	if logger != nil {
		logger.Println(v)
	}
	log.Println(v...)
}

func print2(color2 color.Color, v ...interface{}) {
	if logger != nil {
		logger.Println(v)
	}
	color2.Light().Println(v...)
}

func PrintlnSuccess(v ...interface{}) {
	print2(color.Green, v...)
}

func PrintlnInfo(v ...interface{}) {
	print2(color.LightCyan, v...)
}

func PrintlnWarning(v ...interface{}) {
	print2(color.Yellow, v...)
}

func PrintErr(v ...interface{}) {
	print2(color.FgLightRed, v...)
}

func Fatal(v ...interface{}) {
	log.Fatal(v...)
}