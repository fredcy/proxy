package main

import (
	"bytes"
	"flag"
	"github.com/elazarl/goproxy"
	_ "io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type BufferCloser struct {
	bytes.Buffer
}
func (b BufferCloser) Close() error {
	return nil
}

// HTTP/HTTPS proxy for debugging
func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	hostmatch := flag.String("hostmatch", ".*", "hosts to trace (regexp pattern)")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(*hostmatch))).
		HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func (req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Println(req.Method, req.URL)
		var b BufferCloser
		n, err := b.ReadFrom(req.Body)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("request body (%s): %s\n", n, b.String())
			req.Body = b
		}
		return req, nil
	})

	proxy.Verbose = *verbose
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
