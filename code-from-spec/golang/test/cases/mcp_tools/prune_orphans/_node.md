---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/prune_orphans
  - SPEC/golang/implementation/oslayer(interface)
output: internal/mcppruneorphans/mcppruneorphans_test.go
---

# SPEC/golang/test/cases/mcp_tools/prune_orphans

# Agent

## Test setup guidance

`MCPPruneOrphans` scans the spec tree, reads the
manifest, deletes orphan artifact files, and removes
orphan entries from the manifest. Tests must create
spec nodes, manifest entries, and artifact files to
set up the conditions.

Use `testutils.Chdir` for isolation. Create spec nodes
with `testutils.CreateSpecNode`. Create manifest
entries with `manifest.OpenManifest(false)` + `m.Save()`.
Create artifact files with `oslayer.OpenFile` in
"overwrite" mode + `handle.Close()`.

## Test cases

### Happy path

#### Removes orphan entry and deletes artifact file

Setup:
- Create a spec node at "alpha" with
  `output: "out/alpha.go"` in its frontmatter and
  some public content.
- Create a manifest with two entries:
  - `ARTIFACT/alpha` with path `out/alpha.go`.
  - `ARTIFACT/removed` with path `out/removed.go`.
- Create the artifact file at `out/removed.go` on disk.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- Summary contains `"pruned ARTIFACT/removed (out/removed.go) — deleted from disk"`.
- Summary contains `"pruned orphans: 1 entries removed"`.
- Re-open manifest read-only: `ARTIFACT/alpha` entry
  still exists, `ARTIFACT/removed` does not.
- `oslayer.OpenFile("out/removed.go", "read", 0)`
  returns `oslayer.ErrFileUnreadable` (file deleted).

#### Removes orphan when artifact file does not exist on disk

Setup:
- Create a spec node at "alpha" with
  `output: "out/alpha.go"`.
- Create a manifest with two entries:
  - `ARTIFACT/alpha` with path `out/alpha.go`.
  - `ARTIFACT/gone` with path `out/gone.go`.
- Do NOT create `out/gone.go` on disk.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- Summary contains `"pruned ARTIFACT/gone (out/gone.go) — not found on disk"`.
- Summary contains `"pruned orphans: 1 entries removed"`.
- Re-open manifest read-only: `ARTIFACT/gone` does not
  exist.

#### No orphans — zero entries removed

Setup:
- Create a spec node at "alpha" with
  `output: "out/alpha.go"`.
- Create a manifest with one entry:
  `ARTIFACT/alpha` with path `out/alpha.go`.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- Summary contains `"pruned orphans: 0 entries removed"`.
- Manifest unchanged.

#### Orphan because node has no output

Setup:
- Create a spec node at "alpha" with
  `output: "out/alpha.go"`.
- Create a spec node at "docs-only" with no output
  in its frontmatter.
- Create a manifest with two entries:
  - `ARTIFACT/alpha` with path `out/alpha.go`.
  - `ARTIFACT/docs-only` with path `out/docs.go`.
- Create the artifact file at `out/docs.go` on disk.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- Summary contains `"pruned ARTIFACT/docs-only (out/docs.go) — deleted from disk"`.
- Summary contains `"pruned orphans: 1 entries removed"`.
- File `out/docs.go` deleted from disk.

### Edge cases

#### Empty manifest — no errors

Setup:
- Create a spec node at "alpha" with
  `output: "out/alpha.go"`.
- Create a valid manifest with header only (no
  entries). Create the `.manifest.lock` file.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- Summary contains `"pruned orphans: 0 entries removed"`.

#### Multiple orphans pruned in alphabetical order

Setup:
- Create no spec nodes (empty spec tree — create
  the `code-from-spec/` directory only).
- Create a manifest with entries:
  - `ARTIFACT/zebra` with path `out/z.go`.
  - `ARTIFACT/apple` with path `out/a.go`.
- Create both artifact files on disk.

Actions:
1. Call `mcppruneorphans.MCPPruneOrphans()`.

Expected:
- No error.
- The first pruned line is for `ARTIFACT/apple`.
- The second pruned line is for `ARTIFACT/zebra`.
- Summary contains `"pruned orphans: 2 entries removed"`.

## Go-specific guidance

- The package name is `mcppruneorphans_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- Use `testutils.CreateSpecNode(t, ...)` to create spec
  nodes with frontmatter.
- Use `manifest.OpenManifest(false)` + `m.Save()` to
  create manifest fixtures.
- Use `oslayer.OpenFile(path, "overwrite", 0)` +
  `handle.Close()` to create artifact files on disk.
- Use `oslayer.OpenFile(path, "read", 0)` to verify
  file existence/absence after pruning.
- Use `errors.Is` for error sentinel checks.
- Use `strings.Contains` for checking summary output.
