package main

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/tidwall/gjson"
	"github.com/zqijzqj/mtSecKill/chromedpEngine"
	"github.com/zqijzqj/mtSecKill/logs"
	"strings"
	"time"
)

func main() {
	c, cc := chromedpEngine.NewExecRemoteCtx("ws://127.0.0.1:9222/devtools/page/73708D2E9D2F674696E09367368A6AF2")
	defer cc()
	err := chromedp.Run(c, chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch e := ev.(type) {
				case *network.EventResponseReceived:
					go func() {
						if strings.Contains(e.Response.URL, "AsyncUpdateCart.do") {
							b, _ := network.GetResponseBody(e.RequestID).Do(ctx)
							r := gjson.ParseBytes(b)

							logs.PrintlnInfo("当前时间戳：", r.Get("globalData").Get("currentTime").Int())
						}
					}()
				}
			})
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 3; i++ {
				go func() {
					tid, _ := target.CreateTarget("https://cart.taobao.com/cart.htm").Do(ctx)
					cctx, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(tid))
					_ = chromedp.Run(cctx, chromedp.Tasks{
						chromedp.ActionFunc(func(ctx context.Context) error {
							var jNodes []*cdp.Node
							_ = chromedp.Nodes("#J_OrderList div", &jNodes).Do(ctx)
							for _, n := range jNodes {
								nHtml, _ := dom.GetOuterHTML().WithNodeID(n.NodeID).Do(ctx)
								if strings.Contains(nHtml, "id=20739895092") {
									var jNodes2 []*cdp.Node
									_ = chromedp.Nodes("#" + n.AttributeValue("id") + " li input", &jNodes2).Do(ctx)
									for _, n2 := range jNodes2 {
										if n2.AttributeValue("type") == "checkbox" {
											err := chromedp.Click(`document.querySelector("[for='`+n2.AttributeValue("id")+`']")`, chromedp.ByJSPath).Do(ctx)
											if err != nil {
												logs.PrintErr(err)
											}
											logs.Println(n2.AttributeValue("checked"), "====",n2.AttributeValue("value"), "========", n2.AttributeValue("id"))
										}
									}
									break
								}
								logs.PrintlnInfo(n.AttributeValue("id"))
							}
							_ = chromedp.Sleep(2 * time.Second).Do(ctx)
							_ = chromedp.Click("J_SmallSubmit", chromedp.ByID).Do(ctx)
							return nil
						}),
					})
				}()
			}

		logs.Println("OK")
			return nil
		}),
		chromedp.Sleep(1000 * time.Second),
	})
	if err != nil {
		logs.PrintErr(err)
	}
}
