package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewReverseProxy(target string) *httputil.ReverseProxy {
	if target == "" {
		panic("reverse proxy target is empty")
	}

	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		panic(fmt.Sprintf("invalid reverse proxy target %q: %v", target, err))
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
	}

	return proxy
}
