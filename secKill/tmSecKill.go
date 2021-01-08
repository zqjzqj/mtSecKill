package secKill

import (
	"context"
	"errors"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
	"github.com/zqijzqj/mtSecKill/chromedpEngine"
	"github.com/zqijzqj/mtSecKill/global"
	"github.com/zqijzqj/mtSecKill/logs"
	"strconv"
	"strings"
	"sync"
	"time"
)

type tmSecKill struct {
	ctx *ContextStruct
	bCtx context.Context
	bWorksCtx []context.Context
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
	IsSyncTime bool
}

func NewTmSecKill(execPath string, skuId string, num, works int) *tmSecKill {
	if works <= 0 {
		works = 2
	}
	logs.PrintlnInfo("运行线程数：", works)
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
		IsSyncTime: false,
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
					if strings.Contains(e.Response.URL, "newlogin/qrcode/query.do") {
						b, err := network.GetResponseBody(e.RequestID).Do(ctx)
						if err == nil {
							r := global.FormatJsonpResponse(b, e.Response.URL, false)
							if r.Get("content").Get("data").Get("qrCodeStatus").String() == "CONFIRMED" {
								tsk.isLogin = true
							}
						}
					}

					if strings.Contains(e.Response.URL, "AsyncUpdateCart.do") {
						b, _ := network.GetResponseBody(e.RequestID).Do(ctx)
						r := gjson.ParseBytes(b)
						tbCurrent := r.Get("globalData").Get("currentTime").Int()
						if tbCurrent > 0 {
							tsk.DiffTime = global.UnixMilli() - tbCurrent
							tsk.IsSyncTime = true
							tbTime := time.Unix(tbCurrent / 1e3, 0)
							logs.PrintlnInfo("淘宝时间戳：", tbCurrent, tbTime.Format(global.DateTimeFormatStr))
							logs.PrintlnInfo("服务器与本地时间差为: ", tsk.DiffTime, "ms")
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
				select {
				case <-tsk.ctx.Ctx.Done():
					logs.PrintErr("浏览器被关闭，退出进程")
					return nil
				case <-tsk.bCtx.Done():
					logs.PrintErr("浏览器被关闭，退出进程")
					return nil
				default:
				}
				if tsk.isLogin {
					logs.PrintlnSuccess("登陆成功........")
					break
				}
			}
			tsk.SelectSkuCat(ctx)
			return nil
		}),
		tsk.WaitStart(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, c := range tsk.bWorksCtx {
				go func(ctx2 context.Context) {
					for {
						logs.PrintlnInfo("开始提交订单............")
						select {
						case <-tsk.ctx.Ctx.Done():
							logs.PrintErr("浏览器被关闭，退出进程")
							return
						case <-tsk.bCtx.Done():
							logs.PrintErr("浏览器被关闭，退出进程")
							return
						default:
						}
						if err := tsk.SubmitOrder(ctx2); err != nil {
							tsk.SelectSkuCat(ctx2)
							logs.PrintErr("订单提交错误，等待重试")
							continue
						}
						break
					}
				}(c)
			}
			select {
			case <-tsk.IsOkChan:
				logs.PrintlnInfo("抢购成功。。。10s后关闭进程...")
				_ = chromedp.Sleep(10 * time.Second).Do(ctx)
			case <-tsk.ctx.Ctx.Done():
			case <-tsk.bCtx.Done():
			}
			return nil
		}),
	})
}

func (tsk *tmSecKill) WaitStart() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		st := tsk.StartTime.UnixNano() / 1e6

		logs.PrintlnInfo("等待时间同步，如没有自动同步时间，可手动在购物车页面取消/选中对应的sku商品，期间请勿关闭浏览器")
		for {
			select {
			case <-tsk.ctx.Ctx.Done():
				logs.PrintErr("浏览器被关闭，退出进程")
				return nil
			case <-tsk.bCtx.Done():
				logs.PrintErr("浏览器被关闭，退出进程")
				return nil
			default:
			}
			if tsk.IsSyncTime {
				break
			}
		}

		 wg := sync.WaitGroup{}
		for i := 0; i < tsk.Works; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tid, err := target.CreateTarget("https://www.taobao.com").Do(ctx)
				if err == nil {
					c, _ := chromedp.NewContext(tsk.bCtx, chromedp.WithTargetID(tid))
					_ = chromedp.Run(c, chromedp.Tasks{
						chromedp.ActionFunc(func(ctx context.Context) error {
							logs.PrintlnInfo("打开新的抢购标签.....")
							tsk.mu.Lock()
							tsk.bWorksCtx = append(tsk.bWorksCtx, ctx)
							tsk.mu.Unlock()
							tsk.SelectSkuCat(ctx)
							return nil
						}),
					})
				}
			}()
		}
		wg.Wait()
		logs.PrintlnInfo("等待时间到达"+tsk.StartTime.Format(global.DateTimeFormatStr)+"...... 请勿关闭浏览器")
		return nil
		for {
			select {
			case <-tsk.ctx.Ctx.Done():
				logs.PrintErr("浏览器被关闭，退出进程")
				return nil
			case <-tsk.bCtx.Done():
				logs.PrintErr("浏览器被关闭，退出进程")
				return nil
			default:
			}

			if global.UnixMilli() - tsk.DiffTime >= st {
				logs.PrintlnInfo("时间到达。。。。开始执行")
				break
			}
		}
		return nil
	}
}

//选中购物车中对应的商品
func (tsk *tmSecKill) SelectSkuCat(ctx context.Context) {
	_, _, _, _ = page.Navigate("https://cart.taobao.com/cart.htm").WithReferrer("https://www.taobao.com/").Do(ctx)
	var jNodes []*cdp.Node
	_ = chromedp.Nodes("#J_OrderList div", &jNodes).Do(ctx)
	step := 0
	for _, n := range jNodes {
		nHtml, _ := dom.GetOuterHTML().WithNodeID(n.NodeID).Do(ctx)
		if strings.Contains(nHtml, "id=" + tsk.SkuId) {
			var jNodes2 []*cdp.Node
			_ = chromedp.Nodes("#" + n.AttributeValue("id") + " li input", &jNodes2).Do(ctx)
			for _, n2 := range jNodes2 {
				//选中对应sku的商品
				if n2.AttributeValue("type") == "checkbox" {
					err := chromedp.Click(`document.querySelector("[for='`+n2.AttributeValue("id")+`']")`, chromedp.ByJSPath).Do(ctx)
					if err != nil {
						logs.PrintErr(err)
					}
					step++
					logs.Println(n2.AttributeValue("checked"), "====",n2.AttributeValue("value"), "========", n2.AttributeValue("id"))
				}

				//设置数量
				if n2.AttributeValue("type") == "text" {
					if n2.AttributeValue("value") != strconv.Itoa(tsk.SecKillNum) {
						//这里如果不使用输入到input而是直接修改input.value的话 是不会触发js请求的 也就是说在淘宝服务器上并没有成功的修改数量
						_ = dom.SetAttributeValue(n2.NodeID, "id", "mtNum").Do(ctx)
						_ = chromedp.SendKeys("mtNum", strconv.Itoa(tsk.SecKillNum), chromedp.ByID).Do(ctx)
					}
					step++
				}
				if step == 2 {
					break
				}
			}
			break
		}
	}
	logs.PrintlnInfo("选中商品完成.......")
}

func (tsk *tmSecKill) SubmitOrder(ctx context.Context) error {
	for {
		select {
		case <-tsk.ctx.Ctx.Done():
			logs.PrintErr("浏览器被关闭，退出进程")
			return nil
		case <-tsk.bCtx.Done():
			logs.PrintErr("浏览器被关闭，退出进程")
			return nil
		default:
		}
		logs.PrintlnInfo("准备结算...........")
		var JGOValue string
		for {
			//方式点击过快 淘宝js还没有移除这个class
			_ = chromedp.AttributeValue("J_Go", "class", &JGOValue, nil, chromedp.ByID).Do(ctx)
			if !strings.Contains(JGOValue, "submit-btn-disabled") {
				break
			}
		}
		err := chromedp.Click("J_Go", chromedp.ByID).Do(ctx)
		if err == nil {
			break
		}
	}
	logs.PrintlnInfo("等待跳转结算页面.....")
	ch, cc := chromedpEngine.WaitDocumentUpdated(ctx)
	defer cc()
	<-ch
	tInfo, err := target.GetTargetInfo().Do(ctx)
	if err != nil {
		return err
	}
	logs.PrintlnInfo("跳转页面URL：", tInfo.URL)
	if !strings.Contains(tInfo.URL, "order/confirm_order.htm") {
		return errors.New("提交订单错误")
	}

	root, _ := dom.GetDocument().Do(ctx)
	html, _ := dom.GetOuterHTML().WithNodeID(root.NodeID).Do(ctx)
	if !strings.Contains(html, "go-btn") {
		return errors.New("未找到提交支付按钮.....")
	}
	var subNodes []*cdp.Node
	err = chromedp.Nodes(".go-btn", &subNodes).Do(ctx)
	if err != nil {
		return err
	}
	if len(subNodes) == 0 {
		return errors.New("未找到提交支付按钮.............")
	}
	logs.PrintlnInfo("准备提交支付...........")
	isOk := false
	for _, n := range subNodes {
		if n.AttributeValue("title") == "提交订单" {
			err = chromedp.MouseClickNode(n).Do(ctx)
			if err != nil {
				return err
			}
			isOk = true
		}
	}
	if !isOk {
		return errors.New("订单提交失败")
	}
	tsk.IsOk = true
	tsk.IsOkChan <- struct{}{}
	return nil
}