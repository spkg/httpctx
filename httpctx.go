// Package httpctx provides a convenient way to handle HTTP requests
// using "context-aware" handler functions. The "context-aware" handler
// functions are different from the standard http.Handler functions in two
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
// a result. (Thanks to Justinas Stankevičius for the idea for this. See github.com/justinas/alice).
package httpctx

import (
	"context"
	"net/http"
)

// newContext creates a new context.Context for the HTTP request and
// a cancel function to call when the HTTP request is finished. The context's
// Done channel will be closed if the HTTP client closes the connection
// while the request is being processed. (This feature relies on w implementing
// the http.CloseNotifier interface).
//
// Note that the caller must ensure that the cancel function is called
// when the HTTP request is complete, or a goroutine leak could result.
func newContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc) {
	// TODO: the request r is not used in this function, and perhaps it should
	// be removed. Is there any reason to keep it. A future version of Go might
	// keep a context in the request object, so it is kept here for now.
	var cancelFunc context.CancelFunc
	if ctx == nil {
		ctx = context.Background()
	}

	// create a context without a timeout
	ctx, cancelFunc = context.WithCancel(ctx)

	if closeNotifier, ok := w.(http.CloseNotifier); ok {
		// need to acquire the channel prior to entering
		// the go-routine, otherwise CloseNotify could be
		// called after the request is finished, which
		// results in a panic
		closeChan := closeNotifier.CloseNotify()
		go func() {
			select {
			case <-closeChan:
				cancelFunc()
				return
			case <-ctx.Done():
				return
			}
		}()
	}

	return ctx, cancelFunc
}

// The Handler interface used for registering handlers that serve
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
		// Note that if the handler h has been created using a stack (ie Context
		// or Use functions), the first middleware in the stack will replace the context.
		// Pass the background context here just in case h has been constructed a
		// different way, but this will be rare.
		err := h.ServeHTTPContext(context.Background(), w, r)
		if err != nil {
			sendError(w, r, err)
		}
	})
}

// HandleFunc returns a http.Handler (compatible with the standard library http package), which
// calls the handler function f.
func HandleFunc(f func(context.Context, http.ResponseWriter, *http.Request) error) http.Handler {
	return Handle(HandlerFunc(f))
}

// Context returns a middleware stack that applies the context to any
// handlers added to the stack. This is useful when the main program creates a
// context that should be used as the base context for all HTTP handlers.
func Context(ctx context.Context) *Stack {
	if ctx == nil {
		ctx = context.Background()
	}
	// Create middleware that ignores the supplied context, and sets up a
	// context based on ctx that cancels if the request cancels.
	m := func(h Handler) Handler {
		return HandlerFunc(func(_ context.Context, w http.ResponseWriter, r *http.Request) error {
			var cancel func()
			var requestCtx context.Context
			requestCtx, cancel = newContext(ctx, w, r)
			defer cancel()
			return h.ServeHTTPContext(requestCtx, w, r)
		})
	}
	return &Stack{
		middleware: m,
		previous:   nil,
	}
}

// Use creates a Stack of middleware functions.
func Use(f ...func(h Handler) Handler) *Stack {
	var stack = Context(context.Background())

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

// HandleFunc returns a http.Handler (compatible with the standard library http package), which
// calls the middleware handlers in the stack s, followed by  the handler function f.
func (s *Stack) HandleFunc(f func(context.Context, http.ResponseWriter, *http.Request) error) http.Handler {
	return s.Handle(HandlerFunc(f))
}
