---
output: internal/testutils/helpers.go
---

# SPEC/golang/test/utils/helpers

Small utility functions for tests that don't warrant
their own file.

# Public

## Package

`package testutils`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"`

```go
func Ptr[T any](v T) *T
```

Returns a pointer to `v`. Useful for constructing
`*string`, `*int`, and other pointer fields in test
fixtures without a temporary variable.

# Agent

## Ownership

This file declares and implements:
- Functions: `Ptr`

## Reference implementation

```go
func Ptr[T any](v T) *T { return &v }
```
