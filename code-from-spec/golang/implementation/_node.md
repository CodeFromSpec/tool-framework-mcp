# SPEC/golang/implementation

Go implementation of the functional specifications.

# Public

## Implementation rules

- Implement the pseudocode from the `input` artifact.
- Declare types, error sentinels, and function signatures
  exactly as specified in the interface artifact from
  `depends_on` — same names, same receiver types, same
  return types. The interface is the contract. The output
  file is the sole `.go` file in the package — it must
  contain all declarations from the interface.
- Use the package name declared in the interface artifact.
- Write idiomatic Go: camelCase for local variables and
  parameters, exported names for public API, receiver
  methods where the interface specifies them.
- Wrap all errors with `fmt.Errorf` using `%w` so callers
  can match with `errors.Is()`.
- Write straightforward code. Simple and readable over
  clever and compact.
