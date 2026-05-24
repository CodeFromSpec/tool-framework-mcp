---
outputs:
  - id: artifact_tag
    path: code-from-spec/functional/utils/artifact_tag/output.md
---

# ROOT/functional/utils/artifact_tag

Extracts the artifact tag from generated files for
staleness detection.

# Public

## Artifact tag format

Generated files contain the string:

```
code-from-spec: <logical-name>@<hash>
```

The tag may appear inside any comment syntax (`//`, `#`,
`/* */`, `--`, `<!-- -->`). The tool does not parse
comment syntax — it scans each line for the pattern
regardless of context.

## Behavior

### Input

A file path.

### Output

A record with:
- `logical_name` — the node that generated this file.
- `hash` — the chain hash at the time of generation
  (base64url, 27 characters).

## Detection

Read the file line by line from the top. For each line,
look for the substring `code-from-spec: `. Stop reading
as soon as a match is found.

## Extraction

Once a line containing `code-from-spec: ` is found:

1. Take everything after `code-from-spec: ` to the end
   of the line (trimming trailing whitespace).
2. Find the last occurrence of `@`.
3. The logical name is everything before `@`.
4. The hash is everything after `@`.
5. Validate: logical name must not be empty, hash must
   be exactly 27 characters.

## Error conditions

| Condition | Description |
|---|---|
| File unreadable | The file cannot be opened or read. |
| No tag found | The file has no `code-from-spec: ` substring. |
| Malformed tag | The tag exists but cannot be parsed (no `@`, empty name, wrong hash length). |
