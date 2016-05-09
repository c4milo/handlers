package main

import (
	"fmt"
	"net/http"
)

func ExampleHandler() {
	mux := http.NewServeMux()

	mux = csrf.Handler(mux, Session(new(SessionImpl)), Secret("my secret!"), Domain("localhost"), Name("_csrf"))
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})
}
