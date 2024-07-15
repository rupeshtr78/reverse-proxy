package test

import (
	"fmt"
	"net/http"
	"os"
)

type CustomHandler struct{}

func (p *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("x_header_set", "xhs")
	r.Header.Set("x_custom_proxy", "true")
	r.Header.Set("x_custom_values", "rtr")
	r.Header.Add("x_custom_values", "rtr-2")
	r.Header.Write(os.Stdout)
	w.Header().Write(os.Stdout)

	r.Host = "0.0.0.0:1085"

	fmt.Fprintf(w, "Custon Handler Server Http %v, %v\n", r.Host, w.Header().Get("X_custom_values"))
}
func ProxyMux() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ServerMux Handle Func %v\n", r.URL.String())
	})

	mux.Handle("/custom", &CustomHandler{})

	http.ListenAndServe(":1080", mux)

}
