package httpctx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// shouldSendJson decides whether it is appropriate to send a JSON
// error response to the HTTP client.
func shouldSendJSON(r *http.Request) bool {
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
	if errWithStatusCode, ok := err.(interface {
		StatusCode() int
	}); ok {
		if errWithStatusCode.StatusCode() != 0 {
			statusCode = errWithStatusCode.StatusCode()
		}
	}

	var code string
	type ErrorWithCode interface {
		Code() string
	}
	if errWithCode, ok := err.(interface {
		Code() string
	}); ok {
		code = errWithCode.Code()
	}

	// remove headers that might have been set upstream
	w.Header().Del("Content-Encoding")

	if shouldSendJSON(r) {
		var b []byte

		// Put the relevant information into a map and marshal it.
		// The Result will look like:
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

		// If this does not succeed, then all we can do is to
		// send back the status code to the client, but cannot
		// send any payload.
		b, err = json.Marshal(resp)
		if err != nil {
			b = nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.WriteHeader(statusCode)
		if b != nil {
			w.Write(b)
		}
	} else {
		http.Error(w, err.Error(), statusCode)
	}
}
