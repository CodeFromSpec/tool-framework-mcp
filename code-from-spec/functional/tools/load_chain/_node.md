---
outputs:
  - id: load_chain
    path: code-from-spec/functional/tools/load_chain/output.md
---

# ROOT/functional/tools/load_chain

Loads the complete spec chain for a given node and returns it
as a single text response, including the chain hash for the
artifact tag.

# Public

## Behavior

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the target node. |

### Output

A formatted text block containing all files in the chain,
separated by heredoc-style delimiters. Each section includes
metadata headers (`node:`, `path:`) and the file content.

The response also includes the **chain hash** — the SHA-1
digest (base64url, 27 characters) computed from all positions
in the chain, as defined in `CHAIN_HASH.md`.

### Chain content

| Section | Content included |
|---|---|
| Ancestors | `# Public` body only (heading stripped). Skipped if empty. |
| Target | Full file with reduced frontmatter (only `outputs`). |
| Dependencies | `# Public` body or specific subsections per qualifier. Skipped if empty. |
| Code files | Existing source files as-is. Non-existing files omitted. |

### Validation

Before loading the chain:
1. The logical name must be a valid `ROOT/` reference.
2. The target node must have `outputs` declared.
3. Each output path must pass path validation.

## Error conditions

| Condition | Description |
|---|---|
| Invalid logical name | Not a recognized `ROOT/` reference. |
| No outputs | Target node has no `outputs` field. |
| Invalid output path | An output path fails path validation. |
| Chain resolution failure | A dependency cannot be resolved. |
| Unreadable file | A file in the chain cannot be read or parsed. |

## Contracts

- Returns the entire chain in one call — no pagination.
- If any file in the chain is unreadable, returns an error
  (no partial results).
