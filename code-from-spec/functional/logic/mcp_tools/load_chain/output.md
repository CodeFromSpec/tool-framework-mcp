<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@JGEUbhQgMJONpjBHO0nuv26vdoI -->

## Namespace

    namespace: mcploadchain

## Interface

```
record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
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

| Field | Always present | Content |
|---|---|---|
| `chain_hash` | yes | The 27-character base64url chain hash. |
| `context` | yes | All chain content concatenated as a single stream. |
| `input` | only if `input` field exists | Content of the input artifact, excluding frontmatter. |

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

### Step 4 — Extract input

If `chain.input` is present, open the file at
`chain.input.file_path` with `FileOpen`, strip
frontmatter (if present), read remaining content.
Call `FileClose`. Store as the `input` field of the
result.

If `chain.input` is absent, `input` is absent in the
result.

### Step 5 — Return result

Return `MCPLoadChainResult` with `chain_hash`,
`context` (the concatenated stream), and `input`.

## Contracts

- Returns everything in one call — no pagination.
- If any file in the chain is unreadable, returns an
  error (no partial results).
- The context stream contains no metadata or structural
  markers — only spec content, except for the target's
  reduced frontmatter block.
- Input is separated from context so the subagent can
  distinguish context (what informs) from input (what
  to transform).
