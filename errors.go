package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Trace represents a single entry in an error's stack trace.
// It captures the location and time when an error was thrown or stamped.
type Trace struct {
	// The function where the error occurred
	Function string `json:"function,omitempty" yaml:"function,omitempty"`

	// The file where the error occurred
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// The line where the error occurred
	Line int `json:"line,omitempty" yaml:"line,omitempty"`

	// The timestamp when trace was generated
	Timestamp time.Time `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
}

// Error is an enhanced error implementation that supports structured error information,
// error wrapping, stack traces, and additional metadata.
//
// Error implements the standard error interface and can be used anywhere
// a standard Go error is expected.
type Error struct {
	// The error title
	Title string `json:"title" yaml:"title"`

	// A numeric error code for programmatic handling
	Identifier int32 `json:"identifier,omitempty" yaml:"identifier,omitempty"`

	// Additional context as a list of strings
	Details []string `json:"details,omitempty" yaml:"details,omitempty"`

	// Key-value pairs for arbitrary metadata
	Properties map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`

	// The underlying error that caused this error
	Cause error `json:"cause,omitempty" yaml:"cause,omitempty"`

	// A trace of where the error was thrown
	Stack []*Trace `json:"stack,omitempty" yaml:"stack,omitempty"`
}

// New creates a new Error with the given title.
func New(title string) error {
	return &Error{
		Title: title,
	}
}

// Wrap wraps an error with a message.
func Wrap(err error, msg string) error {
	trace := trace()
	return from(err, true).WithDetail(msg).throw(trace)
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, format string, args ...any) error {
	trace := trace()
	return from(err, true).WithDetailf(format, args...).throw(trace)
}

// From creates a new *Error from any error type.
// If the error is not an *Error, it creates a new error with title "unknown error"
// and sets the original error as the cause.
// If the error is an *Error, it returns a copy of the original error with the same
// title, identifier, details, properties.
func From(err error) *Error {
	return from(err, false)
}

func from(err error, copyStack bool) *Error {
	var t *Error

	ok := errors.As(err, &t)
	if !ok {
		t, _ = New("unknown error").(*Error)
	}

	e := &Error{
		Title:      t.Title,
		Identifier: t.Identifier,
		Details:    t.Details,
		Properties: t.Properties,
	}
	if copyStack {
		e.Stack = t.Stack
	}

	if !ok {
		e.Cause = err
	}

	return e
}

// Intercept converts any error into an *Error type.
// If the provided error is already an *Error, it returns it as-is.
// Otherwise, it creates a new *Error wrapping the original error using From().
func Intercept(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	return From(err)
}

// Is compares this error with another error for equality.
// Two errors are considered equal if they have the same Title and Identifier.
func (e *Error) Is(err error) bool {
	other := new(Error)
	if ok := errors.As(err, &other); !ok {
		return false
	}

	return e.Title == other.Title && e.Identifier == other.Identifier
}

// Is is a wrapper around errors.Is to compare two errors for equality.
func Is(err, target error) bool { return errors.Is(err, target) }

// As is a wrapper around errors.As to check if the error is of a specific type.
func As(err error, target any) bool { return errors.As(err, target) }

// Unwrap returns the underlying cause of this error, nil if no cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

// Error returns a formatted string representation of the error,
// including title, identifier, details, properties, location and cause.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	b := bytes.NewBuffer(nil)

	fmt.Fprintf(
		b,
		"%s (%d):",
		strings.ToLower(e.Title),
		e.Identifier,
	)

	if len(e.Details) > 0 {
		fmt.Fprintf(
			b,
			" %s:",

			strings.Join(e.Details, ": "),
		)
	}

	for k, v := range e.Properties {
		fmt.Fprintf(b, " %s='%v',", k, v)
	}

	if len(e.Stack) > 0 {
		tail := e.Stack[len(e.Stack)-1]

		if tail != nil {
			fmt.Fprintf(
				b,
				" at=(func='%s', file='%s', line='%d'),",
				path.Base(tail.Function),
				filepath.Base(tail.File),
				tail.Line,
			)
		}
	}

	if e.Cause != nil {
		fmt.Fprintf(b, " caused by: %v", e.Cause.Error())
	}

	return string(bytes.TrimSuffix(bytes.TrimSuffix(b.Bytes(), []byte(",")), []byte(":")))
}

// String returns a JSON representation of the error.
func (e *Error) String() string {
	if e == nil {
		return ""
	}

	b := bytes.NewBuffer(nil)
	json.NewEncoder(b).Encode(e) // nolint: errcheck // No way to get wrong here.

	return b.String()
}

// Stamp adds a stack trace entry to an existing error.
func Stamp(err error) error {
	trace := trace()

	return Intercept(err).throw(trace)
}

// WithIdentifier sets a numeric identifier for the error.
func (e *Error) WithIdentifier(id int32) *Error {
	e.Identifier = id

	return e
}

// WithDetail adds a detail string to the error for additional context.
func (e *Error) WithDetail(detail string) *Error {
	e.Details = append(e.Details, strings.TrimSuffix(detail, "."))

	return e
}

// WithDetailf adds a detail string to the error for additional context using a format string.
func (e *Error) WithDetailf(format string, args ...any) *Error {
	return e.WithDetail(fmt.Sprintf(format, args...))
}

// WithProperties adds multiple key-value properties to the error.
func (e *Error) WithProperties(properties map[string]any) *Error {
	for k, v := range properties {
		e.WithProperty(k, v) // nolint: errcheck // No way to get wrong here.
	}

	return e
}

// WithProperty adds a single key-value property to the error.
func (e *Error) WithProperty(key string, value any) *Error {
	if e.Properties == nil {
		e.Properties = make(map[string]any)
	}

	e.Properties[key] = value

	return e
}

// CausedBy sets the underlying cause of this error.
func (e *Error) CausedBy(err error) *Error {
	e.Cause = err

	return e
}

func (e *Error) throw(trace *Trace) error {
	if trace == nil {
		return e
	}

	e.Stack = append([]*Trace{trace}, e.Stack...)

	return e
}

// Throw adds a stack trace entry to the error and returns it as an error interface.
func (e *Error) Throw() error {
	trace := trace()

	return e.throw(trace)
}

func trace() *Trace {
	// 2 is the depth of the caller.
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return nil
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil
	}

	return &Trace{
		Function:  fn.Name(),
		File:      file,
		Line:      line,
		Timestamp: time.Now().UTC(),
	}
}
