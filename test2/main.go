package main

import (
	_ "embed"
	xxx "github.com/hktalent/PipelineHttp"
	"github.com/projectdiscovery/retryablehttp-go"
	"log"
	"net/url"
	"strings"
	"sync"
)

//go:embed testUrl.txt
var szPath string

//go:embed testUrls.txt
var szUrls string

/*
cp $HOME/MyWork/scan4all/brute/dicts/filedic.txt $HOME/MyWork/PipelineHttp/test2/testUrl.txt
cat $HOME/MyWork/mybugbounty/bak/fingerprint.json|grep -v '"delete":true'|jq ".probeList[].url"|sort -u|sed 's/"//g'>$HOME/MyWork/PipelineHttp/cmd/testUrl.txt
*/
func main() {
	a := strings.Split(szUrls, "\n")
	c01 := make(chan struct{}, 4)
	x := strings.Split(szPath, "\n")
	var wg sync.WaitGroup
	opts := retryablehttp.DefaultOptionsSpraying
	// opts := retryablehttp.DefaultOptionsSingle // use single options for single host
	client := retryablehttp.NewClient(opts)
	for _, i := range a {
		c01 <- struct{}{}
		oUrl, err := url.Parse(i)
		if nil != err {
			continue
		}
		if "" == oUrl.Scheme {
			oUrl.Scheme = "http"
		}
		i = oUrl.Scheme + "://" + oUrl.Host
		wg.Add(1)
		go func(s1 string) {
			defer func() {
				<-c01
				wg.Done()
			}()
			x1 := xxx.NewPipelineHttp()
			defer x1.Close()
			x1.ErrCount = 0
			x1.ErrLimit = len(x) + 1
			log.Println("start ", s1)
			for _, j := range x {
				resp, err := client.Get(s1 + j)
				if err == nil {
					defer resp.Body.Close()
					if nil != resp && 200 == resp.StatusCode {
						log.Printf("%d %s %s\n", resp.StatusCode, resp.Proto, s1+j)
					}
				}
			}
			//
			////x1.DoDirs4Http2(s1, x, 128, func(resp *http.Response, err error, szU string) {
			//x1.DoDirs(s1, x, 128, func(resp *http.Response, err error, szU string) {
			//	if nil != resp && 200 == resp.StatusCode {
			//		log.Printf("%d %s %s", resp.StatusCode, resp.Proto, szU)
			//	} else {
			//		//log.Printf("%d %s", resp.StatusCode, szU)
			//	}
			//})
			//time.Sleep(time.Second * 5)
			//x1.Close()

		}(i)
	}
	wg.Wait()
}
