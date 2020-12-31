package global

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

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