---
depends_on:
  - ROOT/functional/logic/utils/file_reader
  - ROOT/functional/logic/utils/frontmatter
  - ROOT/functional/logic/utils/logical_names
external:
  - path: CHAIN_HASH.md
outputs:
  - id: chain_hash
    path: code-from-spec/functional/logic/utils/chain_hash/output.md
---

# ROOT/functional/logic/utils/chain_hash

Computes the chain hash for a given target node by reading
all chain positions from disk.

# Public

## Interface

```
function ComputeChainHash(logical_name) -> string
  errors:
    - invalid logical name: cannot resolve the logical name.
    - unreadable file: a file in the chain cannot be read.
```

Returns a 27-character base64url encoded SHA-1 hash.

# Agent

## Behavior

Given a logical name, reads all chain positions from disk
and computes the chain hash. Each position's content is
read raw from the file (not from parsed/reconstructed data).
The only normalization is CRLF → LF. Close each reader
after reading its content.

### Algorithm

**Step 1 — Collect ancestors**

Walk upward from the target using parent navigation.
For each ancestor (root first), read the raw file using
`file_reader`. Extract the raw bytes of the `# Public`
section (from the `# Public` heading to the next `#`
heading or end of file). If `# Public` is absent or
empty, skip. Compute SHA-1 of the raw bytes.

**Step 2 — Collect dependencies**

Read the target's frontmatter. For each `depends_on`
entry, in alphabetical order by logical name:
- `ROOT/x/y` — read the raw file, extract `# Public`
  section raw bytes. Compute SHA-1.
- `ROOT/x/y(z)` — read the raw file, extract `## z`
  subsection raw bytes within `# Public`. Compute SHA-1.
- `ARTIFACT/x/y(id)` — resolve the artifact path, read
  the file, strip frontmatter. Compute SHA-1 of the
  remaining content.

**Step 3 — Collect external files**

For each `external` entry, in alphabetical order by path:
- No fragments: read the entire file. Compute SHA-1.
- With fragments: for each fragment, extract the declared
  line range. Concatenate all fragments in declaration
  order. Compute SHA-1 of the concatenation.

**Step 4 — Target # Public**

Read the target's raw file. Extract `# Public` section
raw bytes. Compute SHA-1.

**Step 5 — Target # Agent**

Extract `# Agent` section raw bytes from the same file.
If absent, skip. Compute SHA-1.

**Step 6 — Input**

If the target has an `input` field, resolve the artifact
path, read the file, strip frontmatter. Compute SHA-1.

**Step 7 — Final hash**

Concatenate all SHA-1 digests (raw 20 bytes each) in the
order above. Compute SHA-1 of the concatenation. Encode
as base64url (RFC 4648 §5, no padding) — 27 characters.

## Contracts

- All content is read raw from disk — never from parsed
  or reconstructed data.
- The only normalization before hashing is CRLF → LF.
- Deterministic: same files on disk always produce the
  same hash.
