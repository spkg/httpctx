package httpctx_test

import (
	"net/http"
	"net/http/httptest"
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

func TestHandler(t *testing.T) {
	emptyFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	}
	for i, tc := range []struct {
		http.Handler
	}{
		{httpctx.HandleFunc(emptyFunc)},
		{httpctx.Use(middleware1).Use(middleware2).HandleFunc(emptyFunc)},
	} {
		srv := httptest.NewServer(tc.Handler)
		resp, err := http.Get(srv.URL)
		srv.Close()
		resp.Body.Close()
		if err != nil {
			t.Errorf("%d. %v", i, err)
		}
	}
}
