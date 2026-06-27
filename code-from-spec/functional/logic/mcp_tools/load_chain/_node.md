---
depends_on:
  - SPEC/functional/logic/chain/resolver
  - SPEC/functional/logic/chain/hash
  - SPEC/functional/logic/parsing/node_parsing
  - SPEC/functional/logic/parsing/frontmatter
  - SPEC/functional/logic/os/file
  - SPEC/functional/logic/os/path_utils(interface)
  - SPEC/functional/logic/utils/logical_names(interface)
  - SPEC/functional/logic/utils/text_normalization(interface)
output: code-from-spec/functional/logic/mcp_tools/load_chain/output.md
---

# SPEC/functional/logic/mcp_tools/load_chain

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

Spec node content is boundary-normalized using the same
block extraction rules as the chain hash: leading blank
lines after the heading are removed, trailing blank
lines are removed, content ends with exactly one LF.
When concatenating multiple blocks (e.g. subsections),
heading lines have trailing whitespace removed, and
consecutive blocks are separated by exactly one blank
line. This ensures the delivered content matches what
is hashed — hash and delivery never diverge.

The rendered output of each chain entry is separated
from the next by exactly one blank line. This applies
between consecutive ancestors, between the last ancestor
and the first dependency, between consecutive
dependencies, and so on through the entire chain
assembly order.

For each entry, use `NodeParse` for spec nodes and
`file` for artifacts and external files.

**Ancestors** (from `chain.ancestors`)

For each ancestor, call `NodeParse` with
`ancestor.unqualified_logical_name`. If `node.public` is absent
or has no subsections, skip. Otherwise, include the
block-extracted and concatenated `## subsections` in
document order. Do not include the `# Public` heading
or any content directly under `# Public` (before the
first `##`).

**Dependencies** (from `chain.dependencies`)

For each dependency:
- If `LogicalNameIsArtifact(dep.unqualified_logical_name)`: open
  the file at `dep.file_path` with `FileOpen` (mode `"read"`, timeout 30000), read each
  line at a time, ignore the artifact tag line (the first
  line containing `code-from-spec:`), include all other
  lines. Call `FileClose`.
- If `LogicalNameIsExternal(dep.unqualified_logical_name)`: open
  the file at `dep.file_path` with `FileOpen` (mode `"read"`, timeout 30000), read all
  content. Call `FileClose`.
- If `LogicalNameIsSpec(dep.unqualified_logical_name)` and
  `dep.qualifier` is absent: call `NodeParse` with
  `dep.unqualified_logical_name`, include the block-extracted and
  concatenated `## subsections` in document order. Do
  not include the `# Public` heading or content
  directly under `# Public`.
- If `LogicalNameIsSpec(dep.unqualified_logical_name)` and
  `dep.qualifier` is present: call `NodeParse` with
  `dep.unqualified_logical_name`, find the subsection in
  `node.public` whose `heading` matches
  `NormalizeText(dep.qualifier)`, include the
  block-extracted `## subsection` heading and content.

**Target Public** (from `chain.target`)

First, emit a reduced frontmatter block containing only
the `output` field (formatted as YAML between `---`
delimiters). Then call `NodeParse` with
`chain.target.unqualified_logical_name`. If `node.public` is
present and has subsections, include the block-extracted
and concatenated `## subsections` in document order.
Do not include the `# Public` heading or content
directly under `# Public`. If absent or no subsections,
skip.

**Target Agent**

From the same `NodeParse` result, include the
block-extracted `# Agent` section: heading line (trailing
whitespace removed), section content, all `## subsection`
headings and their content (block-extracted, separated
by blank lines). If absent, skip.

### Step 4 — Assemble output string

Build the output string by concatenating sections with
delimiter lines:

1. First line: `chain_hash: <hash>` (the 27-character
   hash from Step 2).

2. A line containing exactly `--- context ---`.

3. The context stream from Step 3.

4. If `chain.input` is present: a line containing
   exactly `--- input ---`, followed by the content of
   the input file. For `ARTIFACT/` input, read each line
   at a time, ignore the artifact tag line (the first
   line containing `code-from-spec:`), include all other
   lines. For `EXTERNAL/` input, the full file content
   is included.

5. If the output file (at `frontmatter.output`) exists
   on disk and is readable: a line containing exactly
   `--- existing artifact ---`, followed by the full
   file content (read with `FileOpen` (mode `"read"`, timeout 30000)/`FileReadLine`/
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
