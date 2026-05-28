---
depends_on:
  - ROOT/functional/logic/chain/chain_hash
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/utils/text_normalization
  - ROOT/functional/logic/parsing/node_parsing
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: load_chain
    path: code-from-spec/functional/logic/mcp_tools/load_chain/output.md
---

# ROOT/functional/logic/mcp_tools/load_chain

Loads the complete spec chain for a given node and returns
the chain hash, context, and input as separate items.

Review status: pending

# Public

## Interface

```
function LoadChain(logical_name: string) -> list of text items
  errors:
    - invalid logical name: not a recognized ROOT/ reference.
    - no outputs: target node has no outputs field.
    - invalid output path: an output path fails path validation.
    - chain resolution failure: a dependency cannot be resolved.
    - unreadable file: a file in the chain cannot be read or parsed.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the target node. |

### Output

The result contains separate text items:

| Item | Always present | Content |
|---|---|---|
| Chain hash | yes | The 27-character base64url chain hash. |
| Context | yes | All chain content concatenated as a single stream. |
| Input | only if `input` field exists | Content of the input artifact, excluding frontmatter. |

# Agent

## Behavior

### Validation

Before loading the chain:
1. The logical name must be a valid `ROOT/` reference.
2. Read the frontmatter of the node identified by
   `logical_name`. It must have `outputs` declared.
3. Each output path must pass path validation.

### Context stream

The context is a single continuous text block — no
delimiters, no headers, no file boundaries. All files
are read using `file_reader` (close each reader after
reading). Content is concatenated in this exact order:

**Step 1 — Ancestors** (root to target's parent)

For each ancestor, from the root node down to the
target's direct parent, in tree depth order:
- Include the `# Public` section — both the direct
  content and all `##` subsections (with their
  headings). Omit only the `# Public` heading itself.
- If `# Public` is absent or has no content and no
  subsections, skip this ancestor entirely.

**Step 2 — Dependencies** (`depends_on`)

For each entry in the target's `depends_on`, in
alphabetical order by logical name:
- `ROOT/x/y` — include the `# Public` section content
  of the referenced node (without the heading).
- `ROOT/x/y(z)` — include only the `## z` subsection
  content within `# Public` of the referenced node.
- `ARTIFACT/x/y(id)` — include the full content of
  the referenced artifact file, excluding any
  frontmatter.

**Step 3 — External files** (`external`)

For each entry in the target's `external`, in
alphabetical order by path:
- If no `fragments` declared — include the full file
  content.
- If `fragments` declared — include only the content
  at the declared line ranges, concatenated in
  declaration order.

**Step 4 — Target `# Public`**

Preceded by a reduced frontmatter block containing only
`outputs`. Then include the target node's `# Public`
section — both the direct content and all `##`
subsections (with their headings). Omit only the
`# Public` heading itself.

**Step 5 — Target `# Agent`**

Include the target node's `# Agent` section — both the
direct content and all `##` subsections (with their
headings). Omit only the `# Agent` heading itself.
If the section is absent, skip.

### Input separation

If the target node has an `input` field, the referenced
artifact content is returned as a separate text item,
not concatenated into the context stream. This allows
the subagent to distinguish context (what informs) from
input (what to transform).

### Chain hash

Use `chain_hash.ComputeChainHash(logical_name)` to
compute the chain hash. Do not reimplement the hash
computation — use the shared utility.

The chain hash is returned as a separate text item so
the subagent can embed it in the artifact tag.

## Contracts

- Returns everything in one call — no pagination.
- If any file in the chain is unreadable, returns an
  error (no partial results).
- The context stream contains no metadata or structural
  markers — only spec content.
