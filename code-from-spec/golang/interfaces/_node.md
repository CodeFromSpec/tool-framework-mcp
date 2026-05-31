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
- Errors → sentinel variables with `errors.New`. The
  functional spec lists errors in PascalCase (e.g.
  `FileUnreadable`, `NoOutputs`). Export a sentinel
  variable for each: `Err` + the error name (e.g.
  `ErrFileUnreadable`, `ErrNoOutputs`). Errors marked
  as `(Module.*): propagated from ...` are owned by
  another package — do not re-declare them locally.
- Constructor functions (return a record) → package-level
  functions.
- Return `error` from methods that can fail, even if the
  functional spec uses "raise error" without an explicit
  return.
- With the exception of primitive types, and unless you
  have a compelling reason not to, pass and return all
  types by reference, not by value.

## Output format

A markdown document with all Go code inside fenced code
blocks. Include:

- Package declaration (from the leaf node's `## Package`)
- Import path (from the leaf node's `## Import`)
- Struct definitions (exported, no unexported fields)
- Error sentinels
- Function and method signatures with doc comments
- Usage examples demonstrating the typical call pattern,
  using the full import path in import statements
