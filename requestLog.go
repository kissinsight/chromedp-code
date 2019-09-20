package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

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
		//chromedp.Flag("window-size", "50,400"),
		chromedp.UserDataDir(dir),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// create a timeout
	taskCtx, cancel = context.WithTimeout(taskCtx, 10*time.Second)
	defer cancel()

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	// listen network event
	listenForNetworkEvent(taskCtx)

	chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate(`http://www.yahoo.com`),
	)

}

func listenForNetworkEvent(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventResponseReceived:
			resp := ev.Response
			log.Println("---------- Resp URL --------------\n", resp.URL)
			log.Printf("\nReceived headers: %s \n", resp.Headers)
			// if len(resp.Headers) != 0 {
			// 	log.Printf("\nReceived headers: %s \n", resp.Headers)
			// }
		}
	})
}
