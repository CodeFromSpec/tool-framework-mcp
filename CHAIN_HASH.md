# Chain Hash

How the chain hash is computed for artifact staleness detection.
This document assumes familiarity with
[CODE_FROM_SPEC.md](CODE_FROM_SPEC.md).

This level of detail is primarily relevant for tool implementors.

---

## Algorithm

SHA-1, represented as base64url (RFC 4648 §5, no padding).
The output is 27 characters.

---

## Normalization

All text content is normalized before hashing: CRLF line endings
are converted to LF. If the file does not end with LF, a
trailing LF is added. No other normalization is applied.

This applies to spec node content, external file content, and
artifact file content (referenced via `depends_on` or `input`).

---

## Content hash

Each position in the chain contributes a **content hash** — the
SHA-1 of the content that position injects into the chain. The
heading itself (e.g. `# Public`, `## Interface`) is part of the
hashed content.

| Position | Content hashed |
|---|---|
| Ancestor | `# Public` section |
| Target `# Public` | `# Public` section |
| Target `# Agent` | `# Agent` section |
| `depends_on: ROOT/x/y` | `# Public` section of the referenced node |
| `depends_on: ROOT/x/y(z)` | `## z` subsection of `# Public` of the referenced node |
| `depends_on: ARTIFACT/x/y(id)` | Full content of the referenced artifact, excluding any frontmatter |
| `external` (whole file) | Full content of the referenced file |
| `external` (with fragments) | Concatenation of each fragment's content, in declaration order |
| `input: ARTIFACT/x/y(id)` | Full content of the artifact file, excluding any frontmatter |

---

## Chain hash

The chain hash is the SHA-1 of the concatenation of all content
hashes (as raw bytes, not encoded) in chain assembly order:

1. Each ancestor from root to the target's parent — `# Public`
   content hash of each.
2. `depends_on` entries — content hash of each, in alphabetical
   order by path.
3. `external` entries — content hash of each, in alphabetical
   order by path.
4. The target — content hash of `# Public`, then content hash
   of `# Agent`.
5. `input` entry (if present) — content hash of the artifact file.

Redundant `depends_on` entries are deduplicated before hashing.
When an entry without a qualifier exists for a given path, entries
with qualifiers for the same path are removed (the full
`# Public` section already includes every subsection). Exact
duplicates (same path, same qualifier) are also removed. Each
remaining entry contributes its content hash in alphabetical
order by path.

The resulting SHA-1 is encoded as base64url to produce the 27
character string that appears in the artifact tag:

```
code-from-spec: ROOT/payments/fees/calculation@k4Xz9pQ1rLmN3vB7wY2tHsJ8dFa
```

---

## Example

Given the chain for `ROOT/payments/fees/calculation`:

```
ROOT                           [# Public]            → content hash A
ROOT/payments                  [# Public]            → content hash B
ROOT/payments/fees             [# Public]            → content hash C
ROOT/external/database         [# Public]            → content hash D  (depends_on)
proto/payments/v1/transfers.proto [full]             → content hash E  (external)
ROOT/payments/fees/calculation [# Public]            → content hash F  (target)
ROOT/payments/fees/calculation [# Agent]             → content hash G  (target)
ARTIFACT/functional/calc(calc) [file content]        → content hash H  (input)
```

The chain hash is:

```
SHA-1( A || B || C || D || E || F || G || H )
```

where `||` denotes concatenation of raw hash bytes (20 bytes
each), and the result is encoded as base64url.
