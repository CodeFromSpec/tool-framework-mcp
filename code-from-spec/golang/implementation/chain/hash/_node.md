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
func ChainHashCompute(chain chainresolver.Chain) (string, error)
```

Receives a `Chain` (as returned by `ChainResolve`) and
returns a 27-character base64url encoded SHA-1 hash.

### Errors

- `ErrParseFailure`: a node file cannot be parsed.
- Propagated errors from `oslayer`, `parsing` packages.

# Agent

Implement the chain hash component as a Go package.

## Logic

### Block extraction helpers

**ExtractBlock(content: list of string) -> string**

1. Remove leading blank lines from `content`
   (lines that are empty or contain only spaces and tabs).
2. Remove trailing blank lines.
3. If nothing remains, return empty string.
4. Join remaining lines with `\n` and append exactly
   one `\n`.

**FormatSection(raw_heading: string, content: list of string) -> string**

1. Let `head` = `raw_heading` with trailing whitespace
   removed, followed by `\n`.
2. Let `body` = `ExtractBlock(content)`.
3. Return concatenation of `head` and `body`.

**ConcatenateSubsections(subsections: list of parsing.NodeSubsection) -> string**

1. Let `result` = empty string.
2. For each subsection in `subsections`:
   a. Let `block` = `FormatSection(subsection.raw_heading,
      subsection.content)`.
   b. If `result` is not empty and `block` is not empty,
      append `\n` to `result`.
   c. Append `block` to `result`.
3. Return `result`.

### Content hash helpers

**HashPublicSubsections(node: parsing.Node) -> optional raw bytes (20)**

1. If `node.public` is absent, return absent.
2. If `node.public.subsections` is empty, return absent.
3. Let `text` = `ConcatenateSubsections(node.public.subsections)`.
4. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashQualifiedSubsection(node: parsing.Node, qualifier: string) -> optional raw bytes (20)**

1. Let `normalized_qualifier` = `parsing.NormalizeText(qualifier)`.
2. Find the subsection in `node.public.subsections`
   whose `heading` equals `normalized_qualifier`.
3. If not found, return absent.
4. Let `text` = `FormatSection(subsection.raw_heading,
   subsection.content)`.
5. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashAgentSection(node: parsing.Node) -> optional raw bytes (20)**

1. If `node.agent` is absent, return absent.
2. If `node.agent.content` is empty (after blank-line
   removal) and `node.agent.subsections` is empty,
   return absent.
3. Let `text` = `ExtractBlock(node.agent.content)`.
4. For each subsection in `node.agent.subsections`:
   a. Let `sub_block` = `FormatSection(
      subsection.raw_heading, subsection.content)`.
   b. If `text` is not empty and `sub_block` is not
      empty, append `\n` to `text`.
   c. Append `sub_block` to `text`.
5. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

**HashFileContent(file_path: oslayer.CfsPath) -> raw bytes (20)**

1. Call `oslayer.OpenFile(file_path, mode="read",
   timeout_ms=30000)`. If `oslayer.OpenFile` raises
   `oslayer.ErrFileUnreadable`, propagate the error.
2. Let `lines` = empty list.
3. Loop:
   a. Call `handle.ReadLine()`.
   b. If `oslayer.ErrEndOfFile` is raised, exit loop.
   c. Append the line followed by `"\n"` to `lines`.
4. Call `handle.Close()`.
   (Call `handle.Close()` in error paths too before
   re-raising.)
5. Let `text` = concatenation of all strings in `lines`.
6. Compute SHA-1 of `text` (UTF-8 bytes). Return the
   raw 20 bytes.

### Main function

**ChainHashCompute(chain: chainresolver.Chain) -> string**

**Step 1 — Collect content hashes**

Let `hashes` = empty list of raw byte sequences
(each 20 bytes).

1. For each `ancestor` in `chain.Ancestors` (from root
   to target's parent):
   a. Call `parsing.ParseNode(ancestor.LogicalName)`.
      If it fails, raise ErrParseFailure.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, append `h` to `hashes`.

2. For each `dep` in `chain.Dependencies` (already
   sorted alphabetically by logical name):
   a. If dep.LogicalName starts with "ARTIFACT/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(dep.Path))`.
      Append `h` to `hashes`.
   b. Else if dep.LogicalName starts with "EXTERNAL/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(dep.Path))`.
      Append `h` to `hashes`.
   c. Else if dep.LogicalName starts with "SPEC/":
      Call `parsing.ParseNode(dep.LogicalName)`.
      If it fails, raise ErrParseFailure.
      If `dep.Qualifier` is nil:
        Let `h` = `HashPublicSubsections(node)`.
        If `h` is present, append `h` to `hashes`.
      If `dep.Qualifier` is not nil:
        Let `h` = `HashQualifiedSubsection(node,
        *dep.Qualifier)`.
        If `h` is present, append `h` to `hashes`.

3. Target `# Public`:
   a. Call `parsing.ParseNode(chain.Target.LogicalName)`.
      If it fails, raise ErrParseFailure.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, append `h` to `hashes`.
   d. Save this `node` result as `target_node`.

4. Target `# Agent`:
   a. Let `h` = `HashAgentSection(target_node)`.
   b. If `h` is present, append `h` to `hashes`.

5. If `chain.Input` is not nil:
   a. Append a single byte `0x49` (`I`) to
      `concatenated` (see step 2) as a marker before
      the input content hash. In practice: append a
      one-byte slice containing `0x49` to `hashes`.
   b. Let `input` = `chain.Input`.
   c. If input.LogicalName starts with "ARTIFACT/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(input.Path))`.
      Append `h` to `hashes`.
   d. Else if input.LogicalName starts with "EXTERNAL/":
      Let `h` = `HashFileContent(
      oslayer.CfsPath(input.Path))`.
      Append `h` to `hashes`.
   e. Else if input.LogicalName starts with "SPEC/":
      Call `parsing.ParseNode(input.LogicalName)`.
      If it fails, raise ErrParseFailure.
      If `input.Qualifier` is nil:
        Let `h` = `HashPublicSubsections(node)`.
        If `h` is present, append `h` to `hashes`.
      If `input.Qualifier` is not nil:
        Let `h` = `HashQualifiedSubsection(node,
        *input.Qualifier)`.
        If `h` is present, append `h` to `hashes`.

**Step 2 — Compute final hash**

1. Let `concatenated` = concatenation of all byte
   sequences in `hashes` in order. Each entry is
   20 bytes except the `0x49` marker which is 1 byte.
2. Compute SHA-1 of `concatenated`.
3. Encode the resulting 20 bytes as base64url (RFC 4648
   §5, no padding) — producing 27 characters.
4. Return the 27-character string.

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
