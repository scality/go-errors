# go-errors

A Go library for enhanced error handling with stack traces, structured error information, and error wrapping capabilities.

## Features

- **Stack Traces**: Automatic capture of function location, file and line number
- **Error Wrapping**: Chain errors with causes for better error context
- **Structured Errors**: Add identifiers, details, and arbitrary properties via options
- **Single Entry Point**: `Wrap(error, ...Option)` works with both standard errors and go-errors
- **Standard Library Compatible**: Implements standard error interface and works with `errors.Is()`, `errors.As()`, and `errors.Unwrap()`
- **JSON Serialization**: Built-in JSON marshaling for logging and debugging

## Installation

```bash
go get github.com/scality/go-errors
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/scality/go-errors"
)

func main() {
	var ErrDB = errors.New("database error")
	err := errors.Wrap(ErrDB,
		errors.WithIdentifier(1001),
		errors.WithDetail("connection timeout"),
		errors.WithProperty("host", "localhost"),
	)
	fmt.Println(err)
}
```

## Usage Examples

### Adding Details and Properties

```go
// Define domain errors with New()
var (
	ErrValidationFailed = errors.New("validation failed")
	ErrRequestFailed    = errors.New("request failed")
	ErrDatabaseError    = errors.New("database error")
)

// Multiple details: each WithDetail() / WithDetailf() appends to the details slice
err1 := errors.Wrap(ErrValidationFailed,
	errors.WithDetail("email is required"),
	errors.WithDetail("password must be at least 8 characters"),
)

// Formatted details
err2 := errors.Wrap(ErrRequestFailed,
	errors.WithDetailf("failed to connect to %s:%d", "api.example.com", 443),
	errors.WithDetail("timeout after 30 seconds"),
)

// Properties (key-value pairs)
err3 := errors.Wrap(ErrRequestFailed,
	errors.WithProperty("url", "https://api.example.com"),
	errors.WithProperty("status_code", 500),
)

// Multiple properties (one option per property)
err4 := errors.Wrap(ErrDatabaseError,
	errors.WithProperty("host", "localhost"),
	errors.WithProperty("port", 5432),
	errors.WithProperty("database", "myapp"),
)
```

### CausedBy

```go
var ErrUserNotFound = errors.New("user not found")

func getUserByID(id string) (*User, error) {
	user, err := db.Query(id)
	if err != nil {
		return nil, errors.Wrap(ErrUserNotFound,
			errors.WithIdentifier(404000),
			errors.CausedBy(err),
		)
	}
	return user, nil
}

// Convenient wrapping with Wrap() - adds message and stack trace
func getUser(id int) (*User, error) {
	user, err := db.Query(id)
	if err != nil {
		return nil, errors.Wrap(err,
			errors.WithDetail("failed to fetch user from database"),
		)
	}
	return user, nil
}
```

### Identifier Concatenation

Each call to `Wrap` can add an identifier segment; segments are concatenated with `-` to trace the error path through the call stack (e.g. `19-12-2`).

```go
var ErrForbidden = errors.New("forbidden")

func call1() error {
	return errors.Wrap(call2(), errors.WithIdentifier(19))
}

func call2() error {
	return errors.Wrap(call3(),
		errors.WithDetail("missing required role"),
		errors.WithProperty("Role", "Reader"),
		errors.WithIdentifier(12),
	)
}

func call3() error {
	_, err := os.Open("test.txt")
	return errors.Wrap(ErrForbidden,
		errors.WithIdentifier(2),
		errors.WithDetail("permission denied"),
		errors.WithProperty("File", "test.txt"),
		errors.CausedBy(err),
	)
}
```

### Working with Standard Library

```go
// Using errors.Is for comparison (compares Title and Identifier)
if errors.Is(err, notFoundErr) {
	// Handle not found error
}

// Using errors.As to access structured fields
var e *errors.Error
if errors.As(err, &e) {
	fmt.Printf("Title: %s\n", e.Title)
	fmt.Printf("Identifier: %v\n", e.Identifier)   // []uint32
	fmt.Printf("Details: %v\n", e.Details)         // []string
	fmt.Printf("Properties: %v\n", e.Properties)

	for i, detail := range e.Details {
		fmt.Printf("  Detail %d: %s\n", i, detail)
	}
}
```

### Converting Standard Errors

```go
// Wrap any standard error to add stack trace, details, and properties
stdErr := fmt.Errorf("something went wrong")
err := errors.Wrap(stdErr,
	errors.WithDetail("additional context"),
	errors.WithProperty("source", "legacy"),
)
```

### Specific use-case with Is()

Is() compares this error with another error for equality. Two errors match if they have same Title and same Identifier*
(*) or if one is a parent of the other.

For example:
If e1.Identifier: "2-1" and e2.Identifier: "3-2-1", then
```go
e2.Is(e1) return True // e1 is a parent of e2
e1.Is(e2) return False
```

## Output Format

The `Error()` method produces output in the following format:

```
title (id): detail1: detail2: detail3: key1='value1', key2='value2', at=[(func='func1Name', file='file.go', line='21'), (func='func2Name', file='file.go', line='10')], caused by: underlying error
```

Details are stored as a slice and joined with `: ` when the error is formatted. Each `WithDetail()` or `WithDetailf()` option appends to this slice.

Example with multiple details:

```
database error (1001): connection timeout: retry limit exceeded: host='localhost', port='5432', at=[(func='connectDB', file='db.go', line='42')], caused by: dial tcp: connection refused
```

Example with wrapped error:

```
unknown error (0): failed to fetch user from database: at=[(func='getUser', file='user.go', line='25')], caused by: connection refused
```

## JSON formatting message
| Marker | Description                |
| ------ | -------------------------- |
| `%v`   | JSON (without stack)       |
| `%+v`  | Extended JSON (with stack) |

Example with JSON without stack (%v):

```
{"title":"forbidden","identifier":[2,12,19],"details":["permission denied","missing required role"],"properties":{"File":"test.txt","Role":"Reader"},"cause":"open test.txt: permission denied"}
```

Example with extended JSON, with stack (%+v):

```
{"title":"forbidden","identifier":[2,12,19],"details":["permission denied","missing required role"],"properties":{"File":"test.txt","Role":"Reader"},"cause":"open test.txt: permission denied","stack":[{"function":"main.call1","file":"/path/to/main.go","line":25},{"function":"main.call2","file":"/path/to/main.go","line":29},{"function":"main.call3","file":"/path/to/main.go","line":38}]}
```

## Best Practices

1. **Use `Wrap(err, ...options)`** as the single entry point for both standard errors and go-errors; it captures stack traces and applies options.
2. **Use identifiers** for errors that need programmatic handling (e.g., HTTP status codes); they are concatenated across the call stack when re-wrapping.
3. **Add details** with `WithDetail()` or `WithDetailf()`; they are stored as a slice and displayed in order.
4. **Add properties** with `WithProperty(key, value)` for debugging context (IDs, URLs, parameters, etc.).
5. **Use `CausedBy(err)`** when wrapping to record the underlying cause and maintain error chains for `errors.Is` / `errors.Unwrap`.
6. **Use `New(title)`** to define sentinel errors; pass them to `Wrap` and add options at each layer.

### Details vs Properties

- **Details** (slice of strings): Human-readable context that appears in error messages, ordered and concatenated with `: `.
- **Properties** (key-value map): Structured data for debugging/logging, useful for searching and filtering logs.

## License

See [LICENSE](LICENSE) file for details.
