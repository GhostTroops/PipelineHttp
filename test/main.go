package main

import (
	_ "embed"
	xxx "github.com/hktalent/PipelineHttp"
	"log"
	"net/http"
	"net/url"
	"strings"
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
		go func(s1 string) {
			defer func() {
				<-c01
			}()
			x1 := xxx.NewPipelineHttp()
			defer x1.Close()

			log.Println("start ", s1)
			x1.DoDirs(s1, x, 128, func(resp *http.Response, err error, szU string) {
				if nil != resp && 200 == resp.StatusCode {
					log.Printf("%d %s", resp.StatusCode, szU)
				}
			})

		}(i)
	}
}
