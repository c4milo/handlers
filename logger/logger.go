package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type option func(*handler)

// Internal handler
type handler struct {
	name      string
	formatter logrus.Formatter
	out       io.Writer
}

// AppName allows to set the application name to log
func AppName(name string) option {
	return func(l *handler) {
		l.name = name
	}
}

// Formatter allows to set a custom log formatter
func Formatter(formatter logrus.Formatter) option {
	return func(l *handler) {
		l.formatter = formatter
	}
}

// Output allows setting an output writer for logging to be written to
func Output(out io.Writer) option {
	return func(l *handler) {
		l.out = out
	}
}

// Handler does HTTP request logging
func Handler(h http.Handler, opts ...option) http.Handler {
	// Default options
	handler := &handler{
		name:      "unknown",
		formatter: &logrus.TextFormatter{DisableColors: true},
		out:       os.Stdout,
	}

	for _, opt := range opts {
		opt(handler)
	}

	log := logrus.New()
	log.Formatter = handler.formatter
	log.Out = handler.out

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.NewV4().String()
		w.Header().Set("RequestID", requestID)

		log.WithFields(logrus.Fields{
			"request-id": requestID,
			"remote":     r.RemoteAddr,
			"method":     r.Method,
			"request":    r.RequestURI,
		}).Infof("%s started", requestID)

		res := NewResponseWriter(w)
		h.ServeHTTP(res, r)

		latency := time.Since(start)
		log.WithFields(logrus.Fields{
			"request-id": requestID,
			"remote":     r.RemoteAddr,
			"method":     r.Method,
			"request":    r.RequestURI,
			"status":     res.Status(),
			"size":       res.Size(),
			"took":       latency,
			fmt.Sprintf("measure#%s.latency", handler.name): latency.Nanoseconds(),
		}).Infof("%s completed", requestID)
	})
}
