---
depends_on:
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/tests/utils/logical_names(logical_names_tests)
outputs:
  - id: logicalnames_test
    path: internal/logicalnames/logicalnames_test.go
---

# ROOT/golang/tests/utils/logical_names

Unit tests for the logicalnames package.

# Agent

## Context

Pure function tests — no filesystem or temp directories
needed. Each test calls the function with a string input
and asserts the output.
