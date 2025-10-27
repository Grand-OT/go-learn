package router

import (
	"net/http"
	"strings"
	httpx "todo-api/internal/http"
	"todo-api/internal/http/middleware"
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
	base   string
	mws    []middleware.Middleware
}

func joinPath(a, b string) string {
	return "/" + strings.TrimSuffix(strings.TrimPrefix(a, "/"), "/") +
		"/" + strings.TrimPrefix(b, "/")
}

func (router *Router) Group(path string, fn func(*Router)) {
	child := &Router{
		routes: router.routes,
		base:   joinPath(router.base, path),
		mws:    append([]middleware.Middleware{}, router.mws...),
	}
	fn(child)

	router.routes = child.routes
}

func (router *Router) Use(mws ...middleware.Middleware) {
	router.mws = append(router.mws, mws...)
}

func (router *Router) Handle(method, pattern string, h http.Handler) {
	pattern = joinPath(router.base, pattern)
	h = middleware.Chain(router.mws...)(h)
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
			httpx.WriteError(w, http.StatusBadRequest, "invalid_path", "unable to handle path")
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
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "requested method not allowed")
		return
	}

	httpx.WriteError(w, http.StatusNotFound, "path_not_found", "requested path not found")
}
