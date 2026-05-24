---
depends_on:
  - ROOT/functional/utils/chain_resolution
  - ROOT/functional/utils/content_hash
external:
  - path: CHAIN_HASH.md
outputs:
  - id: chain_hash
    path: code-from-spec/functional/utils/chain_hash/output.md
---

# ROOT/functional/utils/chain_hash

Computes the chain hash for a given target node. The chain
hash determines whether an artifact is stale.

# Public

## Behavior

### Input

A target logical name.

### Output

A 27-character base64url hash string.

## Algorithm

Compute a content hash for each position in the chain
(in chain assembly order), then hash the concatenation
of all content hashes.

### Content hashes

Each position contributes a content hash — the SHA-1 of
the content that position injects into the chain:

| Position | Content hashed |
|---|---|
| Ancestor | `# Public` section (including the heading) |
| Target `# Public` | `# Public` section (including the heading) |
| Target `# Agent` | `# Agent` section (including the heading) |
| `depends_on: ROOT/x/y` | `# Public` section of referenced node |
| `depends_on: ROOT/x/y(z)` | `## z` subsection of `# Public` |
| `depends_on: ARTIFACT/x/y(id)` | Full artifact content, excluding frontmatter |
| `external` (whole file) | Full file content |
| `external` (with fragments) | Concatenation of each fragment's content, in declaration order |
| `input: ARTIFACT/x/y(id)` | Full artifact content, excluding frontmatter |

### Chain hash

The chain hash is the SHA-1 of the concatenation of all
content hashes (as raw bytes, not encoded) in chain
assembly order:

1. Each ancestor's content hash.
2. Each `depends_on` entry's content hash (alphabetical by path).
3. Each `external` entry's content hash (alphabetical by path).
4. Target `# Public` content hash.
5. Target `# Agent` content hash.
6. `input` content hash (if present).

The result is encoded as base64url (27 characters).

## Error conditions

| Condition | Description |
|---|---|
| Chain resolution failure | Cannot resolve the chain for the target. |
| Unreadable content | A file in the chain cannot be read. |
| Missing section | Target has no `# Public` or `# Agent` (use empty content). |
