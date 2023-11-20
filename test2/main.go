package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/hbakhtiyor/strsim"
	xxx "github.com/hktalent/PipelineHttp"
	jsoniter "github.com/json-iterator/go"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

//go:embed testUrl.txt
var szPath string

//go:embed testUrls.txt
var szUrls string

// 获取 Sha1
func GetSha1(a ...interface{}) string {
	h := sha1.New()
	//if isMap(a[0]) { // map嵌套map 确保顺序，相同数据map得到相同的sha1
	if data, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(a); nil == err {
		h.Write(data)
	} else {
		for _, x := range a {
			h.Write([]byte(fmt.Sprintf("%v", x)))
		}
	}
	bs := h.Sum(nil)
	return hex.EncodeToString(bs) // fmt.Sprintf("%x", bs)
}

/*
cp -f $HOME/MyWork/scan4all/brute/dicts/filedic.txt $HOME/MyWork/PipelineHttp/test2/testUrl.txt
cat $HOME/MyWork/mybugbounty/bak/fingerprint.json|grep -v '"delete":true'|jq ".probeList[].url"|sort -u|sed 's/"//g'>$HOME/MyWork/PipelineHttp/cmd/testUrl.txt
*/
func main() {
	a := strings.Split(szUrls, "\n")
	c01 := make(chan struct{}, 8)
	x := strings.Split(szPath, "\n")
	var wg sync.WaitGroup
	//opts := retryablehttp.DefaultOptionsSpraying
	// opts := retryablehttp.DefaultOptionsSingle // use single options for single host
	//client := retryablehttp.NewClient(opts)
	fok, err := os.OpenFile("url_ok.txt", os.O_APPEND|os.O_CREATE, os.ModePerm)
	if nil != err {
		return
	}
	defer fok.Close()
	for _, i := range a {
		c01 <- struct{}{}
		oUrl, err := url.Parse(i)
		if nil != err {
			continue
		}
		if "" == oUrl.Scheme {
			oUrl.Scheme = "http"
		}
		i = oUrl.Scheme + "://" + oUrl.Host + oUrl.Path
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
			//x1.Client = x1.GetClient4Http2()
			//x1.UseHttp2 = false
			//x1.TestHttp = false
			var szLastData string
			x1.DoDirs4Http2(s1, x, 8, func(resp *http.Response, err error, szU string) {
				if err == nil {
					defer resp.Body.Close()
					if nil != resp && 200 == resp.StatusCode {
						if data, err := io.ReadAll(resp.Body); nil == err {
							szCur := string(data)
							var nF float64 = 0
							if 0 < len(szLastData) {
								nF = strsim.Compare(szLastData, szCur)
								if 0.89 < nF {
									return
								}
							}
							szLastData = szCur
							szT := fmt.Sprintf("%d %s %d:%s %02f %s\n", resp.StatusCode, resp.Proto, len(data), GetSha1(data), nF, szU)
							fok.WriteString(szT)
							log.Printf("\n%s", szT)
						}
					} else {
						fmt.Printf("%d %s %s\r", resp.StatusCode, resp.Proto, szU)
					}
				}
			})
			//log.Println("start ", s1)
			//for _, j := range x {
			//	resp, err := client.Get(s1 + j)
			//	if err == nil {
			//		defer resp.Body.Close()
			//		if nil != resp && 200 == resp.StatusCode {
			//			log.Printf("%d %s %s\n", resp.StatusCode, resp.Proto, s1+j)
			//		}
			//	}
			//}
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
