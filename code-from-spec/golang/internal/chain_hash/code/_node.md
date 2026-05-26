---
depends_on:
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/logical_names
input: ARTIFACT/functional/utils/chain_hash(chain_hash)
external:
  - path: CHAIN_HASH.md
outputs:
  - id: chainhash
    path: internal/chainhash/chainhash.go
---

# ROOT/golang/internal/chain_hash/code

Generates the chainhash package implementation.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `os.ReadFile` to read files raw.
- Use `strings.ReplaceAll(content, "\r\n", "\n")` for CRLF normalization.
- Use `crypto/sha1` for hashing.
- Use `encoding/base64` with `base64.RawURLEncoding` for encoding.
- Extract raw sections by scanning for `# Public`, `# Agent`
  headings in the raw text — do not use `parsenode`.
- Use `logicalnames` for path resolution and parent navigation.
- Use `frontmatter` for reading depends_on, external, input, outputs.
