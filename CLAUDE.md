# go-errors

This is a **Go library for structured error handling** (`github.com/scality/go-errors`). It contains:

- A single package `errors` in `errors.go`: `New`, `Wrap` with functional options, stack capture, identifier chaining
- Standard library compatibility: `errors.Is`, `errors.As`, `errors.Unwrap`, `fmt` verbs (`%v` / `%+v` as JSON) and `slog.LogValuer`
- Ginkgo/Gomega tests in `errors_test.go`, run with `go test -race ./...`
- Linting with golangci-lint, via the devcontainer in `.devcontainer/`
- Design notes in `DESIGN.md`, usage and exact output formats in `README.md`
- No runtime dependencies: `go.mod` only pulls test dependencies
