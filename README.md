# go-errors

A Go library for enhanced error handling with stack traces, structured error information, and error wrapping capabilities.

## Features

- **Stack Traces**: Automatic capture of function location, file, line number, and timestamp
- **Error Wrapping**: Chain errors with causes for better error context
- **Structured Errors**: Add identifiers, details, and arbitrary properties to errors
- **Method Chaining**: Fluent API for building detailed error information
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
    err := errors.New("database error").
        WithIdentifier(1001).
        WithDetail("connection timeout").
        WithProperty("host", "localhost").
        Throw()
    
    fmt.Println(err)
}
```

## Usage Examples

### Creating Basic Errors

```go
// Simple error
err := errors.New("validation failed").Throw()

// Error with identifier
err := errors.New("not found").
    WithIdentifier(404).
    Throw()
```

### Adding Details and Properties

```go
// Multiple details
// Each WithDetail() call appends to the Details slice
err := errors.New("validation failed").
    WithDetail("email is required").
    WithDetail("password must be at least 8 characters").
    Throw()

// Details can also be formatted
err := errors.New("request failed").
    WithDetailf("failed to connect to %s:%d", "api.example.com", 443).
    WithDetail("timeout after 30 seconds").
    Throw()

// Single property
err := errors.New("request failed").
    WithProperty("url", "https://api.example.com").
    WithProperty("status_code", 500).
    Throw()

// Multiple properties at once
err := errors.New("database error").
    WithProperties(map[string]any{
        "host": "localhost",
        "port": 5432,
        "database": "myapp",
    }).
    Throw()
```

### Error Wrapping

```go
func getUserByID(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, errors.New("user not found").
            WithIdentifier(404).
            CausedBy(err).
            Throw()
    }
    return user, nil
}

// Convenient wrapping with Wrap() - adds message and stack trace
func getUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to fetch user from database")
    }
    return user, nil
}

// Convenient wrapping with Wrapf() - adds formatted message and stack trace
func getUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, errors.Wrapf(err, "failed to fetch user with id %d", id)
    }
    return user, nil
}
```

### Stack Traces

```go
func layer3() error {
    return errors.New("something went wrong").Throw()
}

func layer2() error {
    err := layer3()
    if err != nil {
        return errors.Stamp(err) // Adds layer2's location to stack
    }
    return nil
}

func layer1() error {
    err := layer2()
    if err != nil {
        return errors.Stamp(err) // Adds layer1's location to stack
    }
    return nil
}

// The error will contain a complete stack trace through all layers
```

### Working with Standard Library

```go
// Using errors.Is for comparison
notFoundErr := errors.New("not found").WithIdentifier(404)

if errors.Is(err, notFoundErr) {
    // Handle not found error
}

// Using errors.As for type assertion
var e *errors.Error
if errors.As(err, &e) {
    fmt.Printf("Error ID: %d\n", e.Identifier)
    fmt.Printf("Details: %v\n", e.Details) // []string
    fmt.Printf("Properties: %v\n", e.Properties)
    
    // Access individual details
    for i, detail := range e.Details {
        fmt.Printf("  Detail %d: %s\n", i, detail)
    }
}

// Unwrapping errors
cause := errors.Unwrap(err)
```

### Converting Standard Errors

```go
// Using From() to convert any error
stdErr := fmt.Errorf("something went wrong")
err := errors.From(stdErr).
    WithDetail("additional context").
    Throw()

// Using Intercept() when you're unsure of the error type
func handleError(err error) error {
    e := errors.Intercept(err)
    e.WithProperty("handled_at", time.Now())
    return e.Throw()
}
```

## Output Format

The `Error()` method produces output in the following format:

```
title (id): detail1: detail2: detail3: key1='value1', key2='value2', at=(func='funcName', file='file.go', line='10'), caused by: underlying error
```

**Note**: Details are stored as a slice and joined with `: ` when the error is formatted. Each call to `WithDetail()` or `WithDetailf()` appends to this slice.

Example with multiple details:
```
database error (1001): connection timeout: retry limit exceeded: host='localhost', port='5432', at=(func='connectDB', file='db.go', line='42'), caused by: dial tcp: connection refused
```

Example with wrapped error:
```
unknown error (0): failed to fetch user from database: at=(func='getUser', file='user.go', line='25'), caused by: connection refused
```

## Best Practices

1. **Always use `Throw()`** when returning errors to capture stack traces
2. **Use `Stamp()`** when passing errors up the call stack to track the error path
3. **Use `Wrap()` or `Wrapf()`** for convenient error wrapping with automatic stack traces
4. **Use identifiers** for errors that need programmatic handling (e.g., HTTP status codes)
5. **Add multiple details** using `WithDetail()` or `WithDetailf()` - they're stored as a slice and displayed in order
6. **Add properties** for debugging context (IDs, URLs, parameters, etc.)
7. **Wrap errors with `CausedBy()`** to maintain error chains for programmatic inspection
8. **Use `From()`** when you need to enhance third-party errors with additional context
9. **Use `Intercept()`** when you need to add context to errors from existing errors

### Details vs Properties

- **Details** (slice of strings): Human-readable context that appears in error messages, ordered and concatenated with `: `
- **Properties** (key-value map): Structured data for debugging/logging, useful for searching and filtering logs

## License

See [LICENSE](LICENSE) file for details.
