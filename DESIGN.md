# Introduction
This library aims to simplify error handling in Go through the following key features
* One unique function with argument for standard error and go-errors
* Variadic options following the "Functional Options" pattern
* An error code, obtained by concatenating identifiers defined at each calls of subfunctions, that allow tracing the root cause
* Comply with error interface (Error() string)

# One unique function
The signature of this unique function would be:
```go
func Wrap(error, errors.Opts...) error
```
Where error can be of type 
1. standard errors
2. go-errors

## standard errors
The provided error will be considered as a root error and stored as cause of a new go-errors one

## go-errors
The provided error will be considered as a template for a new go-errors one

# Variadic Options
The variadic options would be the following one

* `WithDetail`/`WithDetailf`: provide a message that details the error
* `WithProperty`: provide a key/value pair for additional informations (filename, path, username, ... )
* `WithIdentifier`: used to provide an identifer, it could be concatenated with other idenfier from previous calls from subfunctions (See example below for clarity)
* `CausedBy`: used to trace a root error

Signature could be as follow

```go
errors.WithDetail(string)
errors.WithDetailf(string, ...any)
errors.WithProperty(string, any)
errors.WithIdentifier(int)
errors.CausedBy(err error)
```

# Identifier

Below an example of using `go-errors` with identifier.  
Also, find the expected error message

```go
package main

import (
    "fmt"
    "os"

    errors "github.com/scality/go-errors"
)

var ErrForbidden = errors.New("forbidden")

func main(){
    err := call1()
    fmt.Println(err)
}

func call1() error {
	return errors.Wrap(
        call2(),
        errors.WithIdentifier(19),
    )
}

func call2() error {
	return errors.Wrap(
        call3(),
        errors.WithDetail("missing required role"),
        errors.WithProperty("Role", "Reader"),
        errors.WithProperty("User", "john.doe"),
        errors.WithIdentifier(12),
    )
}

func call3() error {
    // ls -l test.txt:
    // -rw------- 1 root root 5 Mar  5 08:18 test.txt
    _, err := os.Open("test.txt")

    // Something went wrong here
	return errors.Wrap(
        ErrForbidden,
        errors.WithIdentifer(2),
        errors.WithDetail("permission denied"),
		errors.WithProperty("File", "test.txt"),
        errors.CausedBy(err),
    )
}
```

In the following situation, the program would return
```
forbidden (19-12-2): permission denied: missing required role: Role='Reader', User='john.doe', File='test.txt', at=(func='main.call3', file='main.go', line='41'), caused by: open test.txt: permission denied
```