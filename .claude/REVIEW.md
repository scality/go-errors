# Code Review Criteria

When reviewing pull requests in this repository, check the following areas:

| Area | What to check |
| ---- | ------------- |
| Public API compatibility | Changes to exported identifiers, function signatures, the `Option` type or the `Error` struct fields. A breaking change needs a major version bump and a `/vN` module path. |
| Output formats | Changes to the `Error()` string, the `%v` / `%+v` JSON shape, JSON field names or the `slog` group keys. Consumers parse and match these, so treat them as API. `README.md` and `DESIGN.md` show the exact strings and must stay in sync. |
| Standard library semantics | `Is`, `As` and `Unwrap` keep working with the standard `errors` package. `Is` keeps its identifier-prefix matching. `CausedBy` chains stay reachable through `errors.Unwrap`. |
| Immutability of sentinels | Sentinel errors from `New` are shared package-level values. `Wrap` and every option must clone `Details`, `Identifier` and `Properties` and never mutate the template error. Watch for slice aliasing through `append`. |
| Concurrency | Sentinels are wrapped from many goroutines at once. No option may write to shared maps or slices. Tests run with `-race`. |
| Nil safety | Nil receivers, `Wrap(nil)`, `CausedBy(nil)`, empty property keys and nil maps must not panic. An error library must never crash the caller's error path. |
| Stack capture | `runtime.Caller` depth is hard-coded. A new entry point or an extra call layer shifts the captured frame. Tests should assert the reported function name. |
| Dependencies | The library has no runtime dependency. A new non-test import outside the standard library needs a strong reason. `go.mod` and `go.sum` stay consistent. |
| Tests | New behavior is covered by Ginkgo/Gomega specs in `errors_test.go`, in the existing `Describe` / `It` style. |
| CI and workflows | GitHub Actions are pinned to a commit SHA with a version comment, matching the existing workflows. |
| Security | No secrets or real hostnames in tests or examples. Flag changes that start logging `Properties` or `Details` in new places, since callers may put sensitive values there. |
