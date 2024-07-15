package reverseproxy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reverseproxy/pkg/logger"
)

var log = logger.NewLogger(os.Stdout, "reverseproxy", slog.LevelDebug)

type ReverseProxy struct {
	Route *Route
	Proxy *httputil.ReverseProxy
}

type Config struct {
	Routes []Route `yaml:"routes"`
}
type Route struct {
	Name       string `yaml:"name omitempty=false"`
	ListenPort int    `yaml:"listenport omitempty=false"`
	Protocol   string `yaml:"protocol omitempty=false"`
	Target     Target `yaml:"target omitempty=false"`
}

type Target struct {
	Name     string `yaml:"name omitempty=false"`
	Protocol string `yaml:"protocol omitempty=false"`
	Host     string `yaml:"host omitempty=false"`
	Port     int    `yaml:"port omitempty=false"`
}

func NewReverseProxy(route *Route) (*ReverseProxy, error) {

	target := route.Target
	url, err := GetTargetURL(target)
	if err != nil {
		log.Info("Error parsing target url", err)
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
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
