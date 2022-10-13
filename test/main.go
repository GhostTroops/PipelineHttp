package main

import (
	_ "embed"
	xxx "github.com/hktalent/PipelineHttp"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//go:embed testUrl.txt
var szPath string

//go:embed testUrls.txt
var szUrls string

/*
cat $HOME/MyWork/mybugbounty/bak/fingerprint.json|grep -v '"delete":true'|jq ".probeList[].url"|sort -u|sed 's/"//g'>$HOME/MyWork/PipelineHttp/test/testUrl.txt
*/
func main() {
	a := strings.Split(szUrls, "\n")
	c01 := make(chan struct{}, 4)
	x := strings.Split(szPath, "\n")
	var wg sync.WaitGroup
	n009 := time.Now().UnixMilli()
	for _, i := range a {
		if "" == i {
			continue
		}
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
			x1.ErrLimit = 9999999
			defer x1.Close()

			log.Println("start ", s1)
			//x1.Client = x1.GetClient4Http3()
			//x1.DoDirs4Http2(s1, x, 128, func(resp *http.Response, err error, szU string) {
			x1.DoDirs(s1, x, 128, func(resp *http.Response, err error, szU string) { // auto test http2.0 and use http2.0
				//if nil != err {
				//	log.Println(err)
				//}

				if nil != resp && 200 == resp.StatusCode {
					log.Printf("%d %s %s", resp.StatusCode, resp.Proto, szU)
				} else if nil != resp {
					//log.Printf("%d %s %s", resp.StatusCode, resp.Proto, szU)
				}
			})
			//time.Sleep(time.Second * 5)
			//x1.Close()

		}(i)
	}
	wg.Wait()

	log.Printf("use time: %d/%d sec\n", len(x), (time.Now().UnixMilli()-n009)/1000)
}
