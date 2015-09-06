package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// MinSeverity is the minimum severity that will be logged.
// The calling program can change this value at any time.
var (
	MinSeverity = SeverityInfo
)

var (
	errInvalidSeverity = errors.New("invalid severity")
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

// MarshalJSON implements the json.Marshaler interface.
func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errInvalidSeverity
	}
	switch strings.ToLower(str) {
	case "debug":
		*s = SeverityDebug
	case "info":
		*s = SeverityInfo
	case "warn", "warning":
		*s = SeverityWarning
	case "error":
		*s = SeverityError
	case "fatal":
		*s = SeverityFatal
	default:
		return errInvalidSeverity
	}
	return nil
}
