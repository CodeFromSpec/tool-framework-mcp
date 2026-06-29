# SPEC/golang/tests

Go test files generated from functional test specifications
and implementation artifacts.

# Public

## Test rules

- Translate functional test cases from the `input`
  artifact into Go test functions.
- Use the interface from `depends_on` for types, function
  signatures, and import paths.
- Use the external test package (`package <name>_test`)
  for black-box testing. Import the package under test
  explicitly. This avoids name collisions and ensures
  tests exercise the public API only.
- Each test uses `t.TempDir()` for isolation.
- Create test files with controlled content using
  `os.WriteFile`.

## Temporary files and PathCfs

Tests that create files and pass them as `PathCfs` values
must use the `testChdir` pattern:

1. Create a temp dir with `t.TempDir()`.
2. Call `os.Chdir(tempDir)` to make it the working
   directory. Register `t.Cleanup` to restore the
   original directory.
3. Create files using paths relative to the temp dir
   (e.g., `os.WriteFile("mydir/file.txt", ...)`).
4. Pass those relative paths as `PathCfs.Value`.

This works because `PathGetProjectRoot` returns the
working directory, and `PathCfsToOs` resolves relative
paths against it. Without `testChdir`, `t.TempDir()`
creates directories in the OS temp location, which may
be on a different drive (Windows) or outside the project
root — causing path resolution to fail.

A typical `testChdir` helper:

```go
func testChdir(t *testing.T, dir string) {
    t.Helper()
    orig, err := os.Getwd()
    if err != nil {
        t.Fatalf("testChdir: %v", err)
    }
    if err := os.Chdir(dir); err != nil {
        t.Fatalf("testChdir: %v", err)
    }
    t.Cleanup(func() {
        if err := os.Chdir(orig); err != nil {
            t.Errorf("testChdir cleanup: %v", err)
        }
    })
}
```

Tests that do not create files (pure function tests)
do not need this pattern.

## PathCfs values in tests

`PathCfs` values must always use forward slashes (`/`),
even on Windows. Never use `filepath.Separator` or
backslashes in `PathCfs.Value`. For example, to test a
nonexistent file: `PathCfs{Value: "nonexistent/file.txt"}`,
not `"nonexistent\file.txt"`.

## Error propagation across packages

When a function propagates an error from another package
(e.g. `FrontmatterParse` propagating from `FileOpen`),
the error chain preserves the original sentinel. Use
`errors.Is` with the sentinel from the **originating**
package (e.g. `file.ErrFileUnreadable`), not a
re-declared sentinel in the calling package — unless
the calling package's interface explicitly declares its
own sentinel and wraps it.

## Constructing records from other packages

When tests construct records manually (e.g.
`SpecTreeValidateInput`, `Chain`, `ChainItem`), the
field values must be consistent with what the real
producers would generate:

- `NodeSection.Heading` is the **normalized** form
  (lowercase, whitespace collapsed) as produced by
  `NodeParse`. Example: `"spec/a"` for a node at
  `SPEC/a`.
- `NodeSection.RawHeading` is the original line as read
  from the file. Example: `"# SPEC/a"`.
- `NodeSection.Content` is a `[]string` (list of lines).
- `ChainItem.LogicalName` must be a valid logical name
  (`SPEC/` for spec nodes, `ARTIFACT/` for artifacts).
  For spec nodes, the logical name must resolve to a
  `_node.md` file that exists on disk (via
  `LogicalNameToPath`). Tests must create the spec tree
  files accordingly.
- `ChainItem.FilePath` is a `PathCfs` with forward
  slashes.

## Creating _node.md files in tests

When tests create `_node.md` files on disk:

- The first heading in the file body must be
  `# <logical-name>` (e.g. `# SPEC/root` for a root
  node, `# SPEC/root/a` for a child). `NodeParse`
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
