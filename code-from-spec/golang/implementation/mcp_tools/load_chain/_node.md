---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
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
- Propagated errors from `parsing`, `chainresolver`,
  `chainhash`, `oslayer`, `manifest` packages.

# Agent

Implement the load chain tool as a Go package.

## Logic

### Step 1 — Validate and resolve

1. Call `parsing.ParseNode(logical_name)` to read and
   parse the target node. If it fails, propagate the
   error. If `node.Frontmatter.Output` is nil,
   return error ErrNoOutput. Call
   `oslayer.ValidateCfsPath(*node.Frontmatter.Output)`.
   If it fails, return ErrInvalidOutputPath.

3. Check if the artifact is modified:
   Call `manifest.OpenManifest(true)`. If it succeeds,
   look up the artifact logical name (strip "SPEC/"
   from logical_name, prepend "ARTIFACT/") in
   m.Entries. If an entry exists:
     Construct oslayer.CfsPath from `*node.Frontmatter.Output`. Try
     to read the file on disk and compute its SHA-1
     hash (base64url, 27 chars) using the same
     normalization as validate_specs. If the file
     exists and its hash does not match
     entry.Checksum, return ErrArtifactModified.
   If OpenManifest fails or the entry does not exist
   or the file does not exist, skip this check.

4. Call `chainresolver.ChainResolve(logical_name)` to get the
   resolved `Chain`. If it fails, propagate the error.

### Step 2 — Compute chain hash

5. Call `chainhash.ChainHashCompute(chain)` with the resolved
   chain. If it fails, propagate the error. Store the
   result as `chain_hash`.

### Step 3 — Build XML document

6. Build the XML document. Use a string builder.

   Start with: "chain_hash: <chain_hash>\n"
   Append: "<chain>\n"

   **Existing artifact** (optional):
   If the file at `*node.Frontmatter.Output` exists and is
   readable:
     Call `oslayer.OpenFile` with the `oslayer.CfsPath` of
     `*node.Frontmatter.Output` in "read" mode with
     timeout 30000. Read all lines with
     `handle.ReadLine()` until `oslayer.ErrEndOfFile`. Call
     `handle.Close()`.
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

   For each `ancestor` in `chain.Ancestors` (in order):
     Call `parsing.ParseNode(ancestor.LogicalName)`.
     If `node.public` is absent or
     `node.public.subsections` is empty, skip.
     Otherwise:
       Extract the content.
       Append: `<entry name="<ancestor.LogicalName>">\n`
       Append the extracted content.
       Append: `</entry>\n`

   For each `dep` in `chain.Dependencies` (in order):
     If dep.LogicalName starts with
     "ARTIFACT/":
       Read the full file at oslayer.CfsPath(dep.Path).
       Append: `<entry name="<dep.LogicalName>">\n`
       Append the full content.
       Append: `</entry>\n`
     Else if dep.LogicalName starts with
     "EXTERNAL/":
       Read the full file at oslayer.CfsPath(dep.Path).
       Append: `<entry name="<dep.LogicalName>">\n`
       Append the full content.
       Append: `</entry>\n`
     Else if dep.LogicalName starts with
     "SPEC/":
       Call `parsing.ParseNode(dep.LogicalName)`.
       Extract content (with qualifier if present).
       If content is non-empty:
         Let entry_name = dep.LogicalName.
         If dep.Qualifier is not nil, append
         "(<*dep.Qualifier>)" to entry_name.
         Append: `<entry name="<entry_name>">\n`
         Append the extracted content.
         Append: `</entry>\n`

   For the target node `chain.Target`:
     Call
     `parsing.ParseNode(chain.Target.LogicalName)`.
     If `node.public` is present and
     `node.public.subsections` is non-empty:
       Extract the content.
       Append: `<entry name="<chain.Target.LogicalName>">\n`
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
   If `chain.Input` is not nil:
     Append: "<input>\n"
     If chain.Input.LogicalName starts
     with "ARTIFACT/":
       Read full file. Append content.
     Else if chain.Input.LogicalName
     starts with "EXTERNAL/":
       Read full file. Append content.
     Else if chain.Input.LogicalName
     starts with "SPEC/":
       Call `parsing.ParseNode(chain.Input.LogicalName)`.
       Extract content (with qualifier if present,
       same rules as for SPEC dependencies).
       Append content.
     Append: "</input>\n"

   Append: "</chain>\n"

7. Return the assembled string.

## Go-specific guidance

- Use the `chainresolver` package for `ChainResolve`
  and `Chain`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for file checksum
  computation in the modified check.
- Use the `parsing` package for `ParseNode`,
  `NormalizeText`, `Node`, `NodeSection`,
  `NodeSubsection`, and
  `NodeFrontmatter`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, `ValidateCfsPath`, and
  `CfsPath`.
- Type checks on `LogicalName` use string
  prefix comparisons (`strings.HasPrefix`).
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
