package compressor

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/assert"
)

func TestGzipHandler(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world.")
	})

	gzipHandler := GzipHandler(requestHandler, GzipLevel(DefaultCompression))
	ts := httptest.NewServer(gzipHandler)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	assert.Ok(t, err)
	req.Header.Set(acceptEncoding, gzipEncoding)

	resp, err := http.DefaultClient.Do(req)
	assert.Ok(t, err)

	// Tests for vary header
	assert.Equals(t, acceptEncoding, resp.Header.Get(vary))

	assert.Equals(t, "36", resp.Header.Get(contentLength))
	assert.Equals(t, "application/x-gzip", resp.Header.Get(contentType))

	gr, err := gzip.NewReader(resp.Body)
	assert.Ok(t, err)
	defer gr.Close()

	body, err := ioutil.ReadAll(gr)
	assert.Ok(t, err)

	assert.Equals(t, "Hello world.", string(body))
}
