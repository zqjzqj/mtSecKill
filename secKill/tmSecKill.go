package secKill

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/zqijzqj/mtSecKill/chromedpEngine"
	"github.com/zqijzqj/mtSecKill/logs"
	"strings"
	"sync"
	"time"
)

type tmSecKill struct {
	ctx *ContextStruct
	bCtx context.Context
	bWorksCtx []*ContextStruct
	SecKillNum int
	isLogin bool
	isClose bool
	mu sync.Mutex
	userAgent string
	SkuId string
	Works int
	IsOkChan chan struct{}
	IsOk bool
	StartTime time.Time
	DiffTime int64
}

func NewTmSecKill(execPath string, skuId string, num, works int) *tmSecKill {
	if works < 0 {
		works = 1
	}
	tsk := &tmSecKill{
		ctx:       nil,
		bCtx:      nil,
		bWorksCtx: nil,
		SecKillNum:num,
		isLogin:   false,
		userAgent: chromedpEngine.GetRandUserAgent(),
		SkuId:     skuId,
		Works:     works,
		IsOkChan:make(chan struct{}, 1),
		IsOk:      false,
		DiffTime:  0,
		isClose:false,
	}
	c, cc := chromedpEngine.NewExecCtx(chromedp.ExecPath(execPath), chromedp.UserAgent(tsk.userAgent))
	tsk.ctx = NewContextStruct(c, cc, "")
	return tsk
}

func (tsk *tmSecKill) Stop() {
	tsk.mu.Lock()
	defer tsk.mu.Unlock()
	if tsk.isClose {
		return
	}
	tsk.isClose = true
	c := tsk.ctx.Cancel
	c()
	return
}

//初始化监听请求数据
func (tsk *tmSecKill) InitActionFunc() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		tsk.bCtx = ctx
		_ = network.Enable().Do(ctx)
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch e := ev.(type) {
			case *network.EventResponseReceived:
				go func() {
					if strings.Contains(e.Response.URL, "top-tmm.taobao.com/member/query_member_top.do") {
						b, err := network.GetResponseBody(e.RequestID).Do(ctx)
						if err == nil {
							r := FormatJdResponse(b, e.Response.URL, false)
							if r.Get("login").Bool() {
								tsk.isLogin = true
							}
						}
					}
				}()

			}
		})
		return nil
	}
}

func (tsk *tmSecKill) Run() error {
	return chromedp.Run(tsk.ctx.Ctx, chromedp.Tasks{
		tsk.InitActionFunc(),
		chromedp.Navigate("https://login.taobao.com/member/login.jhtml"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logs.PrintlnInfo("等待登陆......")
			for {
				if tsk.isLogin {
					logs.PrintlnSuccess("登陆成功........")
					break
				}
			}
			return nil
		}),
	})
}