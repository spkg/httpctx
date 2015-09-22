package log

/*

import "golang.org/x/net/context"

type Logger interface {
	Debug(text string, opts ...func(*Message)) *Message
	Info(text string, opts ...func(*Message)) *Message
	Warn(text string, opts ...func(*Message)) *Message
	Error(text string, opts ...func(*Message)) *Message

	WithError(err error) func(*Message)
	WithValue(name string, value interface{}) func(*Message)
	WithStatusCode(code int) func(*Message)
}

func FromContext(ctx context.Context) Logger {
	return &logger{ctx: ctx}
}

type logger struct {
	ctx context.Context
}

func (l *logger) Debug(text string, opts ...func(*Message)) *Message {
	return newMessage(SeverityError, text).
		applyOpt(WithContext(l.ctx)).
		applyOpts(opts)
}

func (l *logger) Info(text string, opts ...func(*Message)) *Message {
	return newMessage(SeverityInfo, text).
		applyOpt(WithContext(l.ctx)).
		applyOpts(opts)
}

func (l *logger) Warn(text string, opts ...func(*Message)) *Message {
	return newMessage(SeverityWarning, text).
		applyOpt(WithContext(l.ctx)).
		applyOpts(opts)
}

func (l *logger) Error(text string, opts ...func(*Message)) *Message {
	return newMessage(SeverityError, text).
		applyOpt(WithContext(l.ctx)).
		applyOpts(opts)
}

func (l *logger) WithError(err error) func(*Message) {
	return WithError(err)
}

func (l *logger) WithValue(name string, value interface{}) func(*Message) {
	return WithValue(name, value)
}

func (l *logger) WithStatusCode(code int) func(*Message) {
	return WithStatusCode(code)
}

*/
