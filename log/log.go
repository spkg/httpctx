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
// by Dave Cheney http://goo.gl/l2KaW3).
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
	"io"
	"net/http"
	"os"
	"time"
)

const (
	timeFormat = "2006-01-02T15:04:05-0700"
)

// Message contains all of the log message information.
// Note that *Message implements the error interface.
type Message struct {
	Timestamp  time.Time
	Severity   Severity
	Text       string
	Parameters []Parameter
	Context    []Parameter
	Err        error
	StatusCode int
}

// Parameter contains additional information about
// a log message.
type Parameter struct {
	Name  string
	Value interface{}
}

func newMessage(severity Severity, text string) *Message {
	m := &Message{
		Timestamp:  time.Now(),
		Severity:   severity,
		Text:       text,
		StatusCode: http.StatusInternalServerError,
	}
	return m
}

func (m *Message) applyOpt(opt Option) *Message {
	opt(m)
	return m
}

// applyOpts applies all of the option functions to the message.
func (m *Message) applyOpts(opts []Option) *Message {
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WriteTo prints the log Text Message to the writer w.
func (m *Message) WriteTo(w io.Writer) {
	// append parameters and context to form one larger list
	parameters := make([]Parameter, 0, len(m.Parameters)+len(m.Context)+1)
	parameters = append(parameters, m.Parameters...)
	if m.Err != nil {
		parameters = append(parameters, Parameter{"error", m.Err.Error()})
	}
	parameters = append(parameters, m.Context...)

	// TODO: more sanitizing of parameter values, particularly
	// strings as they might be malicious client input
	switch len(parameters) {
	case 0, 1, 2, 3, 4:
		fmt.Fprintf(w, "%s %-5s %s", m.Timestamp.Format(timeFormat), m.Severity, m.Text)
		for _, param := range parameters {
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
		fmt.Fprintf(w, "%s %-5s %s\n", m.Timestamp.Format(timeFormat), m.Severity, m.Text)
		for _, param := range parameters {
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
	m.WriteTo(os.Stdout)
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

// Debug logs a debug severity message.
func Debug(text string, opts ...Option) *Message {
	m := newMessage(SeverityDebug, text)
	m.applyOpts(opts)
	doPrint(m)
	return m
}

// Info logs an info severity message.
func Info(text string, opts ...Option) *Message {
	m := newMessage(SeverityInfo, text)
	m.applyOpts(opts)
	doPrint(m)
	return m
}

// Warn logs a warning severity message.
func Warn(text string, opts ...Option) *Message {
	m := newMessage(SeverityWarning, text)
	m.applyOpts(opts)
	doPrint(m)
	return m
}

// Error logs an error severity message.
func Error(text string, opts ...Option) *Message {
	m := newMessage(SeverityError, text)
	m.applyOpts(opts)
	doPrint(m)
	return m
}
