---
depends_on:
  - ROOT/functional/utils/file_reader
outputs:
  - id: artifact_tag
    path: artifacts/functional/utils/parsing/artifact_tag/output.md
---

# ROOT/functional/utils/parsing/artifact_tag

Extracts the artifact tag from generated files for
staleness detection.

# Public

## Interface

```
record ArtifactTag
  logical_name: string
  hash: string

function ExtractArtifactTag(file_path) -> ArtifactTag
  errors:
    - file unreadable: the file cannot be opened or read.
    - no tag found: the file has no code-from-spec: substring.
    - malformed tag: the tag exists but cannot be parsed (no @, empty name, wrong hash length).
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

Read the file line by line from the top. For each line,
look for the substring `code-from-spec: `. Stop reading
as soon as a match is found.

### Extraction

Once a line containing `code-from-spec: ` is found:

1. Take everything after `code-from-spec: ` to the end
   of the line (trimming trailing whitespace).
2. Find the last occurrence of `@`.
3. The logical name is everything before `@`.
4. The hash is everything after `@`.
5. Validate: logical name must not be empty, hash must
   be exactly 27 characters.

## Contracts

- Reads the file only until the first match — does not
  read the entire file.
- The hash is always exactly 27 characters (base64url).
