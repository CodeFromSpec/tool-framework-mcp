---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/artifact_tag
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/parsing/artifact_tag
output: internal/artifacttag/artifacttag.go
---

# ROOT/golang/implementation/parsing/artifact_tag

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
