---
depends_on:
  - ROOT/functional/utils/logical_names
  - ROOT/functional/utils/parsing/frontmatter
outputs:
  - id: chain_resolution
    path: artifacts/functional/utils/chain_resolution/output.md
---

# ROOT/functional/utils/chain_resolution

Assembles the ordered chain of content that a generation
subagent receives for a given target node.

# Public

## Interface

```
record ChainPosition
  source: string
  logical_name: string
  file_path: string
  qualifier: optional string

function ResolveChain(logical_name) -> list of ChainPosition
  errors:
    - invalid logical name: cannot resolve the logical name to a file path.
    - unreadable frontmatter: frontmatter parsing fails for any node in the chain.
    - missing dependency: a depends_on target does not exist on disk.
    - missing external file: an external file does not exist on disk.
    - missing artifact: an ARTIFACT/ reference points to a file that does not exist.
```

# Agent

## Behavior

Given a target logical name, returns the chain — the complete,
ordered context needed to generate the target's artifacts.

### Chain assembly order

The chain is assembled in this exact order:

1. **Ancestors** — the `# Public` content of each ancestor
   from root to the target's parent. Ordered by tree depth
   (root first).

2. **Dependencies** (`depends_on`) — the target node's
   `depends_on` entries, in alphabetical order by path.
   What is included depends on the reference type:
   - `ROOT/x/y` — `# Public` section of the referenced node.
   - `ROOT/x/y(z)` — only the `## z` subsection of `# Public`.
   - `ARTIFACT/x/y(id)` — full content of the referenced
     artifact file, excluding any frontmatter.

3. **External files** (`external`) — the target node's
   `external` entries, in alphabetical order by path.
   For each entry, either the full file content or the
   declared fragments are included.

4. **Target `# Public`** — the target node's `# Public` section.

5. **Target `# Agent`** — the target node's `# Agent` section.

6. **Input** (`input`) — if the target node has an `input`
   field, the full content of the referenced artifact file,
   excluding any frontmatter.

### Example

Chain for `ROOT/payments/fees/calculation`:

```
ROOT                                [# Public]     <- ancestor
ROOT/payments                       [# Public]     <- ancestor
ROOT/payments/fees                  [# Public]     <- ancestor
ROOT/dependencies/database          [# Public]     <- depends_on
ARTIFACT/extraction/proto(proto)    [full]         <- depends_on
proto/payments/v1/transfers.proto   [full]         <- external
ROOT/payments/fees/calculation      [# Public]     <- target
ROOT/payments/fees/calculation      [# Agent]      <- target
ARTIFACT/functional/calc(calc)      [full]         <- input
```

The chain is the complete context. Nothing outside the chain
is needed. Nothing inside the chain is redundant.

## Contracts

- All returned file paths use forward slashes.
- The chain order is deterministic — same input always
  produces the same order.
- Ancestors are never deduplicated (each contributes once
  by tree position).
- Dependencies are deduplicated: when an entry without a
  qualifier exists (full `# Public`), it subsumes entries
  with specific qualifiers for the same file. Keep the
  first occurrence.
