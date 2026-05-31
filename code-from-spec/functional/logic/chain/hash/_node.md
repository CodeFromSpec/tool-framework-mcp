---
depends_on:
  - ROOT/functional/logic/chain/resolver(interface)
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/os/path_utils(interface)
  - ROOT/functional/logic/parsing/frontmatter(interface)
  - ROOT/functional/logic/parsing/node_parsing
  - ROOT/functional/logic/utils/logical_names(interface)
  - ROOT/functional/logic/utils/text_normalization(interface)
external:
  - path: CHAIN_HASH.md
outputs:
  - id: chain_hash
    path: code-from-spec/functional/logic/chain/hash/output.md
---

# ROOT/functional/logic/chain/hash

Computes the chain hash for a resolved chain by reading
all chain positions from disk and hashing their content.

# Public

## Interface

```
function ChainHashCompute(chain: Chain) -> string
  errors:
    - FileUnreadable: a file in the chain cannot be
      read or opened.
    - ParseFailure: a node file cannot be parsed.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.
```

Receives a `Chain` (as returned by `ChainResolve`) and
returns a 27-character base64url encoded SHA-1 hash.

The function reads each position's content from disk,
computes a content hash (SHA-1) for each, concatenates
all content hashes as raw bytes in chain assembly order,
and computes the final SHA-1 of the concatenation.

# Agent

## Behavior

Given a resolved `Chain`, compute the chain hash. For
spec nodes (`_node.md` files), use `NodeParse` to
extract sections. For artifact files and external files,
use `file_reader` directly.

### Hashing convention

When reconstructing content for hashing from lines,
append `\n` (LF) after every line, including the last.
This ensures a deterministic representation regardless
of whether the original file had a trailing newline.

Heading matching is inherited from `NodeParse`: section
classification (public, agent) and subsection lookup
use the normalized `heading` field (case-folded). The
`raw_heading` and `content` used for hashing preserve
the original text — case is preserved in the hash input.

### Content hash per position

Each position contributes a content hash — the SHA-1 of
the content that position injects into the chain.

| Position | Content hashed |
|---|---|
| Ancestor | `# Public` section |
| Target | `# Public` section, then `# Agent` section |
| `ROOT/` dep, no qualifier | `# Public` section of the referenced node |
| `ROOT/` dep, with qualifier | `## <qualifier>` subsection of `# Public` |
| `ARTIFACT/` dep | Full file content, excluding frontmatter |
| External, no fragments | Full file content |
| External, with fragments | Each fragment's line range, concatenated in declaration order |
| Input | Full file content, excluding frontmatter |

### Hashing spec nodes with NodeParse

For positions that reference spec nodes (ancestors,
target, ROOT/ dependencies), call `NodeParse` with the
logical name from the `ChainItem`. Then extract the
relevant content:

**Hashing a full section** (# Public or # Agent):

Collect the section's `raw_heading` line, then all lines
in `content`, then for each subsection: the subsection's
`raw_heading` line followed by all lines in its
`content`. Append `\n` after each line (per hashing
convention). Compute SHA-1 of the result.

If the section is absent or has no content and no
subsections, skip (no hash contributed).

**Hashing a subsection** (## qualifier):

Find the subsection within `node.public` whose `heading`
matches `NormalizeText(dep.qualifier)`. Both are
normalized, so the comparison is a simple string match.
Collect the
subsection's `raw_heading` line, then all lines in its
`content`. Append `\n` after each line. Compute SHA-1.

If not found, skip.

### Hashing artifact files

For `ARTIFACT/` dependencies and `input`: open the file
at `file_path` with `FileOpen`. If `FileOpen` fails,
raise "file unreadable". Check if the first line is
exactly `---`. If so, read lines until the next `---`
line (closing delimiter), discarding the frontmatter.
Read the remaining lines. Append `\n` after each line.
Compute SHA-1. Call `FileClose`.

If the first line is not `---`, read all lines, append
`\n` after each, compute SHA-1. Call `FileClose`.

Call `FileClose` in all cases — including error paths.

### Hashing external files

For external entries without fragments: create a `PathCfs`
from the entry's `path` string, open with `FileOpen`.
If `FileOpen` fails, raise "file unreadable". Read all
lines with `FileReadLine`, append `\n` after each line.
Compute SHA-1. Call `FileClose`.

For external entries with fragments, create a `PathCfs`
from the entry's `path` string. For each fragment:
- Parse the `lines` field as `start-end` (1-based,
  inclusive).
- Open the file with `FileOpen`. If `FileOpen` fails,
  raise "file unreadable".
- Use `FileSkipLines` to skip `start - 1` lines, then
  read `end - start + 1` lines with `FileReadLine`.
- Call `FileClose`.
- Append `\n` after each line.

Concatenate all fragment contents in declaration order.
Compute a single SHA-1 for the concatenation.

### Algorithm

**Step 1 — Collect content hashes**

Process positions in chain assembly order, computing
SHA-1 for each:

1. For each ancestor in `chain.ancestors`: call
   `NodeParse` with `ancestor.logical_name`. If it
   fails, raise "parse failure". Hash the `# Public`
   section (full section). If absent or empty, skip.

2. For each dependency in `chain.dependencies`:
   - If `LogicalNameIsArtifact(dep.logical_name)`:
     hash the artifact file (frontmatter stripped).
   - Else if `dep.qualifier` is absent: call `NodeParse`
     with `dep.logical_name`. If it fails, raise
     "parse failure". Hash the `# Public` section
     (full section).
   - Else: call `NodeParse` with `dep.logical_name`.
     If it fails, raise "parse failure". Hash the
     `## <dep.qualifier>` subsection within `# Public`.

3. For each external in `chain.external`: hash per
   the external rules above.

4. Target `# Public`: call `NodeParse` with
   `chain.target.logical_name`. If it fails, raise
   "parse failure". Hash the `# Public` section
   (full section).

5. Target `# Agent`: from the same `NodeParse` result,
   hash the `# Agent` section. If absent, skip.

6. Input (if `chain.input` is present): hash the
   artifact file (frontmatter stripped).

**Step 2 — Compute final hash**

Concatenate all content hashes from Step 1 as raw bytes
(20 bytes each), in the order computed. Compute SHA-1 of
the concatenation. Encode the 20-byte result as base64url
(RFC 4648 §5, no padding) — 27 characters.

## Contracts

- Spec node content is read via `NodeParse` — headings,
  subsections, and code blocks are handled correctly.
- Artifact and external file content is read via
  `file_reader` directly.
- The only normalization before hashing is CRLF → LF
  (handled by `FileReadLine` and `NodeParse`).
- Deterministic: same files on disk always produce the
  same hash.
- Positions with missing or empty content contribute no
  hash to the concatenation (they are skipped, not
  hashed as empty).
