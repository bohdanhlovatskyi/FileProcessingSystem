package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type GzipHandler struct{}

// register the middleware functino that will be responsible for intermidiate data
// caching when requesting for the big amount data requests
func (g *GzipHandler) GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// create a gzipped response
			we := NewWrappedResponseWriter(&rw)
			we.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(we, r)
			defer we.Flush()

			return
		}

		next.ServeHTTP(rw, r)
	})
}

// creates wrapper for a response (implements the methods to fit the interface)
type WrappedResponseWriter struct {
	rw http.ResponseWriter
	gw *gzip.Writer
}

func NewWrappedResponseWriter(rw *http.ResponseWriter) *WrappedResponseWriter {
	gw := gzip.NewWriter(*rw)

	return &WrappedResponseWriter{gw: gw, rw: *rw}
}

func (wr *WrappedResponseWriter) Header() http.Header {
	return wr.rw.Header()
}

func (wr *WrappedResponseWriter) Write(d []byte) (int, error) {
	return wr.gw.Write(d)
}

func (wr *WrappedResponseWriter) WriteHeader(statusCode int) {
	wr.rw.WriteHeader(statusCode)
}

func (wr *WrappedResponseWriter) Flush() {
	defer wr.gw.Close()
	wr.gw.Flush()
}
