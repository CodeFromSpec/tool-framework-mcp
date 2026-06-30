---
depends_on:
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/chainhash/chainhash.go
---

# SPEC/golang/implementation/chain/hash

Computes the chain hash for a resolved chain by reading
all chain positions from disk and hashing their content.

# Public

## Package

`package chainhash`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/chainhash"`

## Interface

```go
type ContentHash struct {
    Label string
    Hash  string
}

func ChainHashCompute(chain chainresolver.Chain) (string, []ContentHash, error)
```

### ContentHash

A content hash for a single chain position. `Label`
identifies the position using the same conventions as
the cache chain file format:

- Ancestors, dependencies, and the target node's
  `# Public`: the logical name directly
  (e.g. `SPEC/payments`, `SPEC/x(qualifier)`,
  `ARTIFACT/y`, `EXTERNAL/z`).
- Target node's `# Agent`: `AGENT[<logical-name>]`.
- Input: `INPUT[<reference>]`.

`Hash` is the 27-character base64url encoded SHA-1
of the position's processed content.

### ChainHashCompute

Receives a `Chain` (as returned by `ChainResolve`) and
returns the 27-character chain hash and the ordered
list of content hashes for each position that
contributed to the chain.

### Errors

- `ErrParseFailure`: a node file cannot be parsed.
- Propagated errors from `oslayer`, `parsing` packages.

# Agent

Implement the chain hash component as a Go package.

## Logic

### Content hash helpers

Uses `parsing.ConcatenateSubsections`,
`parsing.FormatSection`, `parsing.ExtractAgentContent`,
and `parsing.ReadFileContent` for content extraction.

**HashPublicSubsections(node: parsing.Node) -> optional raw bytes (20)**

1. If `node.public` is absent, return absent.
2. If `node.public.subsections` is empty, return absent.
3. Let `text` = `parsing.ConcatenateSubsections(node.public.subsections)`.
4. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashQualifiedSubsection(node: parsing.Node, qualifier: string) -> optional raw bytes (20)**

1. Let `normalized_qualifier` = `parsing.NormalizeText(qualifier)`.
2. Find the subsection in `node.public.subsections`
   whose `heading` equals `normalized_qualifier`.
3. If not found, return absent.
4. Let `text` = `parsing.FormatSection(subsection.raw_heading,
   subsection.content)`.
5. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashAgentSection(node: parsing.Node) -> optional raw bytes (20)**

1. Let `text` = `parsing.ExtractAgentContent(node)`.
2. If `text` is empty, return absent.
3. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashFileContent(file_path: oslayer.CfsPath) -> raw bytes (20)**

1. Let `text` = `parsing.ReadFileContent(file_path)`.
   If it fails, propagate the error.
2. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

### Main function

**ChainHashCompute(chain: chainresolver.Chain) -> (string, []ContentHash)**

**Step 1 — Collect content hashes**

Let `hashes` = empty list of raw byte sequences
(each 20 bytes).
Let `positions` = empty list of ContentHash.

Helper: `recordPosition(label, rawHash)`:
  Encode rawHash as base64url (27 chars).
  Append ContentHash{Label: label, Hash: encoded}
  to `positions`.
  Append rawHash to `hashes`.

1. For each `ancestor` in `chain.Ancestors` (from root
   to target's parent):
   a. Call `parsing.ParseNode(ancestor.LogicalName)`.
      If it fails, raise ErrParseFailure.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, call
      `recordPosition(ancestor.LogicalName, h)`.

2. For each `dep` in `chain.Dependencies` (already
   sorted alphabetically by logical name):
   a. Let `label` = dep.LogicalName. If dep.Qualifier
      is not nil, append "(" + *dep.Qualifier + ")"
      to label.
   b. If dep.LogicalName starts with "ARTIFACT/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(dep.Path))`.
      Call `recordPosition(label, h)`.
   c. Else if dep.LogicalName starts with "EXTERNAL/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(dep.Path))`.
      Call `recordPosition(label, h)`.
   d. Else if dep.LogicalName starts with "SPEC/":
      Call `parsing.ParseNode(dep.LogicalName)`.
      If it fails, raise ErrParseFailure.
      If `dep.Qualifier` is nil:
        Let `h` = `HashPublicSubsections(node)`.
        If `h` is present, call
        `recordPosition(label, h)`.
      If `dep.Qualifier` is not nil:
        Let `h` = `HashQualifiedSubsection(node,
        *dep.Qualifier)`.
        If `h` is present, call
        `recordPosition(label, h)`.

3. Target `# Public`:
   a. Call `parsing.ParseNode(chain.Target.LogicalName)`.
      If it fails, raise ErrParseFailure.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, call
      `recordPosition(chain.Target.LogicalName, h)`.
   d. Save this `node` result as `target_node`.

4. Target `# Agent`:
   a. Let `h` = `HashAgentSection(target_node)`.
   b. If `h` is present, call
      `recordPosition("AGENT[" + chain.Target.LogicalName + "]", h)`.

5. If `chain.Input` is not nil:
   a. Append a single byte `0x49` (`I`) to
      `hashes` as a marker before the input content
      hash. Do NOT add a position entry for the marker.
   b. Let `input` = `chain.Input`.
   c. Let `inputLabel` = input.LogicalName. If
      input.Qualifier is not nil, append
      "(" + *input.Qualifier + ")" to inputLabel.
      Let `inputLabel` = "INPUT[" + inputLabel + "]".
   d. If input.LogicalName starts with "ARTIFACT/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(input.Path))`.
      Call `recordPosition(inputLabel, h)`.
   e. Else if input.LogicalName starts with "EXTERNAL/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(input.Path))`.
      Call `recordPosition(inputLabel, h)`.
   f. Else if input.LogicalName starts with "SPEC/":
      Call `parsing.ParseNode(input.LogicalName)`.
      If it fails, raise ErrParseFailure.
      If `input.Qualifier` is nil:
        Let `h` = `HashPublicSubsections(node)`.
        If `h` is present, call
        `recordPosition(inputLabel, h)`.
      If `input.Qualifier` is not nil:
        Let `h` = `HashQualifiedSubsection(node,
        *input.Qualifier)`.
        If `h` is present, call
        `recordPosition(inputLabel, h)`.

**Step 2 — Compute final hash**

1. Let `concatenated` = concatenation of all byte
   sequences in `hashes` in order. Each entry is
   20 bytes except the `0x49` marker which is 1 byte.
2. Compute SHA-1 of `concatenated`.
3. Encode the resulting 20 bytes as base64url (RFC 4648
   §5, no padding) — producing 27 characters.
4. Return the 27-character string and `positions`.

## Go-specific guidance

- Use the `chainresolver` package for `Chain`.
- Use the `parsing` package for `ParseNode`,
  `CfsReference`, `Node`, `NodeSection`,
  `NodeSubsection`, and `NormalizeText`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, and `CfsPath`.
- Type checks on `LogicalName` use string prefix
  comparisons (`strings.HasPrefix`).
- For SHA-1 and base64url, use `crypto/sha1` and
  `encoding/base64` (base64.RawURLEncoding).
- The package name should be `chainhash`.
