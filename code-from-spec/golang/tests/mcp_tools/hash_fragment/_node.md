---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/hash_fragment(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
input: ARTIFACT/functional/tests/mcp_tools/hash_fragment(hash_fragment_tests)
outputs:
  - id: mcphashfragment_test
    path: internal/mcphashfragment/mcphashfragment_test.go
---

# ROOT/golang/tests/mcp_tools/hash_fragment

# Agent

## Test setup guidance

`MCPHashFragment` reads files from disk. Tests must use
`testChdir` and create files with known content. When
verifying hash values, compute the expected hash using
SHA-1 of the lines (each with `\n` appended) encoded
as base64url (no padding).
