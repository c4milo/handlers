# Go HTTP Handlers
[![GoDoc](https://godoc.org/github.com/c4milo/handlers?status.svg)](https://godoc.org/github.com/c4milo/handlers)
[![Build Status](https://travis-ci.org/c4milo/handlers.svg?branch=master)](https://travis-ci.org/c4milo/handlers)

This repository contains HTTP middlewares that I use in my own Go projects.
Feel free to use them too!


* **Compressor:** Applies gzip compression to the response body, if the client supports it.
* **Logger:** Logs HTTP requests, including: remote user, remote IP, latency, request id, txbytes, rxbytes, status, etc.
* **HTTP Method Override:** Provides an alternative for clients that don't support methods other than POST or GET  to override the HTTP method.

For examples on how to use these handlers, please refer to the Go documentation linked at the top.
