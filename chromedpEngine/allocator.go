package chromedpEngine

import (
	"context"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/zqijzqj/mtSecKill/logs"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var DefaultOptions = []chromedp.ExecAllocatorOption{
	chromedp.Flag("headless", false),
	chromedp.Flag("hide-scrollbars", false),
	chromedp.Flag("mute-audio", true),
	chromedp.Flag("disable-infobars", true),
	chromedp.Flag("enable-automation", false),
	chromedp.Flag("start-maximized", true),

	chromedp.Flag("disable-default-apps", false),
	chromedp.Flag("no-sandbox", false),
	// 隐身模式启动
	//chromedp.Flag("incognito", true),
	chromedp.Flag("disable-extensions", false),
	chromedp.Flag("disable-plugins", false),
	chromedp.NoDefaultBrowserCheck,
	chromedp.NoFirstRun,
}

var globalCtx *GlobalBackgroundCtx = nil
var mu sync.Mutex
type GlobalBackgroundCtx struct {
	background context.Context
	Cancel context.CancelFunc
}

func GetGlobalCtx() context.Context {
	if globalCtx == nil {
		NewGlobalCtx()
	}
	return globalCtx.background
}


func NewGlobalCtx() {

	mu.Lock()
	defer mu.Unlock()
	if globalCtx != nil {
		return
	}
	c, cc := context.WithCancel(context.Background())
	globalCtx = &GlobalBackgroundCtx{
		background: c,
		Cancel:     cc,
	}
}

func CancelGlobalCtx() {
	mu.Lock()
	defer mu.Unlock()
	if globalCtx == nil {
		return
	}
	cc := globalCtx.Cancel
	cc()
	logs.PrintlnSuccess("cancel global ctx...")
	globalCtx = nil
	return
}

var UserAgent = []string{
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.75 Safari/537.36",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0",
	"Mozilla/5.0 (X11; U; Linux x86_64; zh-CN; rv:1.9.2.10) Gecko/20100922 Ubuntu/10.10 (maverick) Firefox/3.6.10",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 SE 2.X MetaSr 1.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36 QIHU 360SE",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; Trident/7.0; .NET CLR 2.0.50727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET4.0C; .NET4.0E; rv:11.0) like Gecko",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36 OPR/60.0.3255.84",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 UBrowser/6.2.4098.3 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.98 Safari/537.36 LBBROWSER",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.25 Safari/537.36 Core/1.70.3676.400 QQBrowser/10.4.3473.400",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36 Maxthon/5.2.7.2500",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.131 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.106 BIDUBrowser/8.7 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36 QIHU 360EE",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; WOW64; Trident/4.0; SLCC1)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; en; rv:1.9.1.7) Gecko/20091221 Firefox/3.5.7",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36",
}

func GetRandUserAgent() string {
	RE:
	al := len(UserAgent)
	if al > 1 {
		rand.Seed(time.Now().UnixNano())
		return  UserAgent[rand.Intn(al)]
	}
	goto RE
}

func AddDefaultOptions(option ...chromedp.ExecAllocatorOption) {
	DefaultOptions = append(DefaultOptions, option...)
}

func RequestByCookie(ctx context.Context, req *http.Request, isDisableRedirects bool) (*http.Response, error) {
	httpClient := &http.Client{}
	if isDisableRedirects {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	cookies, err := network.GetCookies().WithUrls([]string{req.URL.String()}).Do(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		req.AddCookie(&http.Cookie{
			Name:       c.Name,
			Value:      c.Value,
		})
	}
	return httpClient.Do(req)
}


func CreateOptions(opts ...chromedp.ExecAllocatorOption) []chromedp.ExecAllocatorOption {
	options := append(chromedp.DefaultExecAllocatorOptions[:], DefaultOptions...)
	options = append(options, opts...)
	return options
}

func WaitDocumentUpdated(ctx context.Context) (<-chan struct{}, context.CancelFunc) {
	ctxNew, ccNew := context.WithCancel(ctx)
	_ = dom.Enable().Do(ctxNew)
	_ = page.Enable().Do(ctxNew)
	ch := make(chan struct{}, 1)
	chromedp.ListenTarget(ctxNew, func(ev interface{}) {
		isUpdated := false
		switch ev.(type) {
		case *dom.EventDocumentUpdated:
			isUpdated = true
		case *dom.EventCharacterDataModified:
			isUpdated = true
		case *target.EventTargetInfoChanged:
			isUpdated = true
		case *target.EventTargetCreated:
			isUpdated = true
		default:
			return
		}
		if isUpdated {
			select {
			case <- ctxNew.Done():
			case ch <- struct{}{}:
			}
			close(ch)
			ccNew()
			return
		}
	})
	return ch, ccNew
}

func NewExecCtx(opts ...chromedp.ExecAllocatorOption) (context.Context, context.CancelFunc) {
	topC, _ := context.WithCancel(GetGlobalCtx())
	c, _ := chromedp.NewExecAllocator(topC, CreateOptions(opts...)...)
	ctx, cancel := chromedp.NewContext(c)
	return ctx, cancel
}

func NewExecRemoteCtx(remoteWs string, opts ...chromedp.ExecAllocatorOption) (context.Context, context.CancelFunc) {
	topC, _ := context.WithCancel(GetGlobalCtx())
	c, _ := chromedp.NewExecAllocator(topC, CreateOptions(opts...)...)
	c, _ = chromedp.NewRemoteAllocator(c, remoteWs)
	ctx, cancel := chromedp.NewContext(c)
	return ctx, cancel
}

func NewExecAllocator(tasks chromedp.Tasks, opts ...chromedp.ExecAllocatorOption) error {
	//超时设置
	topC, topCC := context.WithCancel(GetGlobalCtx())
	defer topCC()
	c, cc := chromedp.NewExecAllocator(topC, CreateOptions(opts...)...)
	defer cc()
	ctx, cancel := chromedp.NewContext(c)
	defer cancel()
	_ = log.Disable().Do(ctx)
	err := chromedp.Run(ctx, tasks)
	logs.PrintlnSuccess("chromedp.Run end...")
	if err != nil {
		return err
	}
	return nil
}

//阻塞浏览器方法
func WaitAction(wait sync.WaitGroup) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		wait.Add(1)
		wait.Wait()
		return nil
	}
}