package PipelineHttp

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type PipelineHttp struct {
	Timeout               time.Duration
	KeepAlive             time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	Client                *http.Client
	Ctx                   context.Context
	StopAll               context.CancelFunc
	IsClosed              bool
	ErrLimit              int // 错误次数统计，失败就停止
	ErrCount              int // 错误次数统计，失败就停止
}

func NewPipelineHttp() *PipelineHttp {
	x1 := &PipelineHttp{
		Timeout:               30 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   3000,
		ErrLimit:              10, // 相同目标，累计错误10次就退出
		ErrCount:              0,
		IsClosed:              false,
	}
	x1.SetCtx(context.Background())
	//http.DefaultTransport.(*http.Transport).MaxIdleConns = x1.MaxIdleConns
	//http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = x1.MaxIdleConnsPerHost
	return x1
}

/*
	if err != nil {
	    return nil, err
	}

	sa := &syscall.SockaddrInet4{
	    Port: tcpAddr.Port,
	    Addr: [4]byte{tcpAddr.IP[0], tcpAddr.IP[1], tcpAddr.IP[2], tcpAddr.IP[3]},
	}

fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)

	if err != nil {
	    return nil, err
	}

err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TOS, 128)

	if err != nil {
	    return nil, err
	}

err = syscall.Connect(fd, sa)

	if err != nil {
	    return nil, err
	}

file := os.NewFile(uintptr(fd), "")
conn, err := net.FileConn(file)

	if err != nil {
	    return nil, err
	}

return conn, nil
*/
func (r *PipelineHttp) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := (&net.Dialer{
		Timeout:   r.Timeout,
		KeepAlive: r.KeepAlive,
		//Control:   r.Control,
		//DualStack: true,
	}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	//tcpConn, ok := conn.(*net.TCPConn)
	//if !ok {
	//	err = errors.New("conn is not tcp")
	//	return nil, err
	//}
	//
	//f, err := tcpConn.File()
	//if err != nil {
	//	return nil, err
	//}
	//internet.ApplyInboundSocketOptions("tcp", f.Fd())

	return conn, nil
}
func (r *PipelineHttp) SetCtx(ctx context.Context) {
	r.Ctx, r.StopAll = context.WithCancel(ctx)
}

// https://github.com/golang/go/issues/23427
func (r *PipelineHttp) GetTransport() *http.Transport {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           r.Dial,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10},
		DisableKeepAlives:     false,
		MaxIdleConns:          r.MaxIdleConns,
		IdleConnTimeout:       r.IdleConnTimeout,
		TLSHandshakeTimeout:   r.TLSHandshakeTimeout,
		ExpectContinueTimeout: r.ExpectContinueTimeout,
		MaxIdleConnsPerHost:   r.MaxIdleConnsPerHost,
	}
	return tr
}

func (r *PipelineHttp) GetClient() *http.Client {
	c := &http.Client{
		Transport: r.GetTransport(),
		//CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//	return http.ErrUseLastResponse /* 不进入重定向 */
		//},
	}
	return c
}

func (r *PipelineHttp) DoGet(szUrl string, fnCbk func(resp *http.Response, err error, szU string)) {
	r.DoGetWithClient(nil, szUrl, "GET", nil, fnCbk)
}

func (r *PipelineHttp) DoGetWithClient(client *http.Client, szUrl string, method string, postBody io.Reader, fnCbk func(resp *http.Response, err error, szU string)) {
	if client == nil {
		if nil != r.Client {
			client = r.Client
		} else {
			client = r.GetClient()
			r.Client = client
		}
	} else {
		r.Client = client
	}
	req, err := http.NewRequest(method, szUrl, postBody)
	if nil == err {
		req.Header.Add("Connection", "keep-alive")
		req.Close = false
	} else {
		log.Println("http.NewRequest is error ", err)
		return
	}
	req = req.WithContext(r.Ctx)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close() // resp 可能为 nil，不能读取 Body
	}
	if nil != err {
		r.ErrCount++
	}
	if r.ErrCount >= r.ErrLimit {
		r.Close()
	}
	fnCbk(resp, err, szUrl)
}

func (r *PipelineHttp) Close() {
	r.IsClosed = true
	r.StopAll()
	if nil != r.Client {
		//r.Client
	}
	r.Client = nil
}

// more see test/main.go
func (r *PipelineHttp) DoDirs(szUrl string, dirs []string, nThread int, fnCbk func(resp *http.Response, err error, szU string)) {
	c02 := make(chan struct{}, nThread)
	defer close(c02)
	oUrl, err := url.Parse(szUrl)
	if nil != err {
		log.Printf("url.Parse is error: %v %s", err, szUrl)
		return
	}
	if "" == oUrl.Scheme {
		oUrl.Scheme = "http"
	}
	szUrl = oUrl.Scheme + "://" + oUrl.Host
	var wg sync.WaitGroup
	for _, j := range dirs {
		if r.IsClosed {
			return
		}
		select {
		case <-r.Ctx.Done():
			return
		default:
			{
				c02 <- struct{}{}
				wg.Add(1)
				go func(s2 string) {
					defer func() {
						<-c02
						wg.Done()
					}()
					select {
					case <-r.Ctx.Done():
						return
					default:
						{
							s2 = strings.TrimSpace(s2)
							if !strings.HasPrefix(s2, "/") {
								s2 = "/" + s2
							}
							szUrl001 := szUrl + s2
							r.DoGet(szUrl001, fnCbk)
							return
						}
					}
				}(j)
				continue
			}
		}
	}
	wg.Wait()
}
