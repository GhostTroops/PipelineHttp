package PipelineHttp

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

// for http2
type tConn struct {
	net.Conn
	T io.Writer // receives everything that is read from Conn
}

func (w *tConn) Read(b []byte) (n int, err error) {
	n, err = w.Conn.Read(b)
	w.T.Write(b)
	return
}

func (r *PipelineHttp) DoUrl4Http2(szUrl string, framerCbk func(*http2.Frame, *http.Response, *error)) {
	r.UseHttp2 = true
	r.Client = &http.Client{Transport: r.GetTransport4http2()}
	res, err := r.Client.Get(szUrl)
	if err != nil {
		framerCbk(nil, nil, &err)
		return
	}
	defer res.Body.Close() // res.Write(os.Stdout)
	r.httpDataIO(func(frame *http2.Frame, e *error) {
		framerCbk(frame, res, e)
	})
}

// dialT returns a connection that writes everything that is read to w.
func (r *PipelineHttp) dialT(w io.Writer) func(network, addr string, cfg *tls.Config) (net.Conn, error) {
	return func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		conn, err := tls.Dial(network, addr, cfg)
		return &tConn{conn, w}, err
	}
}
func (r *PipelineHttp) httpDataIO(framerCbk func(*http2.Frame, *error)) {
	framer := http2.NewFramer(ioutil.Discard, r.Buf)
	for {
		f, err := framer.ReadFrame()
		framerCbk(&f, &err)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if nil != err {
			log.Println(err, framer.ErrorDetail())
		}
		//switch err.(type) {
		//case nil:
		//	log.Println(f)
		//case http2.ConnectionError:
		//	// Ignore. There will be many errors of type "PROTOCOL_ERROR, DATA
		//	// frame with stream ID 0". Presumably we are abusing the framer.
		//default:
		//	log.Println(err, framer.ErrorDetail())
		//}
	}
}

// 传输对象配置
func (r *PipelineHttp) GetTransport4http2() *http2.Transport {
	if r.UseHttp2 {
		tr := &http2.Transport{DialTLS: r.dialT(r.Buf)}
		return tr
	}
	return nil
}
