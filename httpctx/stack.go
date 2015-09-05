package httpctx

import (
	"golang.org/x/net/context"
	"net/http"
)

// A Stack is a list of middleware functions.
type Stack struct {
	middleware func(Handler) Handler
	prev       *Stack
}

func (s *Stack) Use(f ...func(h Handler) Handler) *Stack {
	stack := s

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				prev:       stack,
			}
		}
	}

	return stack
}

func (s *Stack) Then(h Handler) http.Handler {
	for stack := s; stack != nil; stack = stack.prev {
		h = stack.middleware(h)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := newContext(w, r)
		defer cancelFunc()
		err := h.ServeHTTPContext(ctx, w, r)
		if err != nil {
			sendError(w, r, err)
		}
	})
}

func Use(f ...func(h Handler) Handler) *Stack {
	var stack *Stack

	for _, m := range f {
		if m != nil {
			stack = &Stack{
				middleware: m,
				prev:       stack,
			}
		}
	}

	return stack
}

// newContext creates a new context.Context for the request, and a cancel function
func newContext(w http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc) {
	var cancelFunc context.CancelFunc
	var ctx context.Context = context.Background()

	if Timeout == 0 {
		// create a context without a deadline
		ctx, cancelFunc = context.WithCancel(ctx)
	} else {
		// create a context with a deadline
		ctx, cancelFunc = context.WithTimeout(ctx, Timeout)

	}

	if closeNotifier, ok := w.(http.CloseNotifier); ok {
		go func() {
			select {
			case <-closeNotifier.CloseNotify():
				cancelFunc()
				break
			case <-ctx.Done():
				break
			}
		}()
	}

	ctx = context.WithValue(ctx, keyRequest, r)

	return ctx, cancelFunc
}
