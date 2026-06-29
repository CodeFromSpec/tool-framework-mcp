---
output: internal/parsing/logical_names.go
---

# SPEC/golang/implementation/parsing/logical_names

Parses CFS references into a structured representation
with all resolved information: node type, logical name,
qualifier, file path, and parent name.

# Agent

Implement the types and functions listed in the
Ownership section as a Go file in package `parsing`.

## Ownership

This file declares and implements:
- Types: `CfsNodeType`, `CfsReference`
- Constants: `CfsNodeTypeSpec`, `CfsNodeTypeArtifact`,
  `CfsNodeTypeExternal`
- Functions: `CfsReferenceFromName`,
  `CfsReferenceFromPath`

The following exist in other files of this package and
can be used but must not be redeclared:
- Types: `NodeFrontmatter` — declared in
  `node_parsing.go`.
- Functions: `ParseNode` — declared in
  `node_parsing.go`. Called for ARTIFACT/ references
  to resolve the output path.
- Error sentinels — declared in `errors.go`.

All unexported helpers must use the suffix `LN`
(e.g. `extractQualifierLN`, `stringPtrLN`). This is
mandatory to avoid name collisions with other files
in the package.

## Logic

### CfsReferenceFromName(logicalName: string) -> *CfsReference

1. **Extract qualifier.** Find the first `(` in
   logicalName. If found, find the matching `)` after
   it. Extract the text between them as the qualifier
   string. Let `stripped` be the portion before `(`.
   If no `(`, qualifier is nil and `stripped` is
   logicalName unchanged.

2. **Classify by prefix.** Examine `stripped`:

   a. If `stripped` starts with `"SPEC/"`:
      Let `relative` = stripped with "SPEC/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Let path = "code-from-spec/" + relative +
      "/_node.md".
      Compute parent: find the last "/" in relative.
      If no "/" found, parent is nil (this is a root
      node — a direct child of code-from-spec/).
      If "/" found, parent = "SPEC/" + substring
      before the last "/".
      Return CfsReference with
      NodeType = CfsNodeTypeSpec,
      LogicalName = stripped, Qualifier = qualifier,
      Path = path, ParentName = parent (nil or
      pointer).

   c. If `stripped` starts with `"ARTIFACT/"`:
      Let `relative` = stripped with "ARTIFACT/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Let generatorName = "SPEC/" + relative.
      Call ParseNode(generatorName).
      If it fails, propagate the error.
      If node.Frontmatter is nil or
      node.Frontmatter.Output is nil, raise
      ErrNoOutput.
      Return CfsReference with
      NodeType = CfsNodeTypeArtifact,
      LogicalName = stripped, Qualifier = nil,
      Path = *node.Frontmatter.Output,
      ParentName = pointer to generatorName.

   d. If `stripped` starts with `"EXTERNAL/"`:
      Let `relative` = stripped with "EXTERNAL/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Return CfsReference with
      NodeType = CfsNodeTypeExternal,
      LogicalName = stripped, Qualifier = nil,
      Path = relative, ParentName = nil.

   e. Otherwise: raise ErrUnrecognizedPrefix.

### CfsReferenceFromPath(cfsPath: oslayer.CfsPath) -> *CfsReference

1. Let `value` = string(cfsPath).

2. If `value` does not start with "code-from-spec/",
   raise ErrInvalidPath.

3. If `value` does not end with "/_node.md",
   raise ErrInvalidPath.

4. Remove "code-from-spec/" prefix and "/_node.md"
   suffix. Let `relative` = the remainder.
   If `relative` is empty, raise ErrInvalidPath.

5. Let logicalName = "SPEC/" + relative.
   Compute parent: find the last "/" in relative.
   If no "/" found, parent is nil (root node).
   If "/" found, parent = "SPEC/" + substring before
   the last "/".

6. Return CfsReference with
   NodeType = CfsNodeTypeSpec,
   LogicalName = logicalName, Qualifier = nil,
   Path = value, ParentName = parent (nil or pointer).

## Go-specific guidance

- Use the `oslayer` package for `CfsPath`.
- Use `ParseNode` from this package for ARTIFACT/
  path resolution.
- Wrap propagated errors with `fmt.Errorf` + `%w`.
- For the qualifier pointer, use a helper like
  `func stringPtrLN(s string) *string { return &s }`.
- Error sentinels are declared in `errors.go` — do
  not redeclare them here.
- The package name should be `parsing`.

# Private

## Decisions

### Struct-based interface replacing 12 functions

The original interface had 12 individual functions
(LogicalNameToPath, LogicalNameIsSpec,
LogicalNameGetParent, etc.). Replaced with a single
`CfsReferenceFromName` returning a `CfsReference`
struct. Eliminates repeated parsing of the same string
and simplifies consumer code (one call instead of 3-4).
`CfsReferenceFromPath` kept as a separate constructor
for reverse resolution.

### ARTIFACT Path resolves via ParseNode I/O

`CfsReferenceFromName("ARTIFACT/x")` reads the
generator node via `ParseNode` to populate `Path` with
the actual output file path. This means the call does
I/O for ARTIFACT types (not for SPEC or EXTERNAL). The
trade-off is that consumers no longer need to manually
resolve artifact paths — `ref.Path` is always usable.
Consequence: callers that only need type classification
(node_ranking, validate) should use string prefix
checks for ARTIFACT/EXTERNAL instead of
CfsReferenceFromName to avoid unnecessary I/O.

### Bare SPEC no longer valid (v5 multi-root)

In v4, `SPEC` (without slash) referred to the single
root node at `code-from-spec/_node.md`. In v5,
`code-from-spec/` is not a node — root nodes are
direct children (e.g. `SPEC/golang`). Bare `SPEC`
now returns `ErrUnrecognizedPrefix`.
