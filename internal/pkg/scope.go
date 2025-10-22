package pkg

import (
	"context"
	"net/http"
)

type Scope struct {
	RealIP string
	Params map[string]string
}

type keyScope struct{}

func WithScope(r *http.Request, s *Scope) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), keyScope{}, s))
}

func ScopeFrom(r *http.Request) *Scope {
	if v, ok := r.Context().Value(keyScope{}).(*Scope); ok {
		return v
	}
	return nil
}
