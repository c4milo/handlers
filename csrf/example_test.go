package csrf_test

import (
	"fmt"
	"net/http"

	"github.com/c4milo/handlers/csrf"
)

func ExampleHandler() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	})

	rack := csrf.Handler(mux, csrf.UserID("user ID"), csrf.Secret("my secret!"), csrf.Domain("localhost"))

	http.ListenAndServe(":8080", rack)
}
