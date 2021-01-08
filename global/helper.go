package global

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/tidwall/gjson"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func FormatJsonpResponse(b []byte, prefix string, isConvertStr bool) gjson.Result {
	r := string(b)
	if isConvertStr {
		r = mahonia.NewDecoder("gbk").ConvertString(r)
	}
	r = strings.TrimSpace(r)
	if prefix != "" {
		//这里针对http连接 自动提取jsonp的callback
		if strings.HasPrefix(prefix, "http") {
			pUrl, err := url.Parse(prefix)
			if err == nil {
				prefix = pUrl.Query().Get("callback")
			}
		}
		r = strings.TrimPrefix(r, prefix)
	}
	if strings.HasSuffix(r, ")") || strings.HasPrefix(r, "(") {
		r = strings.TrimRight(r, ";")
		r = strings.TrimLeft(r, `(`)
		r = strings.TrimRight(r, ")")
	}
	return gjson.Parse(r)
}

func UnixMilli() int64 {
	return time.Now().UnixNano() / 1e6
}

func GenerateRangeNum(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Int63n(max - min) + min
	return randNum
}

func Hour2Unix(hour string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFormatStr, time.Now().Format(DateFormatStr) + " " + hour, time.Local)
}

func Md5(s string) string {
	data := []byte(s)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func Json2Map(j string) map[string]interface{} {
	r := make(map[string]interface{})
	_ = json.Unmarshal([]byte(j), &r)
	return r
}


func RandFloats(min, max float64, n int) float64 {
	rand.Seed(time.Now().UnixNano())
	res := min + rand.Float64() * (max - min)
	res, _ =  strconv.ParseFloat(fmt.Sprintf("%."+strconv.Itoa(n)+"f", res), 64)
	return res
}