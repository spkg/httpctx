package httpctx_test

import (
	"net/http"
	"testing"

	"github.com/spkg/httpctx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func Test1(t *testing.T) {
	assert := assert.New(t)
	stack := httpctx.Use(middleware1, middleware2)
	assert.NotNil(stack)

	http.Handle("/api/whatever", stack.Handle(doWhateverHandler))
}

var doWhateverHandler = httpctx.HandlerFunc(doWhatever)

func doWhatever(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func middleware1(f httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	})
}

func middleware2(f httpctx.Handler) httpctx.Handler {
	return httpctx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	})
}
