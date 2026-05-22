---
outputs:
  - id: logicalnames_test
    path: internal/logicalnames/logicalnames_test.go
---

# ROOT/golang/internal/logical_names/tests

Unit tests for the logicalnames package.

# Agent

## Context

Pure function tests — no filesystem or temp directories
needed. Each test calls the function with a string input
and asserts the output.

## PathFromLogicalName

### ROOT

Input: `"ROOT"`
Expect: `"code-from-spec/_node.md"`, `true`.

### ROOT with path

Input: `"ROOT/payments/processor"`
Expect: `"code-from-spec/payments/processor/_node.md"`, `true`.

### ROOT with qualifier

Input: `"ROOT/payments/processor(interface)"`
Expect: `"code-from-spec/payments/processor/_node.md"`, `true`.

### ROOT with qualifier — strips qualifier from path

Input: `"ROOT/x(y)"`
Expect: `"code-from-spec/x/_node.md"`, `true`.

### ARTIFACT reference returns false

Input: `"ARTIFACT/x(y)"`
Expect: `""`, `false`.

### Unrecognized prefix

Input: `"UNKNOWN/something"`
Expect: `""`, `false`.

### Empty string

Input: `""`
Expect: `""`, `false`.

## HasParent

### ROOT

Input: `"ROOT"`
Expect: `false`, `true`.

### ROOT with path

Input: `"ROOT/domain/config"`
Expect: `true`, `true`.

### ROOT with qualifier

Input: `"ROOT/domain/config(interface)"`
Expect: `true`, `true`.

### ARTIFACT returns false false

Input: `"ARTIFACT/x(y)"`
Expect: `false`, `false`.

### Empty string

Input: `""`
Expect: `false`, `false`.

## ParentLogicalName

### ROOT/x — parent is ROOT

Input: `"ROOT/domain"`
Expect: `"ROOT"`, `true`.

### ROOT/x/y — parent is ROOT/x

Input: `"ROOT/domain/config"`
Expect: `"ROOT/domain"`, `true`.

### ROOT/x/y(z) — parent is ROOT/x

Input: `"ROOT/domain/config(interface)"`
Expect: `"ROOT/domain"`, `true`.

### ROOT has no parent

Input: `"ROOT"`
Expect: `""`, `false`.

### Empty string invalid

Input: `""`
Expect: `""`, `false`.

## HasQualifier

### ROOT without qualifier

Input: `"ROOT/x"`
Expect: `false`, `true`.

### ROOT with qualifier

Input: `"ROOT/x(y)"`
Expect: `true`, `true`.

### ARTIFACT with qualifier

Input: `"ARTIFACT/x(y)"`
Expect: `true`, `true`.

### ROOT alone

Input: `"ROOT"`
Expect: `false`, `true`.

### Empty string

Input: `""`
Expect: `false`, `false`.

## QualifierName

### ROOT with qualifier

Input: `"ROOT/x(y)"`
Expect: `"y"`, `true`.

### ROOT with nested path and qualifier

Input: `"ROOT/x/y(interface)"`
Expect: `"interface"`, `true`.

### ARTIFACT with qualifier

Input: `"ARTIFACT/x(y)"`
Expect: `"y"`, `true`.

### ROOT without qualifier

Input: `"ROOT/x"`
Expect: `""`, `false`.

### ROOT alone

Input: `"ROOT"`
Expect: `""`, `false`.

### Empty string

Input: `""`
Expect: `""`, `false`.

## IsArtifactRef

### ARTIFACT reference

Input: `"ARTIFACT/x(y)"`
Expect: `true`.

### ROOT reference

Input: `"ROOT/x(y)"`
Expect: `false`.

### Empty string

Input: `""`
Expect: `false`.

## ArtifactRefParts

### ARTIFACT/x(y)

Input: `"ARTIFACT/x(y)"`
Expect: `"code-from-spec/x/_node.md"`, `"y"`, `true`.

### ARTIFACT/x/y(z)

Input: `"ARTIFACT/x/y(z)"`
Expect: `"code-from-spec/x/y/_node.md"`, `"z"`, `true`.

### ARTIFACT without qualifier returns false

Input: `"ARTIFACT/x"`
Expect: `""`, `""`, `false`.

### ROOT reference returns false

Input: `"ROOT/x(y)"`
Expect: `""`, `""`, `false`.
