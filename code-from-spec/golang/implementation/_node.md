# SPEC/golang/implementation

Go implementation of the spec tree components.

# Public

## Implementation rules

- Write idiomatic Go: camelCase for local variables and
  parameters, exported names for public API, receiver
  methods where the interface specifies them.
- Wrap all errors with `fmt.Errorf` using `%w` so callers
  can match with `errors.Is()`.
- Write straightforward code. Simple and readable over
  clever and compact.
