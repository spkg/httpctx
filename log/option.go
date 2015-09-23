package log

import (
	"net/http"

	"golang.org/x/net/context"
)

// An Option is a function option that can be applied when logging a message.
// See the example for how they are used. Options is based on Dave Cheney's article
// "Functional options for friendly APIs" (http://goo.gl/l2KaW3)
// that can be applied to a Message.
type Option func(*Message)

// WithError sets the error associated with the log message.
func WithError(err error) Option {
	return func(m *Message) {
		m.Err = err
	}
}

// WithContext sets the context for the log message. See the example for NewContext.
func WithContext(ctx context.Context) Option {
	return func(m *Message) {
		for data := fromContext(ctx); data != nil; data = data.Prev {
			m.Context = append(m.Context, Parameter{data.Name, data.Value})
		}
	}
}

// WithValue sets a parameter with a name and a value.
func WithValue(name string, value interface{}) Option {
	return func(m *Message) {
		m.Parameters = append(m.Parameters, Parameter{name, value})
	}
}

// WithStatusCode sets the HTTP status code associated with the log message.
func WithStatusCode(code int) Option {
	return func(m *Message) {
		m.StatusCode = code
	}
}

// WithStatusUnauthorised is equivalent to WithStatusCode(http.StatusBadRequest)
func WithStatusBadRequest() Option {
	return func(m *Message) {
		m.StatusCode = http.StatusBadRequest
	}
}

// WithStatusUnauthorised is equivalent to WithStatusCode(http.StatusUnauthorized)
func WithStatusUnauthorized() Option {
	return func(m *Message) {
		m.StatusCode = http.StatusUnauthorized
	}
}
