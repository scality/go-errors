package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
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
	// Options for the error
	Options Opts `json:"options,omitempty" yaml:"options,omitempty"`
	// A trace of where the error was thrown
	Stack []*Trace `json:"stack,omitempty" yaml:"stack,omitempty"`
}

type Opts struct {
	// A numeric error code for programmatic handling
	Identifier []uint32 `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	// Additional context as a list of strings
	Details []string `json:"details,omitempty" yaml:"details,omitempty"`
	// Key-value pairs for arbitrary metadata
	Properties map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`
	// The underlying error that caused this error
	Cause error `json:"cause,omitempty" yaml:"cause,omitempty"`
}

// New creates a new Error with the given title.
func New(title string) error {
	return &Error{
		Title: title,
		Options: Opts{
			Details:    []string{},
			Properties: make(map[string]any),
			Cause:      errors.New(title),
		},
	}
}

// Option is a function type that modifies the Options struct.
type Option func(*Opts)

// WithIdentifier sets a numeric identifier for the error.
func WithIdentifier(id uint32) Option {
	return func(c *Opts) {
		c.Identifier = append(c.Identifier, id)
	}
}

// WithDetail sets a detail string for the error.
func WithDetail(msg string) Option {
	return func(c *Opts) {
		c.Details = append(c.Details, msg)
	}
}

// WithDetailf sets a detail string for the error using a format string.
func WithDetailf(format string, args ...any) Option {
	return func(c *Opts) {
		c.Details = append(c.Details, fmt.Sprintf(format, args...))
	}
}

// WithIdentifier sets a numeric identifier for the error.
func WithProperty(key string, value any) Option {
	return func(c *Opts) {
		c.Properties[key] = value
	}
}

// CausedBy sets the underlying cause of this error.
func CausedBy(err error) Option {
	return func(c *Opts) {
		c.Cause = err
	}
}

// Wrap wraps an error with a message.
func Wrap(err error, opts ...Option) error {
	trace := trace()
	return from(err, true, opts...).throw(trace)
}

// Is compares this error with another error for equality.
// Two errors are considered equal if they have the same Title and Identifiers.
func (e *Error) Is(err error) bool {
	other := new(Error)
	if ok := errors.As(err, &other); !ok {
		return false
	}

	return e.Title == other.Title && slices.Equal(e.Options.Identifier, other.Options.Identifier)
}

// Is is a wrapper around errors.Is to compare two errors for equality.
func Is(err, target error) bool { return errors.Is(err, target) }

// As is a wrapper around errors.As to check if the error is of a specific type.
func As(err error, target any) bool { return errors.As(err, target) }

// Unwrap returns the underlying cause of this error, nil if no cause.
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Unwrap returns the underlying cause of this error, nil if no cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.Options.Cause != nil {
		return e.Options.Cause
	}
	return nil
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
		"%s (%s):",
		strings.ToLower(e.Title),
		concatenateUint32Slice(e.Options.Identifier),
	)

	if len(e.Options.Details) > 0 {
		fmt.Fprintf(
			b,
			" %s:",

			strings.Join(e.Options.Details, ": "),
		)
	}

	for k, v := range e.Options.Properties {
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

	if e.Options.Cause != nil {
		fmt.Fprintf(b, " caused by: %v", e.Options.Cause.Error())
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

func from(err error, copyStack bool, opts ...Option) *Error {
	var t *Error

	ok := errors.As(err, &t)
	if !ok {
		t, _ = New("unknown error").(*Error)
	}

	props := make(map[string]any)
	for k, v := range t.Options.Properties {
		props[k] = v
	}
	o := Opts{
		Details:    t.Options.Details,
		Properties: props,
		Cause:      t.Options.Cause,
	}
	if t.Options.Identifier != nil {
		o.Identifier = t.Options.Identifier
	}
	for _, opt := range opts {
		opt(&o)
	}
	e := &Error{
		Title:   t.Title,
		Options: o,
	}
	if copyStack {
		e.Stack = t.Stack
	}

	if !ok {
		e.Options.Cause = err
	}

	return e
}

func (e *Error) throw(trace *Trace) error {
	if trace == nil {
		return e
	}

	e.Stack = append([]*Trace{trace}, e.Stack...)

	return e
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

// concatenateUint32Slice takes a slice of uint32 and returns a single string
// with all elements joined by a hyphen ("-").
func concatenateUint32Slice(nums []uint32) string {
	if len(nums) == 0 {
		return ""
	}

	slices.Reverse(nums)

	// Use a strings.Builder for efficient string concatenation.
	var builder strings.Builder

	// Iterate over the slice elements reversly.
	for i, num := range nums {
		// Convert the int32 to its string representation.
		// We use base 10 (decimal) and specify 32-bit type for clarity,
		// though 'FormatInt' takes an int64 internally (int32 is safely converted).
		str := strconv.FormatInt(int64(num), 10)

		// Write the string representation to the builder.
		builder.WriteString(str)

		// Append the separator for all elements except the last one.
		if i < len(nums)-1 {
			builder.WriteString("-")
		}
	}

	// Return the final concatenated string.
	return builder.String()
}
