---
depends_on:
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/utils/logical_names
output: internal/logicalnames/logicalnames_test.go
---

# SPEC/golang/tests/utils/logical_names

Unit tests for the logicalnames package.

# Agent

## Context

Tests for `LogicalNameParse` and `LogicalNameFromPath`.

SPEC and EXTERNAL tests are pure string parsing — no
filesystem needed. ARTIFACT tests require a temp
directory with a `_node.md` file containing frontmatter
with an `output` field, because `LogicalNameParse` reads
the generator's frontmatter.

ARTIFACT tests use the `testChdir` pattern: create a
temp dir, chdir to it, create the necessary
`code-from-spec/.../_node.md` files, then call
`LogicalNameParse`.

## Test cases

### LogicalNameParse — SPEC type

#### SPEC alone

Input: "SPEC".
Expect: Type = NodeTypeSpec, Name = "SPEC",
Qualifier = nil, Path = "code-from-spec/_node.md",
Parent = nil.

#### SPEC with single segment

Input: "SPEC/domain".
Expect: Type = NodeTypeSpec, Name = "SPEC/domain",
Qualifier = nil,
Path = "code-from-spec/domain/_node.md",
Parent = pointer to "SPEC".

#### SPEC with nested path

Input: "SPEC/payments/fees/calculation".
Expect: Type = NodeTypeSpec,
Name = "SPEC/payments/fees/calculation",
Qualifier = nil,
Path = "code-from-spec/payments/fees/calculation/_node.md",
Parent = pointer to "SPEC/payments/fees".

#### SPEC with qualifier

Input: "SPEC/x/y(interface)".
Expect: Type = NodeTypeSpec, Name = "SPEC/x/y",
Qualifier = pointer to "interface",
Path = "code-from-spec/x/y/_node.md",
Parent = pointer to "SPEC/x".

#### SPEC with qualifier — root level

Input: "SPEC(context)".
Expect: Type = NodeTypeSpec, Name = "SPEC",
Qualifier = pointer to "context",
Path = "code-from-spec/_node.md",
Parent = nil.

#### SPEC with qualifier — parent is computed from unqualified name

Input: "SPEC/domain/config(interface)".
Expect: Type = NodeTypeSpec,
Name = "SPEC/domain/config",
Qualifier = pointer to "interface",
Path = "code-from-spec/domain/config/_node.md",
Parent = pointer to "SPEC/domain".

### LogicalNameParse — EXTERNAL type

#### Simple external path

Input: "EXTERNAL/proto/v1/api.proto".
Expect: Type = NodeTypeExternal,
Name = "EXTERNAL/proto/v1/api.proto",
Qualifier = nil, Path = "proto/v1/api.proto",
Parent = nil.

#### Root-level external file

Input: "EXTERNAL/docker-compose.yaml".
Expect: Type = NodeTypeExternal,
Name = "EXTERNAL/docker-compose.yaml",
Qualifier = nil, Path = "docker-compose.yaml",
Parent = nil.

#### Deeply nested external path

Input: "EXTERNAL/a/b/c/d/schema.proto".
Expect: Type = NodeTypeExternal,
Name = "EXTERNAL/a/b/c/d/schema.proto",
Qualifier = nil, Path = "a/b/c/d/schema.proto",
Parent = nil.

### LogicalNameParse — ARTIFACT type

These tests use the `testChdir` pattern. Before each
test, create a temp dir, chdir to it, and create the
generator's `_node.md` with frontmatter.

#### Simple artifact

Setup: create file
`code-from-spec/extraction/proto/_node.md` with:
```
---
output: internal/extraction/proto.go
---
# SPEC/extraction/proto
```

Input: "ARTIFACT/extraction/proto".
Expect: Type = NodeTypeArtifact,
Name = "ARTIFACT/extraction/proto",
Qualifier = nil,
Path = "internal/extraction/proto.go",
Parent = pointer to "SPEC/extraction/proto".

#### Artifact with nested generator

Setup: create file
`code-from-spec/payments/fees/calculation/_node.md` with:
```
---
output: internal/fees/calculation.go
---
# SPEC/payments/fees/calculation
```

Input: "ARTIFACT/payments/fees/calculation".
Expect: Type = NodeTypeArtifact,
Name = "ARTIFACT/payments/fees/calculation",
Qualifier = nil,
Path = "internal/fees/calculation.go",
Parent = pointer to "SPEC/payments/fees/calculation".

#### Artifact generator has no output

Setup: create file
`code-from-spec/docs/overview/_node.md` with:
```
---
---
# SPEC/docs/overview
```

Input: "ARTIFACT/docs/overview".
Expect error: ErrNoOutput.

#### Artifact generator does not exist on disk

Input: "ARTIFACT/nonexistent/node".
Expect error: propagated from FrontmatterParse
(generator's _node.md file is missing).

### LogicalNameParse — errors

#### Unrecognized prefix

Input: "UNKNOWN/something".
Expect error: ErrUnrecognizedPrefix.

#### Empty string

Input: "".
Expect error: ErrUnrecognizedPrefix.

#### ROOT prefix

Input: "ROOT/x".
Expect error: ErrUnrecognizedPrefix.

#### SPEC/ with empty relative path

Input: "SPEC/".
Expect error: ErrInvalidName.

#### ARTIFACT/ with empty relative path

Input: "ARTIFACT/".
Expect error: ErrInvalidName.

#### EXTERNAL/ with empty relative path

Input: "EXTERNAL/".
Expect error: ErrInvalidName.

### LogicalNameFromPath

#### Root node

Input: PathCfs "code-from-spec/_node.md".
Expect: Type = NodeTypeSpec, Name = "SPEC",
Qualifier = nil,
Path = "code-from-spec/_node.md",
Parent = nil.

#### Nested node

Input: PathCfs "code-from-spec/x/y/_node.md".
Expect: Type = NodeTypeSpec, Name = "SPEC/x/y",
Qualifier = nil,
Path = "code-from-spec/x/y/_node.md",
Parent = pointer to "SPEC/x".

#### Deeply nested node

Input: PathCfs "code-from-spec/a/b/c/d/_node.md".
Expect: Type = NodeTypeSpec, Name = "SPEC/a/b/c/d",
Qualifier = nil,
Path = "code-from-spec/a/b/c/d/_node.md",
Parent = pointer to "SPEC/a/b/c".

#### Rejects non-spec path

Input: PathCfs "internal/config/config.go".
Expect error: ErrInvalidPath.

#### Rejects path without _node.md

Input: PathCfs "code-from-spec/x/y/output.md".
Expect error: ErrInvalidPath.

#### Rejects path not starting with code-from-spec/

Input: PathCfs "other/x/_node.md".
Expect error: ErrInvalidPath.

## Go-specific guidance

- The package name is `logicalnames_test` (external
  test package).
- SPEC and EXTERNAL tests are pure — no filesystem
  needed.
- ARTIFACT tests and LogicalNameFromPath tests do not
  need filesystem either — except for the ARTIFACT
  tests that call `LogicalNameParse` (which reads
  frontmatter). Those use the `testChdir` pattern.
- Use table-driven tests where appropriate.
- Compare pointer fields: for nil, check `== nil`;
  for non-nil, dereference and compare the string value.
