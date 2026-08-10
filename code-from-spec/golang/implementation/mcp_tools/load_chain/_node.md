---
depends_on:
  - ARTIFACT/external/code-from-spec/chain-assembly-details
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcploadchain/mcploadchain.go
---

# SPEC/golang/implementation/mcp_tools/load_chain

Loads the complete spec chain for a given node and
returns everything the subagent needs in a single
formatted string. The target node is identified by an
opaque token (see `mcp_tools/create_token`), not a raw
logical name, so a subagent cannot request the chain of
a node other than the one it was dispatched for.

# Public

## Package

`package mcploadchain`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcploadchain"`

## Interface

```go
func MCPLoadChain(token string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `token` | yes | Opaque token identifying the target node, as returned by `create_token`. |

### Output

An XML document as a string, as defined in
CHAIN_ASSEMBLY.md. The document has these sections in
this order:

1. **`<previous_constraints>`** — old content for
   `<constraints>` positions that changed or were
   removed. Present only when cache is available and
   the existing artifact is present on disk.
2. **`<previous_references>`** — old content for
   `<references>` positions (imports) that changed or
   were removed. Present only when cache is available
   and the existing artifact is present on disk.
3. **`<previous_instructions>`** — previous `# Agent`
   content. Present only when cache is available, the
   existing artifact is present, and instructions
   changed or were removed.
4. **`<previous_input>`** — old content for `input`
   entries that changed or were removed. Contains an
   `<entry>` element per affected entry, with a `name`
   attribute and a `disposition` of `changed` or
   `removed`. Present only when cache is available, the
   existing artifact is present, and at least one
   `input` entry changed or was removed.
5. **`<existing_artifact>`** — current content of the
   artifact file on disk. Present only when the file
   exists.
6. **`<constraints>`** — the inheritance line: ancestors
   from root to the target's parent, plus the target
   node's own `# Public` as the last entry. Each
   position is an `<entry>` element with a `name`
   attribute. When cache is available and the existing
   artifact is present, each entry carries a
   `disposition` attribute (`unchanged`, `changed`, or
   `added`).
7. **`<references>`** — the target node's `imports`,
   in alphabetical order. Same `<entry>` shape and
   disposition rules as `<constraints>`. Present only
   when the node declares `imports`.
8. **`<instructions>`** — the target node's `# Agent`
   section (heading not included). Present only when
   the node has an `# Agent` section. May carry a
   `disposition` attribute.
9. **`<input>`** — the content referenced by the target
   node's `input` field. Each position is an `<entry>`
   element with a `name` attribute, one per `input`
   reference. Present only when the node declares
   `input`. When cache is available and the existing
   artifact is present, each entry may carry a
   `disposition` attribute.

For nodes with `type: verdict`, the document contains
only `<constraints>`, `<references>`, `<instructions>`,
and `<input>` — never `<existing_artifact>` or
`<previous_*>` sections. Entries carry no `disposition`
attributes.

### Errors

- `ErrNoOutput`: target node has no type field.
- `ErrInvalidOutputPath`: the output path fails path
  validation.
- `ErrModified`: the artifact or verdict file was
  modified outside the framework (checksum in manifest
  does not match file on disk). It must be accepted
  or deleted before regeneration.
- Propagated errors from `subagenttoken`, `parsing`,
  `chainresolver`, `chainhash`, `oslayer`, `manifest`
  packages.

# Agent

Implement the load chain tool as a Go package.

## Logic

### Step 0 — Resolve token

1. Call `subagenttoken.SubagentTokenValidate(token)` to
   recover the target node's logical name. If it fails,
   propagate the error. Store the result as
   `logical_name` for the remaining steps.

### Step 1 — Validate and resolve

2. Call `parsing.ParseNode(logical_name)` to read and
   parse the target node. If it fails, propagate the
   error. Let `resolved_output` =
   `parsing.ResolvedOutput(node)`. If `resolved_output`
   is nil, return error ErrNoOutput. Call
   `oslayer.ValidateStringIsCfsPath(*resolved_output)`.
   If it fails, return ErrInvalidOutputPath.

3. Determine whether this is a verdict node:
   Let `is_verdict` = (*node.Frontmatter.Type == "verdict").

4. Check if the artifact or verdict is modified:
   Derive the manifest key: strip "SPEC/" from
   logical_name, prepend "VERDICT/" if is_verdict,
   otherwise "ARTIFACT/".
   Call `manifest.OpenManifest(true)`. If it succeeds,
   look up the manifest key in m.Entries. If an entry
   exists:
     Construct oslayer.CfsPath from `*resolved_output`. Try
     to read the file on disk and compute its SHA-1
     hash (base64url, 27 chars) using the same
     normalization as validate_specs. If the file
     exists and its hash does not match
     entry.Checksum, return ErrModified.
   If OpenManifest fails or the entry does not exist
   or the file does not exist, skip this check.

5. Call `spectree.SpecTreeScan()`. If it fails, propagate
   the error. Build `knownSpecNodes` as a `[]string`:
   for each ref, append ref.LogicalName.

   Call `chainresolver.ChainResolve(logical_name, knownSpecNodes)` to get the
   resolved `Chain`. If it fails, propagate the error.

### Step 2 — Compute content hashes

6. Call `chainhash.ChainHashCompute(chain)` with the resolved
   chain. It returns `(chain_hash, positions, err)`.
   If it fails, propagate the error. Store
   `chain_hash` and `positions`.

### Step 3 — Build XML document

7. Build the XML document. Use a string builder.

   Append: "<chain>\n"

   See `ARTIFACT/external/code-from-spec/chain-assembly-details`
   for the exact XML section order, presence conditions,
   and a worked example. Follow it precisely.

   The following four `<previous_*>` sections and the
   `<existing_artifact>` section are skipped entirely
   when `is_verdict` is true. Disposition attributes on
   `<constraints>`, `<references>`, `<instructions>`,
   and `<input>` entries are also omitted for verdict
   nodes.

   **Previous constraints** (optional, artifact only):
   If not is_verdict, and cache is available and the
   existing artifact is present on disk: for each position among ancestors
   and the target's `# Public` whose cached content
   hash differs from its current hash, or which is no
   longer present in the current chain (removed), look
   up its old content in the cache by the cached hash.
   Emit one `<entry name="..." disposition="changed">`
   (or `disposition="removed"` if no longer present) per
   such position, containing the old content, all wrapped
   together in a single
   `<previous_constraints>...</previous_constraints>`
   block. Positions whose hash is unchanged are omitted
   entirely. Omit the whole block if there is nothing to
   report.

   **Previous references** (optional, artifact only):
   If not is_verdict, and cache is available and the
   existing artifact is present on disk: for each position among imports
   whose cached content hash differs from its current
   hash, or which is no longer present in the current
   chain (removed), look up its old content in the
   cache by the cached hash. Emit one
   `<entry name="..." disposition="changed">` (or
   `disposition="removed"` if no longer present) per
   such position, containing the old content, all
   wrapped together in a single
   `<previous_references>...</previous_references>`
   block. Positions whose hash is unchanged are omitted
   entirely. Omit the whole block if there is nothing
   to report.

   **Previous instructions** (optional, artifact only):
   If not is_verdict, and cache is available, the
   existing artifact is present, and the target's `# Agent` content hash
   differs from its cached hash (or the node no longer
   has an `# Agent` section): look up the old `# Agent`
   content in the cache. The element is
   `<previous_instructions disposition="changed">` (or
   `disposition="removed"` if the section is gone),
   with the old content as its entire body. Append
   `<previous_instructions disposition="...">`, the old
   content, then `</previous_instructions>`. Omit the
   whole block otherwise.

   **Previous input** (optional, artifact only):
   If not is_verdict, and cache is available and the
   existing artifact is present on disk: for each position among `chain.Input`
   entries whose cached content hash differs from its
   current hash, or which is no longer present in the
   current chain (removed), look up its old content in the
   cache by the cached hash. Emit one
   `<entry name="..." disposition="changed">` (or
   `disposition="removed"` if no longer present) per such
   position, containing the old content, all wrapped
   together in a single
   `<previous_input>...</previous_input>` block. Entries
   whose hash is unchanged are omitted entirely. Omit the
   whole block if there is nothing to report.

   **Existing artifact** (optional, artifact only):
   If not is_verdict, and the file at
   `*resolved_output` exists and is readable:
     Call `oslayer.OpenFile` with the `oslayer.CfsPath` of
     `*resolved_output` in "read" mode with
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
   ancestors, SPEC imports, and the target's
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

   **References** (optional):
   If `chain.Imports` is non-empty:
     Append: "<references>\n"
     For each `dep` in `chain.Imports` (in order):
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
     Append: "</references>\n"

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
   If `chain.Input` is non-empty:
     Append: "<input>\n"
     For each `inp` in `chain.Input` (in order):
       Let entry_name = inp.LogicalName. If
       inp.Qualifier is not nil, append
       "(<*inp.Qualifier>)" to entry_name.
       If inp.LogicalName starts with "ARTIFACT/":
         Read the full file at oslayer.CfsPath(inp.Path).
         Append: `<entry name="<entry_name>">\n`
         Append the full content.
         Append: `</entry>\n`
       Else if inp.LogicalName starts with "EXTERNAL/":
         Read the full file at oslayer.CfsPath(inp.Path).
         Append: `<entry name="<entry_name>">\n`
         Append the full content.
         Append: `</entry>\n`
       Else if inp.LogicalName starts with "SPEC/":
         Call `parsing.ParseNode(inp.LogicalName)`.
         Extract content (with qualifier if present,
         same rules as for SPEC imports).
         If content is non-empty:
           Append: `<entry name="<entry_name>">\n`
           Append the extracted content.
           Append: `</entry>\n`
     Append: "</input>\n"

   Append: "</chain>\n"

8. Return the assembled string.

### Step 4 — Write to cache (artifact only)

   Skip this entire step if `is_verdict` is true.

9. Build a map from position label to extracted content:
   during Step 3, each time content is extracted for a
   constraints entry, references entry, instructions,
   or input, store the extracted content string in a
   map keyed by the label that `ChainHashCompute`
   would use for that position:
   - Ancestors: the logical name.
   - Imports (references entries): the logical name
     (with qualifier if present for SPEC/ imports).
   - ARTIFACT/ and EXTERNAL/ imports: the logical
     name.
   - Target node's `# Public`: the logical name.
   - Target node's `# Agent`:
     `"AGENT[" + logicalName + "]"`.
   - Input: `"INPUT[" + referenceName + "]"` per entry
     (with qualifier if present).

10. For each position in `positions` (from Step 2):
    Look up position.Label in the content map. If
    found, call `cache.WriteContent(position.Hash,
    content)`. Ignore errors — cache is best-effort.

11. Call `cache.WriteChain(chain_hash, positions)`.
    Ignore errors.

## Go-specific guidance

- Use the `subagenttoken` package for
  `SubagentTokenValidate`.
- Use the `spectree` package for `SpecTreeScan`.
- Use the `chainresolver` package for `ChainResolve`
  and `Chain`.
- Use the `chainhash` package for `ChainHashCompute`
  and `ContentHash`.
- Use the `cache` package for `WriteContent` and
  `WriteChain`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for file checksum
  computation in the modified check.
- Use the `parsing` package for `ParseNode`,
  `ResolvedOutput`, `NormalizeText`, `Node`,
  `NodeSection`, `NodeSubsection`, and
  `NodeFrontmatter`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, `ValidateStringIsCfsPath`, and
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
