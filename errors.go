package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

const separator = "-"

// Trace represents a single entry in an error's stack trace.
// It captures the location and time when an error was thrown or stamped.
type Trace struct {
	// The function where the error occurred
	Function string `json:"function,omitempty" yaml:"function,omitempty"`

	// The file where the error occurred
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// The line where the error occurred
	Line int `json:"line,omitempty" yaml:"line,omitempty"`
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
	opts `json:",inline" yaml:",inline"`
	// A trace of where the error was thrown
	stack []*Trace
}

type causeError struct {
	error
}

type opts struct {
	// A numeric error code for programmatic handling
	Identifier []uint32 `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	// Additional context as a list of strings
	Details []string `json:"details,omitempty" yaml:"details,omitempty"`
	// Key-value pairs for arbitrary metadata
	Properties map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`
	// The underlying error that caused this error
	Cause *causeError `json:"cause,omitempty" yaml:"cause,omitempty"`
}

type errorWithStack struct {
	Error
	Stack []*Trace `json:"stack,omitempty" yaml:"stack,omitempty"`
}

func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil {
		return
	}
	switch verb {
	case 'v':
		if s.Flag('+') {
			json.NewEncoder(s).Encode(errorWithStack{Error: *e, Stack: e.stack}) //nolint:errcheck
			return
		}
		json.NewEncoder(s).Encode(e) //nolint:errcheck
	case 's', 'q':
		io.WriteString(s, e.Error()) //nolint:errcheck
	}
}

func (c causeError) MarshalJSON() ([]byte, error) {
	if t, ok := c.error.(*Error); ok {
		return json.Marshal(t)
	}
	return json.Marshal(c.Error())
}

// New creates a new Error with the given title.
func New(title string) error {
	return &Error{
		Title: title,
		opts: opts{
			Details:    []string{},
			Properties: make(map[string]any),
		},
	}
}

// Option is a function type that modifies the Options struct.
type Option func(*opts)

// WithIdentifier sets a numeric identifier for the error.
func WithIdentifier(id uint32) Option {
	return func(c *opts) {
		c.Identifier = append(c.Identifier, id)
	}
}

// WithDetail sets a detail string for the error.
func WithDetail(msg string) Option {
	return func(c *opts) {
		c.Details = append(c.Details, msg)
	}
}

// WithDetailf sets a detail string for the error using a format string.
func WithDetailf(format string, args ...any) Option {
	return func(c *opts) {
		c.Details = append(c.Details, fmt.Sprintf(format, args...))
	}
}

// WithProperty sets a property for the error.
func WithProperty(key string, value any) Option {
	if key == "" {
		return func(c *opts) {}
	}
	return func(c *opts) {
		c.Properties[key] = value
	}
}

// WithProperties sets multiple properties for the error.
func WithProperties(properties map[string]any) Option {
	if properties == nil {
		return func(c *opts) {}
	}
	return func(c *opts) {
		maps.Copy(c.Properties, properties)
	}
}

// CausedBy sets the underlying cause of this error.
func CausedBy(err error) Option {
	return func(c *opts) {
		c.Cause = &causeError{error: err}
	}
}

// Wrap wraps an error with a message
// returns an "unknown error" when err is nil
func Wrap(err error, opts ...Option) error {
	trace := trace()
	return from(err, true, opts...).throw(trace)
}

// Is compares this error with another error for equality.
// Two errors match if they have same Title and same Identifier*
// (*) or if given argument is a parent of the other.
func (e *Error) Is(err error) bool {
	other := new(Error)
	if ok := errors.As(err, &other); !ok {
		return false
	}

	if e.Title != other.Title {
		return false
	}
	a, b := e.Identifier, other.Identifier
	if len(a) < len(b) {
		return false
	}
	return slices.Equal(b, a[:len(b)])
}

// Is is a wrapper around errors.Is to compare two errors for equality.
func Is(err, target error) bool { return errors.Is(err, target) }

// As is a wrapper around errors.As to check if the error is of a specific type.
func As(err error, target any) bool { return errors.As(err, target) }

// IdentifierStartsWith checks if the error's identifier string starts with the given prefix.
func IdentifierStartsWith(err error, prefix string) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}

	if prefix == "" {
		return true
	}

	// To avoid uint32 conversion and errors handling, we decided to compare:
	// * the identifier, suffixed with an hyphen
	// * the prefix, suffixed with an hyphen
	return strings.HasPrefix(
		e.GetIdentifier()+separator,
		prefix+separator,
	)
}

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
	if e.Cause != nil {
		return e.Cause.error
	}
	return nil
}

// Error returns a human-readable representation of the error: title,
// identifier, details, properties and cause. The stack trace is
// intentionally NOT included here — Error() (and the %s/%q verbs) is the
// message form consumed by CLIs, %w chains and plain logging, where a
// stack is noise. Use the %+v verb for the full message-plus-stack dump.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	b := bytes.NewBuffer(nil)

	if id := e.GetIdentifier(); id != "" {
		fmt.Fprintf(b, "%s (%s)", strings.ToLower(e.Title), id)
	} else {
		fmt.Fprint(b, strings.ToLower(e.Title))
	}

	if len(e.Details) > 0 {
		fmt.Fprintf(b, ": %s", strings.Join(e.Details, ": "))
	}

	if len(e.Properties) > 0 {
		// order by keys for deterministic output
		keys := make([]string, 0, len(e.Properties))
		for k := range e.Properties {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		props := make([]string, 0, len(keys))
		for _, k := range keys {
			props = append(props, fmt.Sprintf("%s='%v'", k, e.Properties[k]))
		}
		fmt.Fprintf(b, ": %s", strings.Join(props, ", "))
	}

	if e.Cause != nil {
		fmt.Fprintf(b, ", caused by: %v", e.Cause.Error())
	}

	return b.String()
}

func from(err error, copyStack bool, options ...Option) *Error {
	var t *Error

	ok := errors.As(err, &t)
	if !ok {
		t, _ = New("unknown error").(*Error)
	}

	props := make(map[string]any, len(t.Properties))
	maps.Copy(props, t.Properties)
	o := opts{
		Details:    slices.Clone(t.Details),
		Properties: props,
		Cause:      t.Cause,
	}
	if t.Identifier != nil {
		o.Identifier = slices.Clone(t.Identifier)
	}
	for _, opt := range options {
		opt(&o)
	}
	e := &Error{
		Title: t.Title,
		opts:  o,
	}
	if copyStack {
		e.stack = t.stack
	}

	if !ok && err != nil {
		e.Cause = &causeError{error: err}
	}

	return e
}

func (e *Error) throw(trace *Trace) error {
	if trace == nil {
		return e
	}

	e.stack = append([]*Trace{trace}, e.stack...)

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
		Function: fn.Name(),
		File:     file,
		Line:     line,
	}
}

// GetIdentifier returns a string with all identifiers reversed and joined by a hyphen (-).
func (e *Error) GetIdentifier() string {
	if len(e.Identifier) == 0 {
		return ""
	}

	// Clone to avoid modifying the original slice.
	clone := slices.Clone(e.Identifier)
	slices.Reverse(clone)

	// Use a strings.Builder for efficient string concatenation.
	var builder strings.Builder

	// Iterate over the slice elements reversly.
	for i, cl := range clone {
		// Convert the uint32 to its string representation.
		// We use base 10 (decimal) and specify 32-bit type for clarity,
		// though 'FormatInt' takes an int64 internally (uint32 is safely converted).
		str := strconv.FormatInt(int64(cl), 10)

		// Write the string representation to the builder.
		builder.WriteString(str)

		// Append the separator for all elements except the last one.
		if i < len(clone)-1 {
			builder.WriteString(separator)
		}
	}

	// Return the final concatenated string.
	return builder.String()
}
