---
depends_on:
  - ROOT/tech_design/internal/frontmatter
  - ROOT/tech_design/internal/logical_names
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: chainresolver
    path: internal/chainresolver/chainresolver.go
---

# ROOT/tech_design/internal/chain_resolver/implementation

Generates the chainresolver package implementation.

# Agent

## Implementation

**Step 1 — Ancestors and Target**

Starting from the target logical name, repeatedly call
`logicalnames.ParentLogicalName` to walk upward, collecting each
logical name. Sort the list by logical name alphabetically.

For each logical name, call `logicalnames.PathFromLogicalName` to
resolve the file path and create a `ChainItem` with
`Qualifier` = nil.

The last item in the sorted list is the `Target`; the
remaining items form `Ancestors`.

**Step 2 — Dependencies**

Read the target node's frontmatter using `ParseFrontmatter`.
Collect all `DependsOn` entries and process them.

For each entry in `DependsOn`:
1. Call `logicalnames.PathFromLogicalName` to get the file path.
2. Determine the qualifier: call `logicalnames.HasQualifier` and
   `logicalnames.QualifierName` on the logical name. If the logical name
   has a qualifier, set `Qualifier` to that value. Otherwise,
   set `Qualifier` to nil.
3. Verify the file exists on disk (using `os.Stat`). If it
   does not exist, return error:
   `"cannot resolve logical name: <name>"`.
4. Add a `ChainItem` with the file path and qualifier to
   `Dependencies`.

Sort `Dependencies` alphabetically by `FilePath`, then by
`Qualifier` (nil sorts before non-nil).

**Step 3 — Code**

Read the target node's frontmatter using `ParseFrontmatter`
and extract the `Implements` list. For each path in
`Implements`, check if the file exists on disk (using
`os.Stat`). If it exists, add the path to `Code`. If it does
not exist, skip it. `Code` contains only files that already
exist.

**Step 4 — Normalize file paths**

Convert all file paths in `Ancestors`, `Target`,
`Dependencies`, and `Code` to use forward slashes as
separators, regardless of the operating system. Use
`filepath.ToSlash`.

**Step 5 — Deduplicate**

Review `Ancestors` and `Dependencies` and remove duplicate
entries. Two entries are considered duplicates when they have
the same `FilePath` and the same `Qualifier`.

Additionally, when an entry exists with a given `FilePath`
and `Qualifier` = nil (meaning the entire `# Public` section),
any other entry with the same `FilePath` and a non-nil
`Qualifier` is redundant and must be removed — the full
`# Public` already includes every subsection.

When removing duplicates, keep the first occurrence.
