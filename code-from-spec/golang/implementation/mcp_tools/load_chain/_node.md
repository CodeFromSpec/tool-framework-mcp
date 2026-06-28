---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
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

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcploadchain"`

## Interface

```go
func MCPLoadChain(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. |

### Output

An XML document as a string. The first line is the
chain hash: `chain_hash: <27-character hash>`. The
rest is the XML document.

The XML has up to four sections in this order:

1. **`<existing_artifact>`** — current content of the
   artifact file on disk. Present only when the file
   exists.
2. **`<constraints>`** — the current spec content. Each
   position is an `<entry>` element with a `name`
   attribute. Entries appear in chain assembly order:
   ancestors, then depends_on (sorted), then target
   node's `# Public`.
3. **`<instructions>`** — the target node's `# Agent`
   section (heading not included). Present only when
   the node has an `# Agent` section.
4. **`<input>`** — the content referenced by the target
   node's `input` field. Present only when the node
   declares `input`.

No `<previous_*>` sections or `disposition` attributes
in this version (cache is not implemented yet).

### Errors

- `ErrNoOutput`: target node has no output field.
- `ErrInvalidOutputPath`: the output path fails path
  validation.
- `ErrArtifactModified`: the artifact file was modified
  outside the framework (checksum in manifest does not
  match file on disk). The artifact must be accepted
  or deleted before regeneration.
- Propagated errors from `logicalnames`, `chainresolver`,
  `chainhash`, `parsenode`, `file`, `manifest` packages.

# Agent

Implement the load chain tool as a Go package.

## Logic

### Step 1 — Validate and resolve

1. Call `LogicalNameParse(logical_name)` to parse
   the target node. If it fails, propagate the error.
   Let `ln` be the result.

2. Call `FrontmatterParse(PathCfs{Value: ln.Path})`
   to read the target node's frontmatter. If
   `frontmatter.output` is empty, return error
   ErrNoOutput. Call `PathValidateCfs(frontmatter.output)`.
   If it fails, return ErrInvalidOutputPath.

3. Check if the artifact is modified:
   Call `ManifestOpen("read")`. If it succeeds, look
   up the artifact logical name (strip "SPEC/" from
   logical_name, prepend "ARTIFACT/") in
   manifest_handle.Entries. If an entry exists:
     Construct PathCfs from frontmatter.output. Try
     to read the file on disk and compute its SHA-1
     hash (base64url, 27 chars) using the same
     normalization as validate_specs. If the file
     exists and its hash does not match
     entry.Checksum, return ErrArtifactModified.
   If ManifestOpen fails or the entry does not exist
   or the file does not exist, skip this check.

4. Call `ChainResolve(logical_name)` to get the
   resolved `Chain`. If it fails, propagate the error.

### Step 2 — Compute chain hash

5. Call `ChainHashCompute(chain)` with the resolved
   chain. If it fails, propagate the error. Store the
   result as `chain_hash`.

### Step 3 — Build XML document

6. Build the XML document. Use a string builder.

   Start with: "chain_hash: <chain_hash>\n"
   Append: "<chain>\n"

   **Existing artifact** (optional):
   If the file at `frontmatter.output` exists and is
   readable:
     Call `FileOpen` with the `PathCfs` of
     `frontmatter.output` in "read" mode with
     timeout 30000. Read all lines with
     `FileReadLine` until `EndOfFile`. Call
     `FileClose`.
     Append: "<existing_artifact>\n"
     Append the full file content.
     Append: "</existing_artifact>\n"
     If the file does not exist or cannot be read,
     omit this section silently.

   **Constraints:**
   Append: "<constraints>\n"

   Helper for extracting SPEC content (used for
   ancestors, SPEC dependencies, and the target's
   Public): given a node and an optional qualifier,
   extract the content using the same boundary
   normalization rules defined in chain/hash:
   - No qualifier: concatenate all `##` subsections
     of `# Public` in document order. Each subsection
     rendered as raw_heading (trailing whitespace
     removed) + extracted content, separated by one
     blank line.
   - With qualifier: find the matching `##` subsection,
     render as raw_heading + content.

   For each `ancestor` in `chain.ancestors` (in order):
     Call `NodeParse(ancestor.unqualified_logical_name)`.
     If `node.public` is absent or
     `node.public.subsections` is empty, skip.
     Otherwise:
       Extract the content.
       Append: `<entry name="<ancestor.unqualified_logical_name>">\n`
       Append the extracted content.
       Append: `</entry>\n`

   For each `dep` in `chain.dependencies` (in order):
     If dep.unqualified_logical_name starts with
     "ARTIFACT/":
       Read the file at dep.file_path. Skip the first
       line containing "code-from-spec:" (artifact tag).
       Append: `<entry name="<dep.unqualified_logical_name>">\n`
       Append the file content (without tag line).
       Append: `</entry>\n`
     Else if dep.unqualified_logical_name starts with
     "EXTERNAL/":
       Read the full file at dep.file_path.
       Append: `<entry name="<dep.unqualified_logical_name>">\n`
       Append the full content.
       Append: `</entry>\n`
     Else if dep.unqualified_logical_name starts with
     "SPEC/":
       Call `NodeParse(dep.unqualified_logical_name)`.
       Extract content (with qualifier if present).
       If content is non-empty:
         Let entry_name = dep.unqualified_logical_name.
         If dep.qualifier is present, append
         "(<dep.qualifier>)" to entry_name.
         Append: `<entry name="<entry_name>">\n`
         Append the extracted content.
         Append: `</entry>\n`

   For the target node `chain.target`:
     Call
     `NodeParse(chain.target.unqualified_logical_name)`.
     If `node.public` is present and
     `node.public.subsections` is non-empty:
       Extract the content.
       Append: `<entry name="<chain.target.unqualified_logical_name>">\n`
       Append the extracted content.
       Append: `</entry>\n`

   Append: "</constraints>\n"

   **Instructions** (optional):
   Using the target node parsed above:
   If `node.agent` is present:
     Build agent content: the `# Agent` heading is
     NOT included. Include:
       `node.agent.content` (leading blank lines
       removed, trailing blank lines removed).
       For each subsection in
       `node.agent.subsections`:
         Separate from previous block with exactly
         one blank line.
         Add the subsection `raw_heading` (trailing
         whitespace removed) and content.
       Ensure ends with exactly one LF.
     Append: "<instructions>\n"
     Append the agent content.
     Append: "</instructions>\n"

   **Input** (optional):
   If `chain.input` is present:
     Append: "<input>\n"
     If chain.input.unqualified_logical_name starts
     with "ARTIFACT/":
       Read file, skip artifact tag line.
       Append content.
     Else if chain.input.unqualified_logical_name
     starts with "EXTERNAL/":
       Read full file. Append content.
     Else if chain.input.unqualified_logical_name
     starts with "SPEC/":
       Call `NodeParse(chain.input.unqualified_logical_name)`.
       Extract content (with qualifier if present,
       same rules as for SPEC dependencies).
       Append content.
     Append: "</input>\n"

   Append: "</chain>\n"

7. Return the assembled string.

## Go-specific guidance

- Use the `chainresolver` package for `ChainResolve` and
  the `Chain`, `ChainItem` records.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `ManifestOpen`,
  `ManifestHandle`, `ManifestEntry`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for file checksum
  computation in the modified check.
- Use the `parsenode` package for `NodeParse` and the
  `Node`, `NodeSection`, `NodeSubsection` records.
- Use the `file` package for `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
- Use the `frontmatter` package for `FrontmatterParse`
  and the `Frontmatter`, `FrontmatterExternal` records.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `logicalnames` package for `LogicalNameParse`.
  Type checks on `unqualified_logical_name` use string
  prefix comparisons (`strings.HasPrefix`).
- Use the `textnormalization` package for `NormalizeText`.
- The package name should be `mcploadchain`.
- Build the XML using string concatenation or
  `strings.Builder`. Do not use `encoding/xml` — the
  output is a simple structured document, not a
  general-purpose XML serialization.
- When reconstructing content from lines, append `\n`
  after each line including the last.
- XML element names and attribute names are lowercase
  with underscores (e.g. `existing_artifact`, not
  `ExistingArtifact`).
