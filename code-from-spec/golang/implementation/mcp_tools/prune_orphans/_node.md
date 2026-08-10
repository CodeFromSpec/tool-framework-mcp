---
depends_on:
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcppruneorphans/mcppruneorphans.go
---

# SPEC/golang/implementation/mcp_tools/prune_orphans

Removes orphan entries from the manifest — entries
whose corresponding spec node no longer exists or no
longer declares a type.

# Public

## Package

`package mcppruneorphans`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcppruneorphans"`

## Interface

```go
func MCPPruneOrphans() (string, error)
```

### Output

A summary message listing each pruned entry and a
total count:

```
pruned ARTIFACT/foo/bar (path/to/file.go) — deleted from disk
pruned ARTIFACT/baz/qux (path/to/other.go) — not found on disk
pruned orphans: 2 entries removed
```

When no orphans are found:

```
pruned orphans: 0 entries removed
```

### Errors

- Propagated errors from `manifest`, `spectree`,
  `parsing`, `oslayer` packages.

# Agent

Implement the prune orphans tool as a Go package.

## Logic

### Step 1 — Discover and parse nodes

1. Call `spectree.SpecTreeScan()` to discover all spec
   nodes. If it returns `spectree.ErrNoNodesFound`,
   treat as an empty node list (no nodes exist, so
   every manifest entry is an orphan). If it fails
   with any other error, propagate it.

2. For each discovered CfsReference, call
   `parsing.ParseNode(ref.LogicalName)`. If parsing
   fails, skip the node. Collect successfully parsed
   nodes into a map keyed by logical name.

### Step 2 — Open manifest for writing

3. Call `manifest.OpenManifest(false)`. If it fails,
   propagate the error. Store as `m`.
   Immediately defer `m.Discard()`.

### Step 3 — Identify orphans

4. For each entry in `m.Entries`:
   Derive the generating node's logical name: strip
   "ARTIFACT/" prefix and prepend "SPEC/".
   An entry is orphan if:
   - No parsed node has that logical name, OR
   - The node's frontmatter Type is nil.

### Step 4 — Delete artifact files from disk

5. For each orphan entry, in alphabetical order by
   key:
   a. Construct `oslayer.CfsPath` from entry.Path.
   b. Attempt to open the file with
      `oslayer.OpenFile(path, "read", 0)`.
      - If it returns `oslayer.ErrFileUnreadable`: the
        file does not exist. Append a line:
        `"pruned <key> (<entry.Path>) — not found on disk"`.
        Mark entry as handled. Continue to next.
      - If it returns `oslayer.ErrLockTimeout` or any
        other error: skip this entry (do not mark as
        handled). Continue to next.
      - If it succeeds: close the handle immediately.
   c. Call `oslayer.DeleteFile(path)`.
   d. If DeleteFile succeeds, append a line:
      `"pruned <key> (<entry.Path>) — deleted from disk"`.
      Mark entry as handled.
   e. If DeleteFile fails: skip this entry (do not
      mark as handled). Continue to next.

### Step 5 — Remove entries from manifest

6. For each orphan marked as handled (deleted or not
   found on disk), delete the entry from `m.Entries`.
   Entries not marked as handled remain in the
   manifest.

### Step 6 — Save and return

7. Call `m.Save()`. If it fails, propagate the error.

8. Append the summary line:
   `"pruned orphans: N entries removed"` where N is
   the count of entries removed.

9. Return the full message (all lines joined with
   newline) and nil error.

## Go-specific guidance

- The package name is `mcppruneorphans`.
- Use the `spectree` package for `SpecTreeScan`.
- Use the `parsing` package for `ParseNode`,
  `CfsReference`, `Node`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use the `oslayer` package for `DeleteFile`,
  `ValidateStringIsCfsPath`, and `CfsPath`.
- Use `sort.Strings` for ordering orphan keys.
- Use `strings.Builder` for assembling the output
  message.
- Use `strings.TrimPrefix` and `"SPEC/" + ...` for
  logical name derivation.
