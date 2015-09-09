// Package httpctx provides a convenient way to handle HTTP requests
// using "context-aware" handler functions. The "context-aware" handler
// functions are different to the standard http.Handler functions in two
// important ways:
//
// 1. They accept an additional parameter of the (almost) standard type context.Context
// (golang.org/x/net/context); and
//
// 2. They return an error result.
//
// Passing a context.Context to the handler functions is a useful addition because the
// handler functions can make use of any values associated with the context; and also
// because the Done channel of the context will be closed if the HTTP client closes the
// connection prematurely; or a timeout period has elapsed.
//
// Returning an error value simplifies error handling in the handler functions. Common
// error handling is provided, which can be enhanced by middleware.
//
// A simple middleware mechanism is also provided by this package. A middleware function
// is one that accepts a httpctx.Handler as a parameter and returns a httpctx.Handler as
// a result.
package httpctx // import "sp.com.au/exp/httpctx"

import (
	"golang.org/x/net/context"
	"net/http"
)

// NewContext creates a new context.Context for the HTTP request and
// a cancel function to call when the HTTP request is finished. The context's
// Done channel will be closed if the HTTP client closes the connection
// while the request is being processed. (This feature relies on w implementing
// the http.CloseNotifier interface).
//
// Note that the caller must ensure that the cancel function is called
// when the HTTP request is complete, or a goroutine leak could result.
func NewContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc) {
	var cancelFunc context.CancelFunc
	if ctx == nil {
		ctx = context.Background()
	}

	// create a context without a timeout
	ctx, cancelFunc = context.WithCancel(ctx)

	if closeNotifier, ok := w.(http.CloseNotifier); ok {
		go func() {
			select {
			case <-closeNotifier.CloseNotify():
				cancelFunc()
				return
			case <-ctx.Done():
				return
			}
		}()
	}

	return ctx, cancelFunc
}

// Objects implementing the Handler interface can be registered to serve
// a particular path or subtree in the HTTP server. If the ServeHTTPContext function
// returns an error, it is returned to the client as a HTTP error code.
type Handler interface {
	ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as HTTP handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// ServeHTTPContext calls f(ctx, w, r)
func (f HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return f(ctx, w, r)
}

// A Stack is a stack of middleware functions that are common to one or more
// HTTP handlers. A middleware function is any function that accepts a Handler as a
// parameter and returns a Handler.
type Stack struct {
	middleware func(Handler) Handler
	previous   *Stack
}

// Handle converts a httpctx.Handler into a http.Handler.
func Handle(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := NewContext(context.Background(), w, r)
		defer cancelFunc()
		err := h.ServeHTTPContext(ctx, w, r)
		if err != nil {
			sendError(w, r, err)
		}
	})
}

func HandleFunc(f func(context.Context, http.ResponseWriter, *http.Request) error) http.Handler {
	return Handle(HandlerFunc(f))
}

// Use creates a Stack of middleware functions.
func Use(f ...func(h Handler) Handler) *Stack {
	var stack *Stack

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				previous:   stack,
			}
		}
	}

	return stack
}

// Use creates a new stack by appending the middleware functions to
// the existing stack.
func (s *Stack) Use(f ...func(h Handler) Handler) *Stack {
	stack := s

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				previous:   stack,
			}
		}
	}

	return stack
}

// Handle creates a http.Handler from a stack of middleware
// functions and a httpctx.Handler.
func (s *Stack) Handle(h Handler) http.Handler {
	for stack := s; stack != nil; stack = stack.previous {
		h = stack.middleware(h)
	}

	return Handle(h)
}

func (s *Stack) HandleFunc(f func(context.Context, http.ResponseWriter, *http.Request) error) http.Handler {
	return s.Handle(HandlerFunc(f))
}
