# SPEC/golang/test/cases

Test case specifications.

# Public

## Test rules

- Use an external test package for black-box testing.
  Import the package under test explicitly. This
  ensures tests exercise the public API only.
- Tests that perform file I/O use `testutils.Chdir(t)`
  for isolation. Pure function tests (no file I/O)
  do not need it.
- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.

## Imports

Import the `testutils` package for shared helpers:
- `testutils.Chdir(t)` — create temp dir and set
  working directory.
- `testutils.CreateSpecNode(t, logicalName)` — builder
  for valid `_node.md` files.
- `testutils.WriteRawNode(t, logicalName, content)` —
  write arbitrary `_node.md` content (for error tests).
- `testutils.Ptr(value)` — create a pointer to any
  value (for `*string` fields like `Output`, `Input`,
  `Qualifier`, `ParentName`).

Do not define local helpers for these operations.

## Temporary files and CfsPath

Tests that create files and pass them as `CfsPath`
values must change the working directory to a temp dir
so that `GetProjectRoot` and `CfsPathToOs` resolve
paths correctly. Without this, `t.TempDir()` creates
directories in the OS temp location, which may be on a
different drive (Windows) or outside the project root —
causing path resolution to fail.

Use `testutils.Chdir(t)`:

1. Call `dir := testutils.Chdir(t)`.
2. Create spec nodes with `testutils.CreateSpecNode`
   or files with `os.WriteFile`.
3. Pass relative paths as `CfsPath` values.

## CfsPath values in tests

`CfsPath` values must always use forward slashes (`/`),
even on Windows. Never use `filepath.Separator` or
backslashes in `CfsPath` values. For example, to test a
nonexistent file: `CfsPath("nonexistent/file.txt")`,
not `"nonexistent\file.txt"`.

## Error propagation across packages

When a function propagates an error from another package
(e.g. `parsing.ParseNode` propagating from `OpenFile`),
the error chain preserves the original sentinel. Use
`errors.Is` with the sentinel from the **originating**
package (e.g. `oslayer.ErrFileUnreadable`), not a
re-declared sentinel in the calling package — unless
the calling package's interface explicitly declares its
own sentinel and wraps it.

## Constructing records from other packages

When tests construct records manually (e.g.
`parsing.Node`, `Chain`), the field values
must be consistent with what the real producers would
generate:

- `NodeSection.Heading` is the **normalized** form
  (lowercase, whitespace collapsed) as produced by
  `parsing.ParseNode`. Example: `"spec/a"` for a node
  at `SPEC/a`.
- `NodeSection.RawHeading` is the original line as read
  from the file. Example: `"# SPEC/a"`.
- `NodeSection.Content` is a `[]string` (list of lines).
- `Chain` fields hold `parsing.CfsReference` values.
  `CfsReference.LogicalName` must be a valid logical
  name (`SPEC/` for spec nodes, `ARTIFACT/` for
  artifacts). For spec nodes, the logical name must
  resolve to a `_node.md` file that exists on disk.
  Tests must create the spec tree files accordingly.
- Use `testutils.Ptr("value")` for `*string` fields
  (`Output`, `Input`, `Qualifier`, `ParentName`).

## Creating _node.md files in tests

Use `testutils.CreateSpecNode` for valid nodes — it
handles path derivation, directory creation, frontmatter
format, and the node name heading:

```go
b := testutils.CreateSpecNode(t, "SPEC/a/b")
b.SetOutput("internal/a/b.go")
b.AddDependsOn("SPEC/other")
b.SetPublic("## Interface\ncontent")
b.Write()
```

Use `testutils.WriteRawNode` for malformed content
(testing parse error cases):

```go
testutils.WriteRawNode(t, "SPEC/a", "not valid markdown")
```

Rules for `_node.md` content (enforced by
`CreateSpecNode`, relevant for `WriteRawNode`):

- The first heading in the file body must be
  `# <logical-name>` (e.g. `# SPEC/root` for a root
  node, `# SPEC/root/a` for a child).
  `parsing.ParseNode` validates that the first heading
  matches the logical name.
- Bare `SPEC` (without a trailing slash) is not a valid
  logical name. Root nodes are direct children of
  `code-from-spec/` (e.g. `SPEC/root` at
  `code-from-spec/root/_node.md`).
- Frontmatter is optional. Only include frontmatter
  fields that the node actually uses.

## Hash format

- Hashes are exactly 27 characters of base64url
  (RFC 4648 §5, no padding). Characters allowed:
  `A-Z`, `a-z`, `0-9`, `-`, `_`.
- When the test spec specifies exact hash values or
  other string literals, use them verbatim — do not
  invent substitutes.
- When a test needs a placeholder stale hash (one that
  will not match the computed chain hash), use a
  27-character string like `AAAAAAAAAAAAAAAAAAAAAAAAAAA`
  — easy to verify the length is correct.
