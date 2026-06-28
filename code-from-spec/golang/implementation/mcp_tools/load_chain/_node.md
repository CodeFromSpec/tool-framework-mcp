---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/parsing/node_parsing
  - SPEC/golang/implementation/utils/logical_names
  - SPEC/golang/implementation/utils/text_normalization
output: internal/mcploadchain/mcploadchain.go
---

# SPEC/golang/implementation/mcp_tools/load_chain

Loads the complete spec chain for a given node and
returns everything the subagent needs in a single
formatted string.

# Public

## Package

`package mcploadchain`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"`

## Interface

```go
func MCPLoadChain(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. |

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

### Errors

- `ErrNoOutput`: target node has no output field.
- `ErrInvalidOutputPath`: the output path fails path
  validation.
- Propagated errors from `logicalnames`, `chainresolver`,
  `chainhash`, `parsenode`, `file` packages.

# Agent

Implement the load chain tool as a Go package.

## Logic

### Step 1 — Validate and resolve

1. Call `LogicalNameToPath(logical_name)` to get the
   target node's file path. If it fails, propagate
   the error.

2. Call `FrontmatterParse(target_file_path)` to read
   the target node's frontmatter. If
   `frontmatter.output` is empty, return error
   "NoOutput". Call `PathValidateCfs(frontmatter.output)`.
   If it fails, return error "InvalidOutputPath".

3. Call `ChainResolve(logical_name)` to get the
   resolved `Chain`. If it fails, propagate the error.

### Step 2 — Compute chain hash

4. Call `ChainHashCompute(chain)` with the resolved
   chain. If it fails, propagate the error. Store the
   result as `chain_hash`.

### Step 3 — Build context stream

5. Build the context stream:

   Set `context_parts` to an empty list of strings.

   For each `ancestor` in `chain.ancestors` (in order):
     Call `NodeParse(ancestor.unqualified_logical_name)`.
     If `node.public` is absent or
     `node.public.subsections` is empty, skip.
     Otherwise:
       Build `block` by concatenating all subsections
       in document order:
         For each subsection in
         `node.public.subsections`:
           Add the subsection `raw_heading` (trailing
           whitespace removed).
           Add each line in `subsection.content` with
           leading blank lines after the heading removed
           and trailing blank lines removed.
           Ensure the block ends with exactly one LF.
         Separate consecutive subsection blocks with
         exactly one blank line.
       Append `block` to `context_parts`.

   For each `dep` in `chain.dependencies` (in order):
     If `LogicalNameIsArtifact(dep.unqualified_logical_name)`
     is true:
       Call `FileOpen(dep.file_path, "read", 30000)`.
       Read all lines with `FileReadLine` until
       `EndOfFile`. Skip the first line that contains
       "code-from-spec:" (the artifact tag line).
       Include all other lines. Call `FileClose`.
       Append the resulting text to `context_parts`.
     Else if `LogicalNameIsExternal(dep.unqualified_logical_name)`
     is true:
       Call `FileOpen(dep.file_path, "read", 30000)`.
       Read all lines with `FileReadLine` until
       `EndOfFile`. Call `FileClose`.
       Append the full file content to `context_parts`.
     Else if `LogicalNameIsSpec(dep.unqualified_logical_name)`
     is true and `dep.qualifier` is absent:
       Call `NodeParse(dep.unqualified_logical_name)`.
       If `node.public` is absent or
       `node.public.subsections` is empty, skip.
       Otherwise:
         Build `block` by concatenating all subsections
         in document order (same boundary normalization
         rules as for ancestors).
         Append `block` to `context_parts`.
     Else if `LogicalNameIsSpec(dep.unqualified_logical_name)`
     is true and `dep.qualifier` is present:
       Call `NodeParse(dep.unqualified_logical_name)`.
       Compute `normalized_qualifier` =
       `NormalizeText(dep.qualifier)`.
       Find the subsection in
       `node.public.subsections` whose `heading`
       equals `normalized_qualifier`.
       If found:
         Build `block` from the subsection
         `raw_heading` (trailing whitespace removed)
         and its content (leading blank lines removed,
         trailing blank lines removed, ends with
         exactly one LF).
         Append `block` to `context_parts`.

   For the target node `chain.target`:
     Build a reduced frontmatter block:
       Line 1: "---"
       Line 2: "output: <frontmatter.output>"
       Line 3: "---"
     Append this block to `context_parts`.

     Call
     `NodeParse(chain.target.unqualified_logical_name)`.
     If `node.public` is present and
     `node.public.subsections` is non-empty:
       Build `block` by concatenating all subsections
       in document order (same boundary normalization
       rules as above).
       Append `block` to `context_parts`.

     If `node.agent` is present:
       Build `agent_block`:
         Add `node.agent.raw_heading` (trailing
         whitespace removed).
         Add each line in `node.agent.content`
         (leading blank lines removed, trailing blank
         lines removed).
         For each subsection in
         `node.agent.subsections`:
           Separate from previous block with exactly
           one blank line.
           Add the subsection `raw_heading` (trailing
           whitespace removed).
           Add each line in `subsection.content`
           (leading blank lines removed, trailing
           blank lines removed).
         Ensure the block ends with exactly one LF.
       Append `agent_block` to `context_parts`.

### Step 4 — Assemble output string

6. Assemble the output string:

   Start with line: "chain_hash: <chain_hash>"
   Append line: "--- context ---"
   Append the context stream: join all entries in
   `context_parts` separated by exactly one blank line.

   If `chain.input` is present:
     Append line: "--- input ---"
     If `LogicalNameIsArtifact(chain.input.unqualified_logical_name)`
     is true:
       Call `FileOpen(chain.input.file_path, "read",
       30000)`. Read all lines with `FileReadLine`
       until `EndOfFile`. Skip the first line that
       contains "code-from-spec:". Include all other
       lines. Call `FileClose`. Append the resulting
       text.
     Else (EXTERNAL/ or other):
       Call `FileOpen(chain.input.file_path, "read",
       30000)`. Read all lines with `FileReadLine`
       until `EndOfFile`. Call `FileClose`. Append
       the full file content.

   If the file at `frontmatter.output` exists and is
   readable:
     Append line: "--- existing artifact ---"
     Call `FileOpen` with the `PathCfs` of
     `frontmatter.output` in "read" mode with
     timeout 30000. Read all lines with
     `FileReadLine` until `EndOfFile`. Call
     `FileClose`. Append the full file content.
     If the file does not exist or cannot be read,
     omit this section silently.

7. Return the assembled output string.

## Go-specific guidance

- Use the `chainresolver` package for `ChainResolve` and
  the `Chain`, `ChainItem` records.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `parsenode` package for `NodeParse` and the
  `Node`, `NodeSection`, `NodeSubsection` records.
- Use the `file` package for `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
- Use the `frontmatter` package for `FrontmatterParse`
  and the `Frontmatter`, `FrontmatterExternal` records.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `logicalnames` package for `LogicalNameToPath`
  and `LogicalNameIsArtifact`.
- Use the `textnormalization` package for `NormalizeText`.
- The package name should be `mcploadchain`.
- `MCPLoadChainResult` is an exported struct.
- When reconstructing content from lines, append `\n`
  after each line including the last.
