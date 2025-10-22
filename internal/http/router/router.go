package router

import (
	"net/http"
	"strings"
	"todo-api/internal/http/path"
	"todo-api/internal/pkg"
)

type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
}

type routeKey struct {
	Method  string // UPPER
	Pattern string // UPPER NORM
}

func newRouteKey(r Route) routeKey {

	normPat := "/" + strings.Trim(strings.TrimSpace(r.Pattern), "/")
	return routeKey{
		Method:  strings.ToUpper(strings.TrimSpace(r.Method)),
		Pattern: normPat,
	}
}

func areRoutesEqual(r1, r2 Route) bool {
	rk1, rk2 := newRouteKey(r1), newRouteKey(r2)
	return rk1 == rk2
}

type Router struct {
	routes []Route
}

func (router *Router) Handle(method, pattern string, h http.Handler) {
	sb := strings.Builder{}
	sb.WriteByte('/')
	sb.WriteString(strings.Trim(strings.TrimSpace(pattern), "/"))

	route := Route{
		Method:  method,
		Pattern: sb.String(),
		Handler: h,
	}

	// find routes with same path and rewrite
	for i, r := range router.routes {
		if areRoutesEqual(r, route) {
			router.routes[i] = route
			return
		}
	}

	// add path if new
	router.routes = append(router.routes, route)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range router.routes {
		ok, vals, err := path.Match(route.Pattern, r.URL.EscapedPath())
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if ok {
			if r.Method != route.Method {
				allow = append(allow, route.Method)
				continue
			}

			scope := &pkg.Scope{Params: vals}
			r = pkg.WithScope(r, scope)
			route.Handler.ServeHTTP(w, r)
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r)
}
