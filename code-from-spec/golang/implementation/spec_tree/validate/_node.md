---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/list_files
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/utils/text_normalization
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
input: ARTIFACT/functional/logic/spec_tree/validate
output: internal/spectreevalidate/spectreevalidate.go
---

# SPEC/golang/implementation/spec_tree/validate

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `filereader` package for `FileOpen`, `FileReadLine`,
  `FileSkipLines`, `FileClose`.
- Use the `pathutils` package for `PathValidateCfs` and `PathCfs`.
- Use the `textnormalization` package for `NormalizeText`.
- Use the `frontmatter` package for the `Frontmatter` record.
- Use the `parsenode` package for the `Node` record.
- The package name should be `spectreevalidate`.
- `SpecTreeValidateInput` and `FormatError` are exported structs
  in this package.
- The function never returns an error — all problems are collected
  as FormatError entries in the returned list.
- For SHA-1 and base64url, use `crypto/sha1` and
  `encoding/base64` (base64.RawURLEncoding).
