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

	opts := []csrf.Option{
		csrf.WithUserID("user ID"),
		csrf.WithSecret("my secret!"),
	}

	rack := csrf.Handler(mux, opts...)

	http.ListenAndServe(":8080", rack)
}
