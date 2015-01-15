package logger

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type option func(*handler)

// Internal handler
type handler struct {
	name   string
	format string
	out    io.Writer
}

// AppName allows to set the application name to log.
func AppName(name string) option {
	return func(l *handler) {
		l.name = name
	}
}

// Format allows to set a custom log format. Although, the timestamp is always logged at the beginning.
// This handler is a bit opinionated.
//
// Directives:
//
// {remote_user}		: Remote user if Basic Auth credentials were sent
// {remote_ip}			: Remote IP address.
// {latency}			: The time taken to serve the request, in microseconds.
// {latency_human}		: The time taken to serve the request, human readable.
// {id}					: The request ID.
// {host}				: The Host header sent to the server
// {method}				: The request method. Ex: GET, POST, DELETE, etc.
// {url}				: The URL path requested.
// {query}				: Request's query string
// {rxbytes}			: Bytes received without headers
// {txbytes}			: Bytes sent, excluding HTTP headers.
// {status}				: Status sent to the client
// {useragent}			: User Agent
// {appname}			: The application name for this server
// {referer}			: The site from where the request came from
//
func Format(format string) option {
	return func(l *handler) {
		l.format = format
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
		name:   "unknown",
		format: `{appname} {id} remote_ip={remote_ip} {method} "{host}{url}?{query}" rxbytes={rxbytes} status={status} latency_human={latency_human} latency={latency} txbytes={txbytes}`,
		out:    os.Stdout,
	}

	for _, opt := range opts {
		opt(handler)
	}

	log.SetOutput(handler.out)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// If there is a request ID already, we use it to keep the transaction
		// traceable. If not, we generate a new request ID.
		reqID := w.Header().Get("RequestID")
		if reqID == "" {
			reqID = uuid.NewV4().String()
		}

		w.Header().Set("RequestID", reqID)

		log.Print(applyLogFormat(handler.format, handler.name, -1, w, r))

		res := NewResponseWriter(w)
		h.ServeHTTP(res, r)

		latency := time.Since(start)
		log.Print(applyLogFormat(handler.format, handler.name, latency, res, r))
	})
}

func applyLogFormat(format, appname string, latency time.Duration, w http.ResponseWriter, r *http.Request) string {
	reqID := w.Header().Get("RequestID")

	if strings.Index(format, "{appname}") > -1 {
		format = strings.Replace(format, "{appname}", appname, -1)
	}

	if strings.Index(format, "{remote_ip}") > -1 {
		format = strings.Replace(format, "{remote_ip}", strings.Split(r.RemoteAddr, ":")[0], -1)
	}

	if strings.Index(format, "{remote_user}") > -1 {
		user, _, _ := r.BasicAuth()
		if user == "" {
			user = r.URL.User.Username()
		}
		format = strings.Replace(format, "{remote_user}", user, -1)
	}

	if strings.Index(format, "{latency_human}") > -1 {
		l := "..."
		if latency > -1 {
			l = latency.String()
		}
		format = strings.Replace(format, "{latency_human}", l, -1)
	}

	if strings.Index(format, "{latency}") > -1 {
		l := "..."
		if latency > -1 {
			l = strconv.FormatInt(latency.Nanoseconds(), 10)
		}
		format = strings.Replace(format, "{latency}", l, -1)
	}

	if strings.Index(format, "{id}") > -1 {
		format = strings.Replace(format, "{id}", reqID, -1)
	}

	if strings.Index(format, "{method}") > -1 {
		format = strings.Replace(format, "{method}", r.Method, -1)
	}

	if strings.Index(format, "{url}") > -1 {
		format = strings.Replace(format, "{url}", r.URL.Path, -1)
	}

	if strings.Index(format, "{query}") > -1 {
		format = strings.Replace(format, "{query}", r.URL.RawQuery, -1)
	}

	if strings.Index(format, "{rxbytes}") > -1 {
		format = strings.Replace(format, "{rxbytes}", strconv.FormatInt(r.ContentLength, 10), -1)
	}

	if strings.Index(format, "{txbytes}") > -1 {
		size := "..."
		if v, ok := w.(ResponseWriter); ok {
			size = strconv.Itoa(v.Size())
		}
		format = strings.Replace(format, "{txbytes}", size, -1)
	}

	if strings.Index(format, "{status}") > -1 {
		status := "..."
		if v, ok := w.(ResponseWriter); ok {
			status = strconv.Itoa(v.Status())
		}
		format = strings.Replace(format, "{status}", status, -1)
	}

	if strings.Index(format, "{useragent}") > -1 {
		format = strings.Replace(format, "{useragent}", r.UserAgent(), -1)
	}

	if strings.Index(format, "{host}") > -1 {
		format = strings.Replace(format, "{host}", r.Host, -1)
	}

	if strings.Index(format, "{referer}") > -1 {
		format = strings.Replace(format, "{referer}", r.Referer(), -1)
	}

	return format
}
