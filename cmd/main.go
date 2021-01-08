package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/zqijzqj/mtSecKill/global"
	"github.com/zqijzqj/mtSecKill/logs"
	"github.com/zqijzqj/mtSecKill/secKill"
	"os"
	"strings"
	"time"
)

var skuId = flag.String("sku", "", "茅台商品ID")
var num = flag.Int("num", 2, "茅台商品ID")
var works = flag.Int("works", 0, "并发数")
var start = flag.String("time", "", "开始时间---不带日期")
var sType = flag.Int("sType", 0, "秒杀类型")
var browserPath = flag.String("execPath", "", "浏览器执行路径，路径不能有空格")
func init() {
	flag.Parse()
}

func main() {
	if *sType == 0 {
		logs.PrintlnInfo("请输入运行类型：1京东，2天猫")
		_, err := fmt.Scan(sType)
		if err != nil {
			logs.Fatal("参数接收错误：", err)
		}
	}

	if *sType == 1 {
		//设置京东的默认sku与时间
		if *skuId == "" {
			*skuId = "100012043978"
		}
		if *start == "" {
			*start = "09:59:58"
		}
		jd()
	} else {
		if *skuId == "" {
			*skuId = "20739895092"//"596118126934"//
		}
		if *start == "" {
			*start = "19:59:58"
		}
		tm()
	}
}

func jd() {
	var err error
	execPath := ""
	if *browserPath != "" {
		execPath = *browserPath
	}
	RE:
	jdSecKill := secKill.NewJdSecKill(execPath, *skuId, *num, *works)
	jdSecKill.StartTime, err = global.Hour2Unix(*start)
	if err != nil {
		logs.Fatal("开始时间初始化失败", err)
	}

	if jdSecKill.StartTime.Unix() < time.Now().Unix() {
		jdSecKill.StartTime = jdSecKill.StartTime.AddDate(0, 0, 1)
	}
	jdSecKill.SyncJdTime()
	logs.PrintlnInfo("开始执行时间为：", jdSecKill.StartTime.Format(global.DateTimeFormatStr))

	err = jdSecKill.Run()
	if err != nil {
		if strings.Contains(err.Error(), "exec") {
			logs.PrintlnInfo("默认浏览器执行路径未找到，"+execPath+"  请重新输入：")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				execPath = scanner.Text()
				if execPath != "" {
					break
				}
			}
			goto RE
		}
		logs.Fatal(err)
	}
}

func tm() {
	var err error
	execPath := ""
	RE:
	tmSecKill := secKill.NewTmSecKill(execPath, *skuId, *num, *works)
	tmSecKill.StartTime, err = global.Hour2Unix(*start)
	if tmSecKill.StartTime.Unix() < time.Now().Unix() {
		tmSecKill.StartTime = tmSecKill.StartTime.AddDate(0, 0, 1)
	}
	logs.PrintlnInfo("开始执行时间为：", tmSecKill.StartTime.Format(global.DateTimeFormatStr))

	err = tmSecKill.Run()
	if err != nil {
		if strings.Contains(err.Error(), "exec") {
			logs.PrintlnInfo("默认浏览器执行路径未找到，"+execPath+"  请重新输入：")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				execPath = scanner.Text()
				if execPath != "" {
					break
				}
			}
			goto RE
		}
		logs.Fatal(err)
	}
}