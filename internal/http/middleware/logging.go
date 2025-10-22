package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &logWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(lw, r)

		d := time.Since(start)
		log.Printf(" %s %s -> %d %dB (%s) UA=%q IP=%s",
			r.Method,
			r.URL.Path,
			lw.status,
			lw.bytes,
			d,
			r.UserAgent(),
			clientIP(r),
		)
	})
}

type logWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *logWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}
