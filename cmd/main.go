package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {

		defer wg.Done()
		for {
			PocHttpServer()
		}

	}()

	go func() {
		defer wg.Done()
		for {
			// SimpleProxyServer()
			ProxyMain()
		}
	}()

	wg.Wait()

}

func PocHttpServer() {

	listenPort := "0.0.0.0:1080"

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Get the full request URL
		requestURL := r.URL.String()
		fmt.Fprintf(w, "Simple Http Server %v\n %v\n", requestURL, r.URL.Port())
	}

	http.HandleFunc("/", handler)

	fmt.Println("Starting server on 0.0.0.1080")
	err := http.ListenAndServe(listenPort, nil)
	if err != nil {
		panic(err)
	}
}

func SimpleProxyServer() {
	var target = "http://0.0.0.0:1080"
	// var inUrl = "http://0.0.0.0:1080"
	targetUrl, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	fmt.Println("Starting proxy server on 0.0.0.0:82")
	err = http.ListenAndServe(":82", proxy)
	if err != nil {
		panic(err)
	}
}

func ProxyMain() {

	var target = "http://0.0.0.0:1080"
	targetUrl, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(targetUrl)
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Set("Routed-To", targetUrl.String())
			r.Header.Write(os.Stdout)
			return nil
		},
	}

	fmt.Println("Starting proxy server listening on 82")
	err = http.ListenAndServe(":82", proxy)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server listeting")

}
