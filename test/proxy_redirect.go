package test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reverseproxy/pkg/logger"
	"sync"
)

var log = logger.NewLogger(os.Stdout, "proxy-redirect", slog.LevelDebug)

func ProxyRedirectMain() {
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
			ProxyRewrite()
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
		log.Error("failed parsing target url", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	log.Info("Starting proxy server on 0.0.0.0:82")
	err = http.ListenAndServe(":82", proxy)
	if err != nil {
		log.Error("error starting proxy server", err)
		os.Exit(1)
	}
}

func ProxyRewrite() {

	var target = "http://0.0.0.0:1080"
	targetUrl, err := url.Parse(target)
	if err != nil {
		log.Error("failed parsing target url", err)
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

	log.Info("Starting proxy server listening on 82")
	err = http.ListenAndServe(":82", proxy)
	if err != nil {
		panic(err)
	}

}

// err = api.ProxyServer(&route1)
// if err != nil {
// 	log.Fatal("Exting Proxy Server failed", err)
// }

// proxy, err := reverseproxy.NewReverseProxy(&route1)
// if err != nil {
// 	log.Error("Error creating proxy", err, route1.Name)
// }

// // http.Handle("/", proxy) // testing

// mux, err := reverseproxy.NewServeMux(&route1, proxy)
// if err != nil {
// 	log.Error("Error creating mux", err)
// }

// log.Info("Staring Porxy Server on port %s\n", route1.ListenPort)
// err = http.ListenAndServe(fmt.Sprintf(":%d", route1.ListenPort), mux)

// if err != nil {
// 	log.Error("Error starting proxy server")
// }
