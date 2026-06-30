---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/test/utils/helpers
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/parsinglogicalnamestest/parsing_logicalnames_test.go
---

# SPEC/golang/test/cases/parsing/logical_names

Unit tests for `parsing.CfsReferenceFromName` and
`parsing.CfsReferenceFromPath`.

# Agent

## Context

Tests for `parsing.CfsReferenceFromName` and
`parsing.CfsReferenceFromPath`.

SPEC and EXTERNAL tests are pure string parsing — no
filesystem needed. ARTIFACT tests require a temp
directory with a `_node.md` file containing frontmatter
with an `output` field, because `CfsReferenceFromName`
reads the generator's frontmatter via `ParseNode`.

ARTIFACT tests use the `testutils.Chdir` pattern: create a
temp dir, chdir to it, create the necessary
`code-from-spec/.../_node.md` files, then call
`CfsReferenceFromName`.

## Test cases

### CfsReferenceFromName — SPEC type

#### Bare SPEC is invalid

Input: "SPEC".
Expect error: ErrUnrecognizedPrefix.

#### Bare ARTIFACT is invalid

Input: "ARTIFACT".
Expect error: ErrUnrecognizedPrefix.

#### Bare EXTERNAL is invalid

Input: "EXTERNAL".
Expect error: ErrUnrecognizedPrefix.

#### SPEC root node (single segment)

Input: "SPEC/domain".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/domain",
Qualifier = nil,
Path = "code-from-spec/domain/_node.md",
ParentName = nil.

#### SPEC with nested path

Input: "SPEC/payments/fees/calculation".
Expect: NodeType = parsing.CfsNodeTypeSpec,
LogicalName = "SPEC/payments/fees/calculation",
Qualifier = nil,
Path = "code-from-spec/payments/fees/calculation/_node.md",
ParentName = pointer to "SPEC/payments/fees".

#### SPEC with qualifier

Input: "SPEC/x/y(interface)".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/x/y",
Qualifier = pointer to "interface",
Path = "code-from-spec/x/y/_node.md",
ParentName = pointer to "SPEC/x".

#### SPEC with qualifier — root level

Input: "SPEC/domain(context)".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/domain",
Qualifier = pointer to "context",
Path = "code-from-spec/domain/_node.md",
ParentName = nil.

#### SPEC with qualifier — parent is computed from unqualified name

Input: "SPEC/domain/config(interface)".
Expect: NodeType = parsing.CfsNodeTypeSpec,
LogicalName = "SPEC/domain/config",
Qualifier = pointer to "interface",
Path = "code-from-spec/domain/config/_node.md",
ParentName = pointer to "SPEC/domain".

### CfsReferenceFromName — EXTERNAL type

#### Simple external path

Input: "EXTERNAL/proto/v1/api.proto".
Expect: NodeType = parsing.CfsNodeTypeExternal,
LogicalName = "EXTERNAL/proto/v1/api.proto",
Qualifier = nil, Path = "proto/v1/api.proto",
ParentName = nil.

#### Root-level external file

Input: "EXTERNAL/docker-compose.yaml".
Expect: NodeType = parsing.CfsNodeTypeExternal,
LogicalName = "EXTERNAL/docker-compose.yaml",
Qualifier = nil, Path = "docker-compose.yaml",
ParentName = nil.

#### Deeply nested external path

Input: "EXTERNAL/a/b/c/d/schema.proto".
Expect: NodeType = parsing.CfsNodeTypeExternal,
LogicalName = "EXTERNAL/a/b/c/d/schema.proto",
Qualifier = nil, Path = "a/b/c/d/schema.proto",
ParentName = nil.

### CfsReferenceFromName — ARTIFACT type

These tests use the `testutils.Chdir` pattern. Before each
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
Expect: NodeType = parsing.CfsNodeTypeArtifact,
LogicalName = "ARTIFACT/extraction/proto",
Qualifier = nil,
Path = "internal/extraction/proto.go",
ParentName = pointer to "SPEC/extraction/proto".

#### Artifact with nested generator

Setup: create file
`code-from-spec/payments/fees/calculation/_node.md`
with:
```
---
output: internal/fees/calculation.go
---
# SPEC/payments/fees/calculation
```

Input: "ARTIFACT/payments/fees/calculation".
Expect: NodeType = parsing.CfsNodeTypeArtifact,
LogicalName = "ARTIFACT/payments/fees/calculation",
Qualifier = nil,
Path = "internal/fees/calculation.go",
ParentName = pointer to
"SPEC/payments/fees/calculation".

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
Expect error: propagated from ParseNode
(generator's _node.md file is missing).

### CfsReferenceFromName — errors

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

#### SPEC name with trailing slash

Input: "SPEC/a/b/".
Expect error: ErrInvalidName.

#### ARTIFACT/ with empty relative path

Input: "ARTIFACT/".
Expect error: ErrInvalidName.

#### EXTERNAL/ with empty relative path

Input: "EXTERNAL/".
Expect error: ErrInvalidName.

### CfsReferenceFromPath

#### Root node (direct child of code-from-spec/)

Input: oslayer.CfsPath "code-from-spec/domain/_node.md".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/domain",
Qualifier = nil,
Path = "code-from-spec/domain/_node.md",
ParentName = nil.

#### Nested node

Input: oslayer.CfsPath "code-from-spec/x/y/_node.md".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/x/y",
Qualifier = nil,
Path = "code-from-spec/x/y/_node.md",
ParentName = pointer to "SPEC/x".

#### Deeply nested node

Input: oslayer.CfsPath "code-from-spec/a/b/c/d/_node.md".
Expect: NodeType = parsing.CfsNodeTypeSpec, LogicalName = "SPEC/a/b/c/d",
Qualifier = nil,
Path = "code-from-spec/a/b/c/d/_node.md",
ParentName = pointer to "SPEC/a/b/c".

#### Rejects bare code-from-spec/_node.md

Input: oslayer.CfsPath "code-from-spec/_node.md".
Expect error: ErrInvalidPath.

#### Rejects non-spec path

Input: oslayer.CfsPath "internal/config/config.go".
Expect error: ErrInvalidPath.

#### Rejects path without _node.md

Input: oslayer.CfsPath "code-from-spec/x/y/output.md".
Expect error: ErrInvalidPath.

#### Rejects path not starting with code-from-spec/

Input: oslayer.CfsPath "other/x/_node.md".
Expect error: ErrInvalidPath.

## Go-specific guidance

- The package name is `parsinglogicalnamestest` (external test
  package).
- SPEC and EXTERNAL tests are pure — no filesystem
  needed.
- ARTIFACT tests and CfsReferenceFromPath tests do not
  need filesystem either — except for the ARTIFACT
  tests that call `parsing.CfsReferenceFromName` (which
  reads frontmatter). Those use the `testutils.Chdir` pattern.
- Use table-driven tests where appropriate.
- Compare pointer fields: for nil, check `== nil`;
  for non-nil, dereference and compare the string value.
