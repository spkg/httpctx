// Package httpctx provides http handling and routing with an associated context
// (golang.org/x/net/context)
package httpctx

import (
	"golang.org/x/net/context"
	"net/http"
	"time"
)

// Timeout is the amount of time any request has before it times out.
var Timeout time.Duration = time.Minute

// Objects implementing the Handler interface can be registered to serve
// a particular path or subtree in the HTTP server. If the ServeHTTPContext function
// returns an error, it is returned to the client as a HTTP error code.
type Handler interface {
	ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// ServeHTTP calls f(ctx, w, r)
func (f HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return f(ctx, w, r)
}

type contextKey int

const (
	keyRequest contextKey = iota
)

// Request returns the http.Request associated with the context.
func Request(ctx context.Context) *http.Request {
	r, ok := ctx.Value(keyRequest).(*http.Request)
	if ok {
		return r
	}
	return nil
}
