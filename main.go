package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/fatih/color"
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
	Id   int64
	R    io.ReadCloser
}

var showBody = flag.Bool("body", false, "display request and response bodies")
var showHeader = flag.Bool("header", false, "display headers")

var reqBodyColor = color.New(color.FgMagenta).SprintFunc()
var respBodyColor = color.New(color.FgBlue).SprintFunc()
var urlColor = color.New(color.FgYellow).SprintFunc()

func (c *BufferCloser) Read(b []byte) (n int, err error) {
	n, err = c.R.Read(b)
	// log.Printf("%s %s n=%d err=%v", c.tipe, c.Id, n, err)
	if n > 0 {
		sep := "\n^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n"
		body := string(b[:n])
		if c.tipe == REQ {
			body = reqBodyColor(body)
		} else {
			body = respBodyColor(body)
		}
		log.Printf("[%d] %s body: \n%s"+sep, c.Id, c.tipe, body)
	}
	return n, err
}

func (c BufferCloser) Close() error {
	//log.Printf("[%d] Close: %s\n", c.Id, c.tipe)
	return c.R.Close()
}

func printHeader(header http.Header) {
	var headerDisplay string
	for k, v := range header {
		headerDisplay = headerDisplay + fmt.Sprintf("    %s: %s\n", k, v)
	}
	log.Printf("headers:\n%s", headerDisplay)
}

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Print(err)
	}

	log.Printf("[%d] %s --> %s %s", ctx.Session, ip, req.Method, urlColor(req.URL))

	if *showHeader {
		printHeader(req.Header)
	}
	if *showBody {
		req.Body = &BufferCloser{REQ, ctx.Session, req.Body}
	}
	return req, nil
}

func handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	log.Printf("[%d] <-- %d %s", ctx.Session, resp.StatusCode, ctx.Req.URL.String())

	location := resp.Header.Get("Location")
	if location != "" {
		log.Printf("Location: %s", location)
	}

	if *showHeader {
		printHeader(resp.Header)
	}
	if *showBody {
		resp.Body = &BufferCloser{RESP, ctx.Session, resp.Body}
	}

	return resp
}

// HTTP/HTTPS proxy for debugging
func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	hostmatch := flag.String("hostmatch", "^.*$", "hosts to trace (regexp pattern)")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	log.SetFlags(log.Lmicroseconds)

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(*hostmatch))).
		HandleConnect(goproxy.AlwaysMitm)

	proxy.OnRequest().DoFunc(handleRequest)

	proxy.OnResponse().DoFunc(handleResponse)

	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
