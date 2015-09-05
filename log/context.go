package log

import "golang.org/x/net/context"

type logData struct {
	Name  string
	Value interface{}
	Prev  *logData
}

type contextKey int

const (
	keyLogData contextKey = iota
)

// NewContext returns a new context that has the name and value associated with it.
// The name and value will become part any log message logged with this context.
func NewContext(ctx context.Context, name string, value interface{}) context.Context {
	prev, ok := ctx.Value(keyLogData).(*logData)
	data := &logData{
		Name:  name,
		Value: value,
	}
	if ok {
		data.Prev = prev
	}
	return context.WithValue(ctx, keyLogData, data)
}

func fromContext(ctx context.Context) *logData {
	if ctx == nil {
		return nil
	}
	data, ok := ctx.Value(keyLogData).(*logData)
	if !ok {
		return nil
	}
	return data
}
