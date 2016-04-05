package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"io"
	"log"
	"net/http"
	"regexp"
)

const (
	REQ  string = "request"
	RESP        = "response"
)

type BufferCloser struct {
	tipe string
	Id   string
	R    io.ReadCloser
}

func (c *BufferCloser) Read(b []byte) (n int, err error) {
	n, err = c.R.Read(b)
	log.Printf("%s %s n=%d err=%v", c.tipe, c.Id, n, err)
	if n > 0 {
		sep := "\n^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n"
		log.Printf("%s %s body=%s"+sep, c.tipe, c.Id, string(b[:n]))
	}
	return n, err
}

func (c BufferCloser) Close() error {
	log.Printf("Close: %s %s\n", c.tipe, c.Id)
	return c.R.Close()
}

// HTTP/HTTPS proxy for debugging
func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	hostmatch := flag.String("hostmatch", "^.*$", "hosts to trace (regexp pattern)")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(*hostmatch))).
		HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		id := fmt.Sprintf("%v %v", req.Method, req.URL)
		req.Body = &BufferCloser{REQ, id, req.Body}
		return req, nil
	})

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		id := fmt.Sprintf("%v", ctx.Req.URL.String())
		resp.Body = &BufferCloser{RESP, id, resp.Body}
		return resp
	})

	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
