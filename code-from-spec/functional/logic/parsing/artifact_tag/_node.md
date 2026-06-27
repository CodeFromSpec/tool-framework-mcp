---
depends_on:
  - SPEC/functional/logic/os/file
  - SPEC/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/parsing/artifact_tag/output.md
---

# SPEC/functional/logic/parsing/artifact_tag

Extracts the artifact tag from generated files for
staleness detection.

# Public

## Interface

```
record ArtifactTag
  logical_name: string
  hash: string

function ArtifactTagExtract(file_path: pathutils.PathCfs) -> ArtifactTag
  errors:
    - NoTagFound: the file has no code-from-spec:
      substring.
    - MalformedTag: the tag exists but cannot be parsed
      (no @, empty name, wrong hash length).
    - (File.*): propagated from FileOpen, FileReadLine.
```

### Artifact tag format

Generated files contain the string:

```
code-from-spec: <logical-name>@<hash>
```

The tag may appear inside any comment syntax (`//`, `#`,
`/* */`, `--`, `<!-- -->`). The tool does not parse
comment syntax — it scans each line for the pattern
regardless of context.

# Agent

## Behavior

### Detection

Open the file with `FileOpen` (mode `"read"`, timeout 30000). Read line by line
using `ReadLine`. For each line, look for the substring
`code-from-spec: `. Stop reading as soon as a match is
found. Close the reader when done (whether a match was
found or not).

### Extraction

Once a line containing `code-from-spec: ` is found:

1. Take the substring starting immediately after
   `code-from-spec: `.
2. Trim leading whitespace.
3. Find the first occurrence of `@`.
4. The logical name is everything between the trimmed
   start and `@`.
5. The hash is the 27 characters immediately after `@`.
6. Validate: logical name must not be empty, `@` must
   exist, and there must be at least 27 characters
   after `@`.

## Contracts

- Reads the file only until the first match — does not
  read the entire file.
- The hash is always exactly 27 characters (base64url).
