package httpctx

import (
	"expvar"
	"net/http"
	"strconv"
	"time"
)

var (
	expvarStatusCodes = expvar.NewMap("responses")
	expvarErrors      = expvar.NewMap("errors")
	expvarDuration    = expvar.NewMap("duration")
)

// responseWriter implements http.ResponseWriter, and records the
// the number of responses for each type of status code
type responseWriter struct {
	responseWriter http.ResponseWriter
	wroteHeader    bool
	started        time.Time
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		responseWriter: w,
		started:        time.Now(),
	}
}

func (w *responseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *responseWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		key := strconv.Itoa(status)
		expvarStatusCodes.Add(key, 1)
		expvarStatusCodes.Add("total", 1)
		if status >= 400 {
			expvarErrors.Add(key, 1)
			expvarErrors.Add("total", 1)
		}
	}

	w.responseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.responseWriter.Write(b)
}

func (w *responseWriter) finished() {
	duration := time.Now().Sub(w.started)

	duration /= time.Millisecond * 100
	duration *= time.Millisecond * 100

	key := strconv.Itoa(int(duration))
	expvarDuration.Add(key, 1)
}
