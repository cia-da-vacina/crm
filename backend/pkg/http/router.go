package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	chi chi.Router
}

func NewRouter() *Router {
	r := chi.NewRouter()

	return &Router{chi: r}
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.chi.Get(pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.chi.Post(pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.chi.Put(pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.chi.Patch(pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.chi.Delete(pattern, handler)
}

func (r *Router) Options(pattern string, handler http.HandlerFunc) {
	r.chi.Options(pattern, handler)
}

func (r *Router) Method(method, pattern string, handler http.Handler) {
	r.chi.Method(method, pattern, handler)
}

// Group creates a sub-router. When pattern is "/" or empty, an inline group is
// used (chi.With) so multiple callers can share the root path without conflict.
// Otherwise, a new Mux is mounted at the given path prefix.
func (r *Router) Group(pattern string, fn func(*Router), middlewares ...func(http.Handler) http.Handler) {
	if pattern == "" || pattern == "/" {
		subRouter := &Router{chi: r.chi.With(middlewares...)}
		fn(subRouter)
	} else {
		subMux := chi.NewRouter()
		for _, mw := range middlewares {
			subMux.Use(mw)
		}
		subRouter := &Router{chi: subMux}
		fn(subRouter)
		r.chi.Mount(pattern, subMux)
	}
}

func (r *Router) Mount(pattern string, handler http.Handler) {
	r.chi.Mount(pattern, handler)
}

func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.chi.Use(middleware)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.chi.HandleFunc(pattern, handler)
}
