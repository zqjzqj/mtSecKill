package global

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/giant-stone/go/ghttp"
	"github.com/giant-stone/go/gutil"

	"github.com/zqijzqj/mtSecKill/logs"
)

func NotifyUser(v ...interface{}) {
	if PushToken == "" {
		return
	}

	chunks := []string{}
	for _, chunk := range v {
		chunks = append(chunks, fmt.Sprintf("%v", chunk))
	}
	msg := strings.Join(chunks, " ")

	fullurl := "https://sre24.com/api/v1/push"
	rqBody, _ := json.Marshal(&map[string]interface{}{
		"token": PushToken,
		"msg": msg,
	})
	rq := ghttp.New().
		SetTimeout(time.Second * 3).
		SetRequestMethod("POST").
		SetUri(fullurl).
		SetPostBody(&rqBody)
	err := rq.Send()
	if gutil.CheckErr(err) {
		logs.PrintErr("notify user fail", string(rq.RespBody))
	}
}
