# ROOT/golang/interfaces

Go interface specifications generated from functional
interfaces. Each leaf node translates a functional interface
into Go types, error sentinels, function signatures, and
usage examples.

# Public

## Translation rules

- Records → Go structs. Use unexported fields unless the
  functional spec says the field is public. The struct
  itself must be exported if it is returned by a function
  or used as a parameter by callers of the interface.
- Errors → sentinel variables with `errors.New`.
- Functions that take a record as first argument → receiver
  methods on the pointer type.
- Constructor functions (return a record) → package-level
  functions.
- Return `error` from methods that can fail, even if the
  functional spec uses "raise error" without an explicit
  return.

## Output format

A markdown document with all Go code inside fenced code
blocks. Include:

- Struct definitions
- Error sentinels
- Function and method signatures with doc comments
- Usage examples demonstrating the typical call pattern
