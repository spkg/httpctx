package httpctx_test

import (
	"context"
	"net/http"

	"github.com/spkg/httpctx"
)

func ExampleContext() {
	ctx := context.Background()

	// Create a new context based on the background context.
	// For example, cancel on SIGINT and/or SIGTERM.
	ctx = makeNewContext(ctx)

	// Any call using the public stack will be passed ctx
	// instead of the background context for the handlers
	// in the stack.
	public := httpctx.Context(ctx).Use(ensureHttps)

	// Because authenticate is based on public, any call
	// using this stack will also be passed ctx instead
	// of the background context for handlers in the stack.
	authenticate := public.Use(authenticate, ensureAdmin)

	http.Handle("/admin", authenticate.HandleFunc(admin))
	http.Handle("/", public.HandleFunc(index))
	http.ListenAndServe(":8080", nil)
}

func makeNewContext(ctx context.Context) context.Context {
	// Just an example of setting up a context, you should
	// avoid using strings for keys in context.WithValue
	ctx = context.WithValue(ctx, "some-key", "some-value")
	return ctx
}
