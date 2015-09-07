package httpctx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// shouldSendJson decides whether it is appropriate to send a JSON
// error response to the HTTP client.
func shouldSendJson(r *http.Request) bool {
	accept := r.Header.Get("Accept")

	// TODO: this is a very weak interpretation of the
	// HTTP Accept header, but in practice it works fine.
	// Note also the hack that if the url starts with "/api", we
	// will return any error as a JSON. In practice this is useful
	// enough when working with a web browser during development.
	return strings.Contains(accept, "application/json") ||
		strings.HasPrefix(r.URL.Path, "/api/")
}

// sendError sends an error message back to the client.
// Note that the error returned to the client contains the
// message returned by err.Error(). It is the calling program's
// responsibility not to return sensitive information in this
// error message.
func sendError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	type Error interface {
		StatusCode() int
	}
	if errWithStatusCode, ok := err.(Error); ok {
		if errWithStatusCode.StatusCode() != 0 {
			statusCode = errWithStatusCode.StatusCode()
		}
	}

	var code string
	type ErrorWithCode interface {
		Code() string
	}
	if errWithCode, ok := err.(ErrorWithCode); ok {
		code = errWithCode.Code()
	}

	// remove headers that might have been set upstream
	w.Header().Del("Content-Encoding")

	if shouldSendJson(r) {
		var b []byte

		// if the error object can marshal itself, let it do so
		if marshaler, ok := err.(json.Marshaler); ok {
			b, _ := json.Marshal(err)
		} else {
			// the error object does not know how to marshal itself,
			// so put the relevant information into a map and marshal
			// that. The Result will look like:
			// {"error":{"message":"message-here","code":"xyz123","status":400}}
			resp := map[string]map[string]interface{}{
				"error": {
					"message": err.Error(),
					"status":  statusCode,
				},
			}

			if code != "" {
				resp["error"]["code"] = code
			}

			b, _ := json.Marshal(resp)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.WriteHeader(statusCode)
		w.Write(b)
	} else {
		http.Error(w, err.Error(), statusCode)
	}
}
