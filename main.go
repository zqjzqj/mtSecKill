package main

import (
	"bufio"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/zqijzqj/mtSecKill/global"
	"github.com/zqijzqj/mtSecKill/logs"
	"github.com/zqijzqj/mtSecKill/secKill"
)

var skuId = flag.String("sku", "100012043978", "茅台商品ID")
var num = flag.Int("num", 2, "商品数量")
var works = flag.Int("works", 6, "并发数")
var start = flag.String("time", "09:59:59.500", "开始时间---不带日期")	// 2.18 10:00预约 12:00抢购
var brwoserPath = flag.String("execPath", "/usr/bin/google-chrome", "浏览器执行路径，路径不能有空格")

var eid = flag.String("eid", "", "如果不传入，可自动获取，对于无法获取的用户可手动传入参数")
var fp = flag.String("fp", "", "如果不传入，可自动获取，对于无法获取的用户可手动传入参数")
var payPwd = flag.String("payPwd", "", "支付密码 可不填")
var isFileLog = flag.Bool("isFileLog", true, "是否使用文件记录日志")

func init() {
	flag.StringVar(&global.PushToken, "token", "", "一键推送微信提醒令牌 可不填")

	flag.Parse()
}

func main() {
	var err error

	// 是否打出文件日志
	if *isFileLog {
		logs.AllowFileLogs()
	}

	// 浏览器路径
	execPath := ""
	if *brwoserPath != "" {
		execPath = *brwoserPath
	}
RE:
	// 新建jd秒杀对象
	jdSecKill := secKill.NewJdSecKill(execPath, *skuId, *num, *works)
	jdSecKill.StartTime, err = global.Hour2Unix(*start)
	if err != nil {
		logs.Fatal("开始时间初始化失败", err)
	}

	// 初始化支付密码
	jdSecKill.PayPwd = *payPwd

	// ？？ eid & fp
	if *eid != "" {
		if *fp == "" {
			logs.Fatal("请传入fp参数")
		}
		jdSecKill.SetEid(*eid)
	}
	if *fp != "" {
		if *eid == "" {
			logs.Fatal("请传入eid参数")
		}
		jdSecKill.SetFp(*fp)
	}

	// 如果超过了秒杀时间，等明天
	if jdSecKill.StartTime.Unix() < time.Now().Unix() {
		jdSecKill.StartTime = jdSecKill.StartTime.AddDate(0, 0, 1)
	}
	// 同步jd服务器时间
	jdSecKill.SyncJdTime()
	logs.PrintlnInfo("开始执行时间为：", jdSecKill.StartTime.Format(global.DateTimeFormatStr))

	// 运行自动化chrome
	err = jdSecKill.Run()
	if err != nil {
		if strings.Contains(err.Error(), "exec") {
			logs.PrintlnInfo("默认浏览器执行路径未找到，" + execPath + "  请重新输入：")
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
