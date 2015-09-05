// Package log provides diagnostic logging.
//
// There are many good logging packages available, and it is worth asking
// why the world needs another one.
//
// Here are some differentiators for this package. Not all of them are
// unique, but this is the only package (to date) that has all of them.
//
// 1. Log messages are not formatted using a printf style interface. Each
// log message should have a constant message, which makes it easier to
// filter and search for messages. Any variable information is passed as
// properties in the message (see the WithValue function).
//
// 2. Uses an api that allows for multiple options and parameters to be
// logged in a single call. (See "Functional options for friendly APIs"
// by Dave Cheney http://bit.ly/1x9WWPi).
//
// 3. When a message is logged, a non-nil *Message value is returned, which
// can be returned as an error value.
//
// 4. This package is context aware (golang.org/x/net/context). Contexts
// can be created with information that will be logged with the message.
//
// 5. Messages can be logged as text messages, or structured (JSON) messages.
package log

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"os"
)

// Severity indicates the severity of a log message.
type Severity int

const (
	SeverityDebug   Severity = iota // Debugging only
	SeverityInfo                    // Informational
	SeverityWarning                 // Warning that might be recoverable
	SeverityError                   // Requires intervention
	SeverityFatal                   // Program will terminate
)

var (
	MinSeverity = SeverityInfo
)

// String implements the String interface.
func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "debug"
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warn"
	case SeverityError:
		return "error"
	case SeverityFatal:
		return "fatal"
	}
	return fmt.Sprintf("unknown %d", s)
}

// Message contains all of the log message information.
// Note that *Message implements the error interface.
type Message struct {
	Severity   Severity
	Text       string
	Parameters []Parameter
	Context    context.Context
	Err        error
	StatusCode int
}

// Parameter contains additional information about
// a log message.
type Parameter struct {
	Name  string
	Value interface{}
}

// apply applies all of the option functions to the message. It also
// adds any information from a
func (m *Message) apply(opts []func(*Message)) {
	for _, opt := range opts {
		opt(m)
	}
	for data := fromContext(m.Context); data != nil; data = data.Prev {
		m.Parameters = append(m.Parameters, Parameter{data.Name, data.Value})
	}
}

// Fprint prints the log Text Message to the writer w.
func (m *Message) Fprint(w io.Writer) {
	// TODO: more sanitizing of parameter values, particularly
	// strings as they might be malicious client input
	switch len(m.Parameters) {
	case 0, 1, 2:
		fmt.Fprintf(w, "%s: %s", m.Severity, m.Text)
		for _, param := range m.Parameters {
			switch v := param.Value.(type) {
			case string:
				fmt.Fprintf(w, ", %s=%q", param.Name, v)
			case *string:
				if v == nil {
					fmt.Fprintf(w, ", %s=nil", param.Name)
				} else {
					fmt.Fprintf(w, ", %s=%q", param.Name, *v)
				}
			default:
				fmt.Fprintf(w, ", %s=%+v", param.Name, param.Value)
			}
		}
		fmt.Fprintf(w, "\n")
	default:
		fmt.Fprintf(w, "%s: %s\n", m.Severity, m.Text)
		for _, param := range m.Parameters {
			switch v := param.Value.(type) {
			case string:
				fmt.Fprintf(w, "    %s=%q\n", param.Name, v)
			case *string:
				if v == nil {
					fmt.Fprintf(w, "    %s=nil\n", param.Name)
				} else {
					fmt.Fprintf(w, "    %s=%q\n", param.Name, *v)
				}
			default:
				fmt.Fprintf(w, "    %s=%+v\n", param.Name, param.Value)
			}
		}
	}
}

// Print prints the log Text Message to standard output.
var Print func(m *Message) = func(m *Message) {
	m.Fprint(os.Stdout)
}

func doPrint(m *Message) {
	if Print != nil && m.Severity >= MinSeverity {
		Print(m)
	}
}

// Error implements the error interface
func (m *Message) Error() string {
	return m.Text
}

// WithValue sets a parameter with a name and a value.
func WithValue(name string, value interface{}) func(*Message) {
	return func(m *Message) {
		m.Parameters = append(m.Parameters, Parameter{name, value})
	}
}

// WithContext sets the context for the log message.
func WithContext(ctx context.Context) func(*Message) {
	return func(m *Message) {
		m.Context = ctx
	}
}

// WithError sets the error associated with the log message.
func WithError(err error) func(*Message) {
	return func(m *Message) {
		m.Err = err
	}
}

// WithSeverity sets the severity of the message. Useful for the Err function,
// for which the default severity is error.
func WithSeverity(s Severity) func(*Message) {
	return func(m *Message) {
		m.Severity = s
	}
}

// WithStatusCode sets the HTTP status code associated with the log message.
func WithStatusCode(code int) func(*Message) {
	return func(m *Message) {
		m.StatusCode = code
	}
}

func WithBadRequest() func(*Message) {
	return func(m *Message) {
		m.StatusCode = http.StatusBadRequest
	}
}

// Debug logs a debug severity message.
func Debug(text string, opts ...func(*Message)) *Message {
	m := &Message{
		Severity: SeverityDebug,
		Text:     text,
	}
	m.StatusCode = http.StatusOK
	m.apply(opts)
	doPrint(m)
	return m
}

// Info logs an info severity message.
func Info(text string, opts ...func(*Message)) *Message {
	m := &Message{
		Severity: SeverityInfo,
		Text:     text,
	}
	m.StatusCode = http.StatusOK
	m.apply(opts)
	doPrint(m)
	return m
}

// Err logs a message based on an error value. The default
// severity is error, but this can be overridden to a different
// value with the WithValue() function.
func Err(err error, opts ...func(*Message)) *Message {
	m := &Message{
		Severity:   SeverityError,
		Text:       err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
	type t interface {
		StatusCode() int
	}
	if e, ok := err.(t); ok {
		m.StatusCode = e.StatusCode()
	}
	m.apply(opts)
	doPrint(m)
	return m
}

// Warn logs a warning severity message.
func Warn(text string, opts ...func(*Message)) *Message {
	m := &Message{
		Severity:   SeverityError,
		Text:       text,
		StatusCode: http.StatusInternalServerError,
	}
	m.apply(opts)
	doPrint(m)
	return m
}

// Error logs an error severity message.
func Error(text string, opts ...func(*Message)) *Message {
	m := &Message{
		Severity:   SeverityError,
		Text:       text,
		StatusCode: http.StatusInternalServerError,
	}
	m.apply(opts)
	doPrint(m)
	return m
}
