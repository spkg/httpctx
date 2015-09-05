package httpctx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sp.com.au/exp/log"
)

// sendError sends an error message back to the client.
// TODO: this needs lots and lots of work.
func sendError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	type Error interface {
		StatusCode() int
	}
	if err1, ok := err.(Error); ok {
		statusCode = err1.StatusCode()
	}
	w.WriteHeader(statusCode)

	// TODO: dodgy test for JSON, should really check the request
	// Accept header, and possibly also the Content-Type.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		resp := map[string]interface{}{
			"error": map[string]interface{}{
				"message": err.Error(),
			},
		}

		b, err := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		_, err = w.Write(b)
		if err != nil {
			log.Warn("failed to write error to client",
				log.WithValue("error", err.Error()))
		}
	} else {
		if statusCode == http.StatusInternalServerError {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), statusCode)
		}
	}
}
