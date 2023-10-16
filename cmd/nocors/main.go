// nocors is a reverse proxy that intercepts all the [preflight] requests
// it receives, and generates a response that always allows the operation.
// [preflight]:https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
		os.Exit(1)
	}
	args := flag.Args()
	src := args[0]
	dst := args[1]
	rv := newNoCorsReverseProxy(dst)
	err := http.ListenAndServe(src, rv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server finished, error: %v", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, "usage: nocors <listen_address> <dest_host>\nexample: nocors localhost:8080 localhost:9090")
}

type noCORSReverseProxy struct {
	*httputil.ReverseProxy
}

func (p *noCORSReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// We always need to add the "Access-Control-Allow-Origin" header.
	allowOrigin := req.Header["Origin"]
	for _, o := range allowOrigin {
		rw.Header().Add("Access-Control-Allow-Origin", o)
	}
	// If the request is not a preflight we forward it upstream.
	if req.Method != http.MethodOptions {
		p.ReverseProxy.ServeHTTP(rw, req)
		return
	}
	// The request is for an OPTIONS method, if it doesn't have the header:
	// "Access-Control-Request-Method" it is not a preflight request, otherwise
	// it is.
	if _, ok := req.Header["Access-Control-Request-Method"]; !ok {
		p.ReverseProxy.ServeHTTP(rw, req)
		return
	}
	// It's a preflight request, generate an "allow" response.
	allowMethods := req.Header["Access-Control-Request-Method"]
	allowHeaders := req.Header["Access-Control-Request-Headers"]

	for _, m := range allowMethods {
		rw.Header().Add("Access-Control-Allow-Methods", m)
	}
	for _, h := range allowHeaders {
		rw.Header().Add("Access-Control-Allow-Headers", h)
	}

	rw.WriteHeader(http.StatusNoContent)
}

func newNoCorsReverseProxy(host string) *noCORSReverseProxy {
	director := func(req *http.Request) {
		req.URL.Host = host
		req.URL.Scheme = "http"
		// Borrowed from the std lib function NewSingleHostReverseProxy.
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}
	rp := &httputil.ReverseProxy{Director: director}
	return &noCORSReverseProxy{rp}
}
