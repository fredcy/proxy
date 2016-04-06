package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"io"
	"log"
	"net"
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
	// log.Printf("%s %s n=%d err=%v", c.tipe, c.Id, n, err)
	if n > 0 {
		sep := "\n^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n"
		log.Printf("%s %s body: \n%s"+sep, c.tipe, c.Id, string(b[:n]))
	}
	return n, err
}

func (c BufferCloser) Close() error {
	log.Printf("Close: %s %s\n", c.tipe, c.Id)
	return c.R.Close()
}

func printHeader(header http.Header) {
	var headerDisplay string
	for k, v := range header {
		headerDisplay = headerDisplay + fmt.Sprintf("    %s: %s\n", k, v)
	}
	log.Printf("headers:\n%s", headerDisplay)
}

// HTTP/HTTPS proxy for debugging
func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	hostmatch := flag.String("hostmatch", "^.*$", "hosts to trace (regexp pattern)")
	verbose := flag.Bool("v", false, "verbose output")
	showBody := flag.Bool("body", false, "display request and response bodies")
	showHeader := flag.Bool("header", false, "display headers")
	flag.Parse()

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(*hostmatch))).
		HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		id := fmt.Sprintf("%v %v", req.Method, req.URL)

		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Print(err)
		}

		log.Printf("%s --> %s", ip, id)

		if *showHeader {
			printHeader(req.Header)
		}
		if *showBody {
			req.Body = &BufferCloser{REQ, id, req.Body}
		}
		return req, nil
	})

	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		id := fmt.Sprintf("%v", ctx.Req.URL.String())
		log.Printf("<-- %d %s", resp.StatusCode, id)

		location := resp.Header.Get("Location")
		if location != "" {
			log.Printf("Location: %s", location)
		}

		if *showHeader {
			printHeader(resp.Header)
		}
		if *showBody {
			resp.Body = &BufferCloser{RESP, id, resp.Body}
		}

		return resp
	})

	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
