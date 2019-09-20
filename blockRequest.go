package main

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserDataDir(dir),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// create a timeout
	taskCtx, cancel = context.WithTimeout(taskCtx, 1000*time.Second)
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	// listen network event
	chromedp.ListenTarget(taskCtx, BlockLoading(taskCtx))

	chromedp.Run(taskCtx,
		network.Enable(),
		fetch.Enable(),
		chromedp.Navigate(`https://www.yahoo.com`),
		chromedp.Sleep(10*time.Second),
	)
}

func BlockLoading(ctx context.Context) func(event interface{}) {
	return func(event interface{}) {
		switch ev := event.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)
        
        // if you want to block only image and css change below if block.
				if ev.ResourceType == network.ResourceTypeStylesheet || ev.ResourceType == network.ResourceTypeImage || ev.ResourceType == network.ResourceTypeMedia || ev.ResourceType == network.ResourceTypeFont || ev.ResourceType == network.ResourceTypeTextTrack || ev.ResourceType == network.ResourceTypeCSPViolationReport {
					fetch.FailRequest(ev.RequestID, network.ErrorReasonConnectionAborted).Do(ctx)
				} else {
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}
			}()
		}
	}
}
