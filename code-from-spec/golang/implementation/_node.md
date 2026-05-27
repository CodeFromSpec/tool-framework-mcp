# ROOT/golang/implementation

Go implementation of the functional specifications.

# Public

## Implementation rules

- Implement the pseudocode from the `input` artifact.
- Declare types, error sentinels, and function signatures
  exactly as specified in the interface artifact from
  `depends_on`.
- Wrap all errors with `fmt.Errorf` using `%w` so callers
  can match with `errors.Is()`.
- Write straightforward code. Simple and readable over
  clever and compact.
