---
depends_on:
  - SPEC/functional/logic/chain/resolver(interface)
  - SPEC/functional/logic/os/file
  - SPEC/functional/logic/os/path_utils(interface)
  - SPEC/functional/logic/parsing/frontmatter(interface)
  - SPEC/functional/logic/parsing/node_parsing
  - SPEC/functional/logic/utils/logical_names(interface)
  - SPEC/functional/logic/utils/text_normalization(interface)
  - EXTERNAL/code-from-spec/_rules/CHAIN_HASH.md
output: code-from-spec/functional/logic/chain/hash/output.md
---

# SPEC/functional/logic/chain/hash

Computes the chain hash for a resolved chain by reading
all chain positions from disk and hashing their content.

# Public

## Namespace

    namespace: chainhash

## Interface

```
function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
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
use `file` directly.

### Block extraction

Spec node content (sections and subsections) is
boundary-normalized before hashing. The same extracted
form is used for both hashing and chain delivery — they
never diverge.

**Single block extraction:**
Given a block's `content` (list of lines from NodeParse):
1. Remove leading blank lines (lines that are empty or
   contain only whitespace U+0020 and U+0009).
2. Remove trailing blank lines.
3. Join the remaining lines with `\n`, then append
   exactly one `\n` at the end.
4. If all lines were blank (nothing remains), the
   extracted content is empty.

**Heading line:** the `raw_heading` with trailing
whitespace removed, followed by `\n`.

**Concatenating multiple blocks** (e.g. `##` subsections
of `# Public`): for each block, emit its heading line
then its extracted content. Separate consecutive blocks
with exactly one blank line (a single `\n` between the
end of one block's content and the next block's heading
line).

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
| Ancestor | `##` subsections of `# Public`, extracted and concatenated in order |
| Target `# Public` | `##` subsections of `# Public`, extracted and concatenated in order |
| Target `# Agent` | `# Agent` section, extracted (heading, content, subsections) |
| `SPEC/` dep, no qualifier | `##` subsections of `# Public` of the referenced node, extracted and concatenated in order |
| `SPEC/` dep, with qualifier | `## <qualifier>` subsection of `# Public`, extracted |
| `ARTIFACT/` dep | Full file content, artifact tag hash neutralized |
| `EXTERNAL/` dep | Full file content |
| `ARTIFACT/` input | Full file content, artifact tag hash neutralized |
| `EXTERNAL/` input | Full file content |

### Hashing spec nodes with NodeParse

For positions that reference spec nodes (ancestors,
target, SPEC/ dependencies), call `NodeParse` with the
logical name from the `ChainItem`. Then extract the
relevant content:

**Hashing `# Public`** (ancestors, target, SPEC/ deps
without qualifier):

Collect only the `##` subsections, in document order.
Apply block extraction and concatenation: for each
subsection, emit its heading line (raw_heading with
trailing whitespace removed + `\n`), then its extracted
content. Separate consecutive subsections with exactly
one blank line. Compute SHA-1 of the result.

The `# Public` heading itself is not included. Content
directly under `# Public` (before the first `##`
subsection) is not included.

If the section is absent or has no subsections, skip
(no hash contributed).

**Hashing `# Agent`** (target only):

Apply block extraction to the entire `# Agent` section.
Emit the section's heading line (raw_heading with
trailing whitespace removed + `\n`), then its extracted
content. Then for each subsection: separate with one
blank line, emit the subsection's heading line, then its
extracted content. Compute SHA-1 of the result.

If the section is absent or has no content and no
subsections, skip (no hash contributed).

**Hashing a subsection** (## qualifier):

Find the subsection within `node.public` whose `heading`
matches `NormalizeText(dep.qualifier)`. Both are
normalized, so the comparison is a simple string match.
Emit the subsection's heading line (raw_heading with
trailing whitespace removed + `\n`), then its extracted
content. Compute SHA-1.

If not found, skip.

### Hashing artifact files

For `ARTIFACT/` dependencies and `input`: open the file
at `file_path` with `FileOpen` (timeout 30000). If `FileOpen` fails,
raise "file unreadable". Read all lines.

Before computing the hash, neutralize the artifact tag:
for any line matching the pattern
`code-from-spec: <anything>@<27 base64url chars>`,
replace the 27-character hash with 27 hyphens
(`---------------------------`). The rest of the line
(including the logical name) is preserved. This prevents
staleness cascading when an upstream artifact is
regenerated but its meaningful content has not changed.

Append `\n` after each line. Compute SHA-1. Call
`FileClose`.

Call `FileClose` in all cases — including error paths.

### Hashing external files

For `EXTERNAL/` dependencies and `EXTERNAL/` input:
open the file at `file_path` with `FileOpen` (timeout 30000). If
`FileOpen` fails, raise "file unreadable". Read all
lines with `FileReadLine`, append `\n` after each line.
Compute SHA-1. Call `FileClose`.

### Algorithm

**Step 1 — Collect content hashes**

Process positions in chain assembly order, computing
SHA-1 for each:

1. For each ancestor in `chain.ancestors`: call
   `NodeParse` with `ancestor.unqualified_logical_name`. If it
   fails, raise "parse failure". Hash `# Public`
   (subsections only, block-extracted). If absent or
   no subsections, skip.

2. For each dependency in `chain.dependencies`:
   - If `LogicalNameIsArtifact(dep.unqualified_logical_name)`:
     hash the artifact file.
   - If `LogicalNameIsExternal(dep.unqualified_logical_name)`:
     hash the external file.
   - If `LogicalNameIsSpec(dep.unqualified_logical_name)` and
     `dep.qualifier` is absent: call `NodeParse`
     with `dep.unqualified_logical_name`. If it fails, raise
     "parse failure". Hash `# Public` (subsections
     only, block-extracted).
   - If `LogicalNameIsSpec(dep.unqualified_logical_name)` and
     `dep.qualifier` is present: call `NodeParse`
     with `dep.unqualified_logical_name`. If it fails, raise
     "parse failure". Hash the
     `## <dep.qualifier>` subsection within `# Public`
     (block-extracted).

3. Target `# Public`: call `NodeParse` with
   `chain.target.unqualified_logical_name`. If it fails, raise
   "parse failure". Hash `# Public` (subsections
   only, block-extracted). If absent or no subsections,
   skip.

4. Target `# Agent`: from the same `NodeParse` result,
   hash the `# Agent` section (block-extracted). If
   absent, skip.

5. Input (if `chain.input` is present):
   - If `LogicalNameIsArtifact(input.unqualified_logical_name)`:
     hash the artifact file.
   - If `LogicalNameIsExternal(input.unqualified_logical_name)`:
     hash the external file.

**Step 2 — Compute final hash**

Concatenate all content hashes from Step 1 as raw bytes
(20 bytes each), in the order computed. Compute SHA-1 of
the concatenation. Encode the 20-byte result as base64url
(RFC 4648 §5, no padding) — 27 characters.

## Contracts

- Spec node content is read via `NodeParse` — headings,
  subsections, and code blocks are handled correctly.
- Spec node content is boundary-normalized (block
  extraction) before hashing. The extracted form is
  exactly what is delivered in the chain — hash and
  delivery never diverge.
- Artifact and external file content is read via
  `file` directly.
- The only normalization before hashing is CRLF → LF
  (handled by `FileReadLine` and `NodeParse`) and block
  boundary normalization for spec nodes.
- Deterministic: same files on disk always produce the
  same hash.
- Positions with missing or empty content contribute no
  hash to the concatenation (they are skipped, not
  hashed as empty).
