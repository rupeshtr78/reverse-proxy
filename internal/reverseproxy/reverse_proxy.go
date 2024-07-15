package reverseproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy is a struct that holds a Route and a Proxy. It is used to proxy HTTP requests to a target URL.
type ReverseProxy struct {
	Route *Route
	Proxy *httputil.ReverseProxy
}

// NewReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
// The ReverseProxy is configured with the provided Route, which contains the target URL information.
// If there is an error parsing the target URL, an error is returned.
func NewReverseProxy(route *Route) (*ReverseProxy, error) {

	target := route.Target
	url, err := GetTargetURL(target)
	if err != nil {
		log.Info("Error parsing target url", err)
		return nil, err
	}

	// proxy := httputil.NewSingleHostReverseProxy(url)

	// Setup the reverse proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
			req.URL.Path = url.Path + req.URL.Path
		},
	}
	reverseProxy := &ReverseProxy{
		Route: route,
		Proxy: proxy,
	}

	return reverseProxy, nil
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("X-Forwarded-Host", r.Host)
	p.Proxy.ServeHTTP(w, r)
}

func GetTargetURL(target Target) (*url.URL, error) {
	urlString := fmt.Sprintf("%s://%s:%d", target.Protocol, target.Host, target.Port)
	targetUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	return targetUrl, nil
}

func NewServeMux(route *Route, handler http.Handler) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	return mux, nil

}
