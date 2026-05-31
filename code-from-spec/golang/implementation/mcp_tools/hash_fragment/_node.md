---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/hash_fragment(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/logic/mcp_tools/hash_fragment(hash_fragment)
outputs:
  - id: mcphashfragment
    path: internal/mcphashfragment/mcphashfragment.go
---

# ROOT/golang/implementation/mcp_tools/hash_fragment

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `filereader` package for `FileOpen`,
  `FileReadLine`, `FileSkipLines`, `FileClose`.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- For SHA-1 and base64url, use `crypto/sha1` and
  `encoding/base64` (base64.RawURLEncoding).
- The package name should be `mcphashfragment`.
- The function receives plain strings from the MCP
  transport layer. Construct `PathCfs` internally.
- Append `\n` after each line including the last before
  hashing.
