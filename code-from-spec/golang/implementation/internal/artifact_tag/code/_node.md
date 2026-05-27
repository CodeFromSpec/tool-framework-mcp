---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
input: ARTIFACT/functional/logic/utils/artifact_tag(artifact_tag)
outputs:
  - id: artifacttag
    path: internal/artifacttag/artifacttag.go
---

# ROOT/golang/implementation/internal/artifact_tag/code

Generates the artifacttag package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use the `filereader` package to open and read the file
  line by line.
- Error sentinels with `errors.New`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
- Scan for the `code-from-spec: ` substring in each line.
  Stop reading as soon as a match is found.
- Parse the tag by finding the first `@` after the prefix.
- The hash is exactly the first 27 characters after `@`.
  Anything after those 27 characters is ignored.
