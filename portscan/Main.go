package main

import (
	"fmt"
	xxx "github.com/hktalent/PipelineHttp"
	"net/http"
)

var ppe = xxx.NewPipelineHttp()

func main() {
	c2 := ppe.GetClient4Http2()
	var c1 = make(chan struct{}, 8)
	for i := 0; i < 65536; i++ {
		c1 <- struct{}{}
		fmt.Printf("start %d\r", i)
		go func(n int) {
			defer func() {
				<-c1
				ppe.ErrCount = 0
				ppe.ErrLimit = 100000000
			}()
			ppe.DoGetWithClient(c2, fmt.Sprintf("https://www.baidu.com:%d", n), "GET", nil, func(resp *http.Response, err error, szU string) {
				if nil != err {
					return
				}
				if nil != resp {
					fmt.Printf("%d %s %s\n", resp.StatusCode, resp.Proto, szU)
				}
			})
		}(i)
	}
}
