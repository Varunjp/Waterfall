package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) *httputil.ReverseProxy {
	url,_ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.Host = url.Host
	}

	return proxy
}