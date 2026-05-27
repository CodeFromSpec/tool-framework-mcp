# ROOT/golang/interfaces

Go interface specifications generated from functional
interfaces. Each leaf node translates a functional interface
into Go types, error sentinels, function signatures, and
usage examples.

# Public

## Translation rules

- Records → Go structs. The struct itself must be exported
  if it is returned by a function or used as a parameter
  by callers. Unexported fields are implementation details
  and should not appear in the interface specification.
- Errors → sentinel variables with `errors.New`.
- Functions that take a record as first argument → receiver
  methods on the pointer type.
- Constructor functions (return a record) → package-level
  functions returning a pointer to the struct when the
  struct has receiver methods.
- Return `error` from methods that can fail, even if the
  functional spec uses "raise error" without an explicit
  return.

## Output format

A markdown document with all Go code inside fenced code
blocks. Include:

- Package declaration (from the leaf node's `## Package`)
- Struct definitions (exported, no unexported fields)
- Error sentinels
- Function and method signatures with doc comments
- Usage examples demonstrating the typical call pattern
