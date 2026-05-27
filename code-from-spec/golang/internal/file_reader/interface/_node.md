---
depends_on:
  - ROOT/functional/utils/file_reader(interface)
outputs:
  - id: interface
    path: code-from-spec/golang/internal/file_reader/interface/output.md
---

# ROOT/golang/internal/file_reader/interface

Generates the Go interface specification for the filereader
package: types, error sentinels, function signatures, and
usage examples.

# Agent

Translate the functional interface into a Go interface
specification for the `filereader` package.

Generate a markdown document with all Go code inside fenced
code blocks.

## Translation rules

- Records → Go structs. Use unexported fields unless the
  functional spec says the field is public.
- Errors → sentinel variables with `errors.New`.
- Functions that take a record as first argument → receiver
  methods on the pointer type.
- Constructor functions (return a record) → package-level
  functions.
- Return `error` from methods that can fail, even if the
  functional spec uses "raise error" without an explicit
  return.

## What to include

- Struct definitions
- Error sentinels
- Function and method signatures with doc comments
- Usage examples demonstrating the typical call pattern
