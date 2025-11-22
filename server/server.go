package server

import (
	"context"
	"net/http"
	"sort"
	"strings"
)

type Router struct {
	routes map[string]map[string]http.Handler
}

type contextKey string

const paramsKey contextKey = "pathParams"

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]http.Handler),
	}
}

func (r *Router) handle(method, path string, handler http.HandlerFunc) {
	if _, ok := r.routes[method]; !ok {
		r.routes[method] = make(map[string]http.Handler)
	}
	r.routes[method][path] = handler
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if methodRoutes, ok := r.routes[req.Method]; ok {
		if handler, ok := methodRoutes[req.URL.Path]; ok {
			handler.ServeHTTP(w, req)
			return
		}

		for pattern, handler := range methodRoutes {
			if params, matched := matchRoute(pattern, req.URL.Path); matched {
				ctx := context.WithValue(req.Context(), paramsKey, params)
				handler.ServeHTTP(w, req.WithContext(ctx))
				return
			}
		}
	}

	var allow []string
	for method, routes := range r.routes {
		if _, ok := routes[req.URL.Path]; ok {
			allow = append(allow, method)
			continue
		}

		for pattern := range routes {
			if _, ok := matchRoute(pattern, req.URL.Path); ok {
				allow = append(allow, method)
				break
			}
		}
	}

	if len(allow) > 0 {
		sort.Strings(allow)
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}

func matchRoute(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)

	for i := range patternParts {
		p := patternParts[i]
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			key := strings.Trim(p, "{}")
			params[key] = pathParts[i]
			continue
		}

		if p != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}

// PathParam retrieves a path parameter by key if the route was matched with parameters.
func PathParam(r *http.Request, key string) string {
	params, ok := r.Context().Value(paramsKey).(map[string]string)
	if !ok {
		return ""
	}

	return params[key]
}
