package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func Chain(mw ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	}
}

func clientIP(r *http.Request) string {
	// sended from proxy or balancer
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return xf
	}
	return r.RemoteAddr
}
