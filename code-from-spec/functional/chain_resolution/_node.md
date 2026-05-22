---
outputs:
  - id: chain_resolution
    path: code-from-spec/functional/chain_resolution/output.md
---

# ROOT/functional/chain_resolution

Resolves the ordered list of files that form the chain for a
given target logical name.

# Public

## Behavior

Given a target logical name, returns the chain separated into
ancestors, target, dependencies, and existing code files.

### Input

A logical name (e.g., `ROOT/golang/server/code`).

### Output

| Field | Description |
|---|---|
| `ancestors` | Ancestor nodes from root to parent, sorted alphabetically. |
| `target` | The target node itself. |
| `dependencies` | Nodes referenced by `depends_on`, sorted by file path. |
| `code` | Existing files declared in `outputs` that already exist on disk. |

Each item in ancestors, target, and dependencies has:
- `logical_name` — the logical name.
- `file_path` — resolved file path.
- `qualifier` — optional subsection qualifier (absent = full `# Public`).

## Algorithm

**Step 1 — Ancestors and target**: Walk upward from the target
using parent navigation, collecting each ancestor. Sort
alphabetically. The target is the last item; the rest are
ancestors.

**Step 2 — Dependencies**: Read the target's frontmatter.
For each `depends_on` entry, resolve the file path. If the
entry has a qualifier, record it. Verify the file exists.
Sort by file path, then by qualifier.

**Step 3 — Code files**: Read the target's frontmatter.
For each output, check if the file exists on disk. Include
only existing files.

**Step 4 — Normalize paths**: Convert all file paths to
forward slashes.

**Step 5 — Deduplicate**: Remove duplicate entries (same file
path and qualifier). When an entry without a qualifier exists
(full `# Public`), it subsumes entries with specific qualifiers
for the same file.

## Error conditions

| Condition | Description |
|---|---|
| Invalid logical name | Cannot resolve the logical name to a file path. |
| Unreadable frontmatter | Frontmatter parsing fails for any node in the chain. |
| Missing dependency | A `depends_on` target does not exist on disk. |

## Contracts

- All returned file paths use forward slashes.
- Deduplication keeps the first occurrence.
