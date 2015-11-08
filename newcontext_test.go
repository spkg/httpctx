package httpctx_test

import (
	"net/http"
	"sync"
	"testing"

	"github.com/spkg/httpctx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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

type fakeWriter struct {
	CloseNotifyChan chan bool
}

func newFakeWriter() *fakeWriter {
	return &fakeWriter{
		CloseNotifyChan: make(chan bool),
	}
}

func (fw *fakeWriter) CloseNotify() <-chan bool {
	return fw.CloseNotifyChan
}

func (fw *fakeWriter) Header() http.Header {
	panic("not implemented")
}

func (fw *fakeWriter) Write([]byte) (int, error) {
	panic("not implemented")
}

func (fw *fakeWriter) WriteHeader(status int) {
	panic("not implemented")
}

func TestNewContext_CloseNotifier(t *testing.T) {
	assert := assert.New(t)
	fw := newFakeWriter()
	fr := &http.Request{}
	ctx := context.Background()

	ctx, cancelFunc := httpctx.NewContext(ctx, fw, fr)
	defer cancelFunc()

	var wg sync.WaitGroup
	finished := false
	wg.Add(1)
	go func(ctx context.Context) {
		<-ctx.Done()
		finished = true
		wg.Done()
	}(ctx)

	fw.CloseNotifyChan <- true
	wg.Wait()
	assert.True(finished)
}

func TestNewContext_CancelFunc(t *testing.T) {
	assert := assert.New(t)
	fw := newFakeWriter()
	fr := &http.Request{}
	ctx := context.Background()

	ctx, cancelFunc := httpctx.NewContext(ctx, fw, fr)

	var wg sync.WaitGroup
	finished := false
	wg.Add(1)
	go func(ctx context.Context) {
		<-ctx.Done()
		finished = true
		wg.Done()
	}(ctx)

	cancelFunc()
	wg.Wait()
	assert.True(finished)
}
