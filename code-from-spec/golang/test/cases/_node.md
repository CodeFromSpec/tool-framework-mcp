# SPEC/golang/test/cases

Test case specifications.

# Public

## Test rules

- Use an external test package for black-box testing.
  Import the package under test explicitly. This
  ensures tests exercise the public API only.
- Each test uses `t.TempDir()` for isolation.
- Create test files with controlled content using
  `os.WriteFile`.

## Temporary files and CfsPath

Tests that create files and pass them as `CfsPath` values
must change the working directory to a temp dir so that
`GetProjectRoot` and `CfsPathToOs` resolve paths
correctly. Without this, `t.TempDir()` creates
directories in the OS temp location, which may be on a
different drive (Windows) or outside the project root —
causing path resolution to fail.

Use `testutils.Chdir(t)` from the `testutils` package
(`internal/testutils`). It creates a temp dir, changes
the working directory to it, and restores on cleanup:

1. Call `dir := testutils.Chdir(t)`.
2. Create files using paths relative to the temp dir.
3. Pass those relative paths as `CfsPath` values.

Do not define a local `testChdir` or `Chdir` helper —
use `testutils.Chdir` instead.

Tests that do not create files (pure function tests)
do not need this pattern.

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
  `parsing.ParseNode`. Example: `"spec/a"` for a node at
  `SPEC/a`.
- `NodeSection.RawHeading` is the original line as read
  from the file. Example: `"# SPEC/a"`.
- `NodeSection.Content` is a `[]string` (list of lines).
- `Chain` fields hold `parsing.CfsReference` values.
  `CfsReference.LogicalName` must be a valid logical
  name (`SPEC/` for spec nodes, `ARTIFACT/` for
  artifacts). For spec nodes, the logical name must
  resolve to a `_node.md` file that exists on disk.
  Tests must create the spec tree files accordingly.

## Creating _node.md files in tests

When tests create `_node.md` files on disk:

- The first heading in the file body must be
  `# <logical-name>` (e.g. `# SPEC/root` for a root
  node, `# SPEC/root/a` for a child). `parsing.ParseNode`
  validates that the first heading matches the logical
  name — tests that omit it or use a different heading
  (e.g. `# Public`) will fail with
  `ErrNodeNameDoesNotMatch`.
- Bare `SPEC` (without a trailing slash) is not a valid
  logical name. Root nodes are direct children of
  `code-from-spec/` (e.g. `SPEC/root` at
  `code-from-spec/root/_node.md`). There is no
  `code-from-spec/_node.md` root node.
- Frontmatter is optional. Only include frontmatter
  fields that the node actually uses. An intermediate
  node (one with children) has no `output`, `depends_on`,
  or `input` — do not write frontmatter for intermediate
  nodes. Only leaf nodes that declare these fields need
  frontmatter.

## Hash format

- Hashes are exactly 27 characters of base64url
  (RFC 4648 §5, no padding). Characters allowed:
  `A-Z`, `a-z`, `0-9`, `-`, `_`.
- When the functional test input specifies exact hash
  values or other string literals, use them verbatim —
  do not invent substitutes.
- When a test needs a placeholder stale hash (one that
  will not match the computed chain hash), use a
  27-character string like `AAAAAAAAAAAAAAAAAAAAAAAAAAA`
  — easy to verify the length is correct.

## Error and style conventions

- Use `errors.Is` to check error sentinels.
- Use table-driven tests where appropriate.
