---
depends_on:
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/os/path_utils
output: internal/logicalnames/logicalnames_test.go
---

# SPEC/golang/tests/utils/logical_names

Unit tests for the logicalnames package.

# Agent

## Context

Pure function tests — no filesystem or temp directories
needed. Each test calls the function with a string input
and asserts the output.

## Test cases

### LogicalNameToPath

#### SPEC alone

Input: "SPEC". Expect: PathCfs
`code-from-spec/_node.md`.

#### SPEC with path

Input: "SPEC/payments/processor". Expect: PathCfs
`code-from-spec/payments/processor/_node.md`.

#### Strips qualifier before resolving

Input: "SPEC/x/y(interface)". Expect: PathCfs
`code-from-spec/x/y/_node.md`.

#### Rejects ROOT reference

Input: "ROOT/x". Expect error ErrUnsupportedReference.

#### Rejects ARTIFACT reference

Input: "ARTIFACT/x". Expect error
ErrUnsupportedReference.

#### Rejects EXTERNAL reference

Input: "EXTERNAL/proto/api.proto". Expect error
ErrUnsupportedReference.

#### Rejects unrecognized prefix

Input: "UNKNOWN/something". Expect error
ErrUnsupportedReference.

#### Rejects empty string

Input: "". Expect error ErrUnsupportedReference.

### LogicalNameFromPath

#### Root node

Input: PathCfs `code-from-spec/_node.md`.
Expect: "SPEC".

#### Nested node

Input: PathCfs `code-from-spec/x/y/_node.md`.
Expect: "SPEC/x/y".

#### Rejects non-node path

Input: PathCfs `internal/config/config.go`.
Expect error ErrInvalidPath.

#### Rejects path without _node.md

Input: PathCfs `code-from-spec/x/y/output.md`.
Expect error ErrInvalidPath.

### LogicalNameGetParent

#### SPEC/x parent is SPEC

Input: "SPEC/domain". Expect: "SPEC".

#### SPEC/x/y parent is SPEC/x

Input: "SPEC/domain/config". Expect: "SPEC/domain".

#### Strips qualifier before computing parent

Input: "SPEC/domain/config(interface)".
Expect: "SPEC/domain".

#### SPEC has no parent

Input: "SPEC". Expect error ErrNoParent.

#### Rejects ROOT reference

Input: "ROOT/domain". Expect error
ErrNotASpecReference.

#### Rejects ARTIFACT reference

Input: "ARTIFACT/x". Expect error
ErrNotASpecReference.

#### Rejects EXTERNAL reference

Input: "EXTERNAL/x". Expect error
ErrNotASpecReference.

### LogicalNameGetQualifier

#### Extracts qualifier from SPEC reference

Input: "SPEC/x/y(interface)". Expect: "interface",
true.

#### ARTIFACT without qualifier returns absent

Input: "ARTIFACT/x/y". Expect: "", false.

#### EXTERNAL without qualifier returns absent

Input: "EXTERNAL/proto/api.proto". Expect: "", false.

#### Returns absent when no qualifier

Input: "SPEC/x/y". Expect: "", false.

#### Returns absent for SPEC alone

Input: "SPEC". Expect: "", false.

### LogicalNameStripQualifier

#### Strips qualifier from SPEC reference

Input: "SPEC/x/y(interface)". Expect: "SPEC/x/y".

#### ARTIFACT without qualifier — returns unchanged

Input: "ARTIFACT/x/y". Expect: "ARTIFACT/x/y".

#### EXTERNAL — returns unchanged

Input: "EXTERNAL/proto/api.proto".
Expect: "EXTERNAL/proto/api.proto".

#### No qualifier — returns unchanged

Input: "SPEC/x/y". Expect: "SPEC/x/y".

#### SPEC alone — returns unchanged

Input: "SPEC". Expect: "SPEC".

#### Empty string — returns unchanged

Input: "". Expect: "".

### LogicalNameHasParent

#### SPEC alone

Input: "SPEC". Expect: false.

#### SPEC with path

Input: "SPEC/domain/config". Expect: true.

#### ARTIFACT reference

Input: "ARTIFACT/x". Expect: false.

#### EXTERNAL reference

Input: "EXTERNAL/x". Expect: false.

#### Empty string

Input: "". Expect: false.

### LogicalNameHasQualifier

#### Without qualifier

Input: "SPEC/x". Expect: false.

#### With qualifier

Input: "SPEC/x(y)". Expect: true.

#### ARTIFACT without qualifier

Input: "ARTIFACT/x". Expect: false.

#### EXTERNAL without qualifier

Input: "EXTERNAL/x". Expect: false.

#### SPEC alone

Input: "SPEC". Expect: false.

#### Empty string

Input: "". Expect: false.

### LogicalNameIsArtifact

#### ARTIFACT reference

Input: "ARTIFACT/x". Expect: true.

#### SPEC reference

Input: "SPEC/x(y)". Expect: false.

#### EXTERNAL reference

Input: "EXTERNAL/x". Expect: false.

#### Empty string

Input: "". Expect: false.

### LogicalNameIsSpec

#### SPEC alone

Input: "SPEC". Expect: true.

#### SPEC with path

Input: "SPEC/x/y". Expect: true.

#### ROOT reference — not SPEC

Input: "ROOT/x". Expect: false.

#### ARTIFACT reference

Input: "ARTIFACT/x". Expect: false.

#### EXTERNAL reference

Input: "EXTERNAL/x". Expect: false.

#### Empty string

Input: "". Expect: false.

### LogicalNameIsExternal

#### EXTERNAL reference

Input: "EXTERNAL/proto/api.proto". Expect: true.

#### SPEC reference

Input: "SPEC/x". Expect: false.

#### ARTIFACT reference

Input: "ARTIFACT/x". Expect: false.

#### Empty string

Input: "". Expect: false.

### LogicalNameGetArtifactGenerator

#### Simple artifact

Input: "ARTIFACT/x". Expect: "SPEC/x".

#### Nested artifact

Input: "ARTIFACT/x/y/z". Expect: "SPEC/x/y/z".

#### Rejects SPEC reference

Input: "SPEC/x(y)". Expect error
ErrNotAnArtifactReference.

#### Rejects EXTERNAL reference

Input: "EXTERNAL/x". Expect error
ErrNotAnArtifactReference.

### LogicalNameExternalToPath

#### Simple path

Input: "EXTERNAL/proto/v1/api.proto".
Expect: PathCfs `proto/v1/api.proto`.

#### Root-level file

Input: "EXTERNAL/docker-compose.yaml".
Expect: PathCfs `docker-compose.yaml`.

#### Rejects SPEC reference

Input: "SPEC/x". Expect error
ErrNotAnExternalReference.

#### Rejects ARTIFACT reference

Input: "ARTIFACT/x". Expect error
ErrNotAnExternalReference.

## Go-specific guidance

- The package name is `logicalnames_test` (external
  test package).
- Pure function tests — no file I/O needed.
