package csrf_test

import (
	"fmt"
	"net/http"

	"github.com/c4milo/handlers/csrf"
)

type MySession struct{}

func (s *MySession) ID() string {
	return "session ID"
}

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

	userSession := new(MySession)
	rack := csrf.Handler(mux, csrf.Session(userSession), csrf.Secret("my secret!"), csrf.Domain("localhost"))

	http.ListenAndServe(":8080", rack)
}
