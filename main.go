package main

import (
	"flag"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"regexp"
)

// HTTP/HTTPS proxy for debugging
func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func (req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Println(req.Method, req.URL)
		return req, nil
	})

	//proxy.Verbose = true
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
