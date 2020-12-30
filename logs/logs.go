package logs

import (
	"github.com/gookit/color"
	"log"
)

//这个包用来统一的日志输出处理
//目前只做简单两个方法 后续根据具体需要在这里增加日志操作
func Println(v ...interface{}) {
	log.Println(v...)
}

func print2(color2 color.Color, v ...interface{}) {
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