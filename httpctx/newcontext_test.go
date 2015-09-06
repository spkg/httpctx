package httpctx_test

import (
	"golang.org/x/net/context"
	"net/http"
	"sp.com.au/exp/httpctx"
)

func ExampleNewContext() {

	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := httpctx.NewContext(context.Background(), w, r)
		defer cancelFunc()
		doSomethingWith(ctx, w, r)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func doSomethingWith(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// ... perform processing here ...
}
