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
	proxy.Verbose = true
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
