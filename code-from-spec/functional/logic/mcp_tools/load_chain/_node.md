---
depends_on:
  - ROOT/functional/logic/chain/resolver
  - ROOT/functional/logic/chain/hash
  - ROOT/functional/logic/parsing/node_parsing
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/os/path_utils(interface)
  - ROOT/functional/logic/utils/logical_names(interface)
  - ROOT/functional/logic/utils/text_normalization(interface)
output: code-from-spec/functional/logic/mcp_tools/load_chain/output.md
---

# ROOT/functional/logic/mcp_tools/load_chain

Loads the complete spec chain for a given node and
returns everything the subagent needs in a single
formatted string.

# Public

## Namespace

    namespace: mcploadchain

## Interface

```
function MCPLoadChain(logical_name: string) -> string
  errors:
    - NoOutput: target node has no output field.
    - InvalidOutputPath: the output path fails path
      validation.
    - (LogicalNames.*): propagated from
      LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the target node. |

### Output

A single string with sections separated by delimiter
lines. The format is:

```
chain_hash: <27-character hash>
--- context ---
<context content>
--- input ---
<input content>
--- existing artifact ---
<existing artifact content>
```

The `--- input ---` section is only present when the
target node's frontmatter has a non-empty `input` field.

The `--- existing artifact ---` section is only present
when the output file exists on disk and is readable.

# Agent

## Behavior

### Step 1 — Validate and resolve

Resolve the logical name to a file path using
`LogicalNameToPath`. If it fails, propagate the error.

Read the target node's frontmatter using
`FrontmatterParse`. If `frontmatter.output` is empty,
raise NoOutputs. Validate the output path using
`PathValidateCfs`. If it fails, raise
InvalidOutputPath.

Call `ChainResolve(logical_name)` to get the resolved
`Chain`. If it fails, propagate the error.

### Step 2 — Compute chain hash

Call `ChainHashCompute(chain)` with the resolved Chain.
If it fails, propagate the error. Store the result as
`chain_hash`.

### Step 3 — Build context stream

The context is a single continuous text block — no
delimiters, no file boundaries. Content is concatenated
in chain assembly order.

When reconstructing content from lines (whether from
`NodeParse` content lists or `FileReadLine`), append
`\n` after each line, including the last.

For each position, use `NodeParse` for spec nodes and
`file_reader` for artifacts and external files.

**Ancestors** (from `chain.ancestors`)

For each ancestor, call `NodeParse` with
`ancestor.logical_name`. If `node.public` is absent
or has empty content and no subsections, skip.
Otherwise, include the `# Public` raw heading, the
public section content, all `## subsection` raw
headings, and their content.

**Dependencies** (from `chain.dependencies`)

For each dependency:
- If `LogicalNameIsArtifact(dep.logical_name)`: open
  the file at `dep.file_path` with `FileOpen`, strip
  frontmatter (if present), include remaining content.
  Call `FileClose`.
- Else if `dep.qualifier` is absent: call `NodeParse`
  with `dep.logical_name`, include `# Public` raw
  heading, section content, all `## subsection` raw
  headings, and their content.
- Else: call `NodeParse` with `dep.logical_name`, find
  the subsection in `node.public` whose `heading`
  matches `NormalizeText(dep.qualifier)`, include the
  `## subsection` raw heading and its content.

**External** (from `chain.external`)

For each external entry, create a `PathCfs` from the
entry's `path`. Open with `FileOpen`, read all content.
Call `FileClose`.

**Target Public** (from `chain.target`)

First, emit a reduced frontmatter block containing only
the `output` field (formatted as YAML between `---`
delimiters). Then call `NodeParse` with
`chain.target.logical_name`. If `node.public` is
present, include `# Public` raw heading, section
content, all `## subsection` raw headings, and their
content. If absent, skip.

**Target Agent**

From the same `NodeParse` result, include `# Agent`
raw heading, section content, all `## subsection` raw
headings, and their content. If absent, skip.

### Step 4 — Assemble output string

Build the output string by concatenating sections with
delimiter lines:

1. First line: `chain_hash: <hash>` (the 27-character
   hash from Step 2).

2. A line containing exactly `--- context ---`.

3. The context stream from Step 3.

4. If `chain.input` is present: a line containing
   exactly `--- input ---`, followed by the content of
   the input artifact file (frontmatter stripped, read
   with `FileOpen`/`FileReadLine`/`FileClose`).

5. If the output file (at `frontmatter.output`) exists
   on disk and is readable: a line containing exactly
   `--- existing artifact ---`, followed by the full
   file content (read with `FileOpen`/`FileReadLine`/
   `FileClose`, no frontmatter stripping). If the file
   does not exist or cannot be read, omit this section
   silently (no error).

Return the assembled string.

## Contracts

- Returns everything in one call — no pagination.
- If any file in the chain is unreadable, returns an
  error (no partial results). The existing artifact is
  the exception: if it is unreadable or absent, its
  section is simply omitted.
- The context stream contains no metadata or structural
  markers — only spec content, except for the target's
  reduced frontmatter block.
- Delimiter lines separate sections so the consumer can
  parse them: `--- context ---`, `--- input ---`,
  `--- existing artifact ---`.
