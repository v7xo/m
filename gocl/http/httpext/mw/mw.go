package mw

import "net/http"

// Middleware wraps handlers.
type Middleware interface {
	// Wrap returns the specified handler wrapped with the middleware.
	Wrap(next http.Handler) http.Handler
}

// Func is the functional implementation of the [Middleware] interface.
type Func func(next http.Handler) http.Handler

// Wrap implements the [Middleware] interface.
// Wraps specified handler next with this handler f and returns new handler.
func (f Func) Wrap(next http.Handler) http.Handler {
	return f(next)
}

// NewMiddleware create a new [Middleware] from specified function fn.
func NewMiddleware(fn func(w http.ResponseWriter, r *http.Request, next http.Handler)) Middleware {
	var res Func = func(next http.Handler) http.Handler {
		var hnd http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) { fn(w, r, next) }

		return hnd
	}

	return res
}
