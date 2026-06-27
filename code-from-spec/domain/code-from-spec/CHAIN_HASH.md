# Chain Hash

How the chain hash is computed for artifact staleness
detection. This level of detail is primarily relevant
for tool implementors.

This document assumes familiarity with
CODE_FROM_SPEC.md.

---

## Algorithm

SHA-1, represented as base64url (RFC 4648 §5, no padding).
The output is 27 characters.

---

## Normalization

All text content is normalized before hashing: CRLF line endings
are converted to LF. If the file does not end with LF, a
trailing LF is added.

Spec node content (sections and subsections) is extracted and
boundary-normalized as defined in FILE_FORMAT.md ("Block
extraction"). The extracted form is what is hashed, and it is
exactly the content delivered in the spec chain — hash
and delivery never diverge.

For whole-file content — external files (`EXTERNAL/`
references) and artifact files (`ARTIFACT/` references via
`depends_on` or `input`) — no other normalization is applied.

---

## Content hash

Each position in the spec chain contributes a **content
hash** — the SHA-1 of the content that position injects
into the spec chain. Content is boundary-normalized as
defined in FILE_FORMAT.md ("Block extraction").

| Position | Content hashed |
|---|---|
| Ancestor | `##` subsections of `# Public`, concatenated in order |
| Target node `# Public` | `##` subsections of `# Public`, concatenated in order |
| Target node `# Agent` | Content of `# Agent` (heading not included) |
| `depends_on: ARTIFACT/x` | Full content of the referenced artifact |
| `depends_on: EXTERNAL/x` | Full content of the referenced file |
| `depends_on: SPEC/x` | `##` subsections of `# Public` of the referenced node, concatenated in order |
| `depends_on: SPEC/x(y)` | `## y` subsection of `# Public` of the referenced node |
| `input: ARTIFACT/x` | Full content of the artifact file |
| `input: EXTERNAL/x` | Full content of the referenced file |
| `input: SPEC/x` | `##` subsections of `# Public` of the referenced node, concatenated in order |
| `input: SPEC/x(y)` | `## y` subsection of `# Public` of the referenced node |

---

## Chain hash

The **chain hash** is the SHA-1 of the concatenation of
all content hashes (as raw bytes, not encoded) in chain
assembly order:

1. Each ancestor from root to the target node's parent — content
   hash of `##` subsections of `# Public`, concatenated in
   document order.
2. `depends_on` entries — content hash of each, in alphabetical
   order by the full logical name (including prefix and
   qualifier). Qualifiers are normalized before comparison
   using the heading normalization rules defined in
   FILE_FORMAT.md.
3. The target node — content hash of `# Public`, then content hash
   of `# Agent`.
4. `input` entry (if present) — the byte `0x49` (`I`), followed
   by the content hash of the referenced content.

Redundant `depends_on` entries are deduplicated before hashing.
When an entry without a qualifier exists for a given path, entries
with qualifiers for the same path are removed (the full
`# Public` section already includes every subsection). Exact
duplicates (same path, same qualifier) are also removed. Each
remaining entry contributes its content hash in alphabetical
order by the full logical name (including prefix and
qualifier).

The `0x49` marker ensures that moving a reference from
`depends_on` to `input` (or vice versa) always changes the
chain hash, even when the target node has no `# Public` or
`# Agent` section and the content hash is the same in both
positions.

The resulting SHA-1 is encoded as base64url to produce the
27-character chain hash recorded in the manifest.

---

## Ordering example

Generating the artifact for `SPEC/payments/transfers`.
A node with mixed dependencies:

```yaml
---
depends_on:
  - SPEC/architecture/backend/config(interface)
  - EXTERNAL/proto/payments/v1/transfers.proto
  - ARTIFACT/extraction/email-templates
  - SPEC/integrations/database
  - EXTERNAL/docs/vendor/api-spec.yaml
  - ARTIFACT/extraction/proto
input: ARTIFACT/functional/transfers
output: internal/transfers/handler.go
---
```

The resulting spec chain order:

```
SPEC/payments                               [# Public]      → A  (ancestor)
ARTIFACT/extraction/email-templates         [full]           → B  (depends_on)
ARTIFACT/extraction/proto                   [full]           → C  (depends_on)
EXTERNAL/docs/vendor/api-spec.yaml          [full]           → D  (depends_on)
EXTERNAL/proto/payments/v1/transfers.proto  [full]           → E  (depends_on)
SPEC/architecture/backend/config(interface) [## Interface]   → F  (depends_on)
SPEC/integrations/database                  [# Public]       → G  (depends_on)
SPEC/payments/transfers                     [# Public]       → H  (target node)
SPEC/payments/transfers                     [# Agent]        → I  (target node)
                                                               0x49
ARTIFACT/functional/transfers               [full]           → J  (input)
```

The `depends_on` entries are sorted alphabetically by
the full logical name — `ARTIFACT/` before `EXTERNAL/`
before `SPEC/` — regardless of the order in the
frontmatter.

---

## Hash examples

### With input

Given the spec chain for `SPEC/payments/fees/calculation`:

```
SPEC/payments                              [# Public]      → content hash A  (ancestor)
SPEC/payments/fees                         [# Public]      → content hash B  (ancestor)
EXTERNAL/proto/payments/v1/transfers.proto [full]          → content hash C  (depends_on)
SPEC/integrations/database                 [# Public]      → content hash D  (depends_on)
SPEC/payments/fees/calculation             [# Public]      → content hash E  (target node)
SPEC/payments/fees/calculation             [# Agent]       → content hash F  (target node)
ARTIFACT/functional/calc                   [full]          → content hash G  (input)
```

The chain hash is:

```
SHA-1( A || B || C || D || E || F || 0x49 || G )
```

where `||` denotes concatenation of raw hash bytes (20 bytes
each), and the result is encoded as base64url.

### Without input

Given the spec chain for `SPEC/payments/fees/rounding`:

```
SPEC/payments                              [# Public]      → content hash A  (ancestor)
SPEC/payments/fees                         [# Public]      → content hash B  (ancestor)
SPEC/payments/fees/rounding                [# Public]      → content hash C  (target node)
SPEC/payments/fees/rounding                [# Agent]       → content hash D  (target node)
```

The chain hash is:

```
SHA-1( A || B || C || D )
```

No `0x49` marker — the input position is absent.

---

## Resources

| Document | Description |
|---|---|
| [CODE_FROM_SPEC.md](https://github.com/CodeFromSpec/framework/blob/main/CODE_FROM_SPEC.md) | Full methodology specification |
| [FILE_FORMAT.md](https://github.com/CodeFromSpec/framework/blob/main/rules/FILE_FORMAT.md) | Block extraction and normalization rules |
| [CHAIN_ASSEMBLY.md](https://github.com/CodeFromSpec/framework/blob/main/rules/CHAIN_ASSEMBLY.md) | Chain format, assembly order, and delivery |
| [MANIFEST.md](https://github.com/CodeFromSpec/framework/blob/main/rules/MANIFEST.md) | Manifest format and artifact status |
