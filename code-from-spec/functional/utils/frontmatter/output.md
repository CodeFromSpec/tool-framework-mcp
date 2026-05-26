<!-- code-from-spec: ROOT/functional/utils/frontmatter@HBZRJbd4xmtRNeilrTuahKd46UA -->

# ParseFrontmatter

Parses the optional YAML frontmatter block from the top of a spec file.
The frontmatter block is delimited by `---` lines. If no frontmatter is
present, an empty record is returned without error. The file body is
never read — parsing stops after the closing `---`.

## Data Structures

```
record ExternalFragment
  description: optional string
  lines: string
  hash: string

record External
  path: string
  fragments: optional list of ExternalFragment

record Output
  id: string
  path: string

record Frontmatter
  depends_on: list of strings   (default: empty list)
  external: list of External    (default: empty list)
  input: string                 (default: empty string)
  outputs: list of Output       (default: empty list)
```

## Functions

---

### ParseFrontmatter(file_path) -> Frontmatter

**Parameters**
- `file_path`: string — path to the spec file to parse

**Returns**
- `Frontmatter` record with fields populated from YAML, or all-empty defaults

**Errors**
- `"file unreadable"`: the file cannot be opened or read
- `"malformed YAML"`: the content between `---` delimiters is not valid YAML

**Steps**

1. Open the file at `file_path` using `OpenFileReader(file_path)`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Read the first line using `ReadLine(reader)`.
   If reading fails with "end of file", return an empty Frontmatter record.
   If the first line is not exactly `"---"`, return an empty Frontmatter record.
   (A missing or non-`---` first line means no frontmatter is present — this is not an error.)

3. Collect lines into a buffer, reading one line at a time using `ReadLine(reader)`,
   until one of the following:
   - The line is exactly `"---"` — stop collecting (do not include this line in the buffer).
   - `ReadLine` raises "end of file" — raise error `"malformed YAML"`,
     because an opening `---` was found but no closing `---` was encountered.

4. Parse the collected buffer as YAML.
   If YAML parsing fails, raise error `"malformed YAML"`.

5. Extract the following known fields from the parsed YAML.
   If a field is absent, use its default value.
   Unknown fields are silently ignored.

   - `depends_on`:
     Expected: list of strings.
     Default: empty list.
     Assign to Frontmatter.depends_on.

   - `external`:
     Expected: list of External records, each with:
       - `path`: string
       - `fragments`: optional list of ExternalFragment records, each with:
           - `description`: optional string
           - `lines`: string
           - `hash`: string
     Default: empty list.
     Assign to Frontmatter.external.

   - `input`:
     Expected: string.
     Default: empty string.
     Assign to Frontmatter.input.

   - `outputs`:
     Expected: list of Output records, each with:
       - `id`: string
       - `path`: string
     Default: empty list.
     Assign to Frontmatter.outputs.

6. Return the populated Frontmatter record.

---

## Contracts and Invariants

- The parser stops reading after the closing `---`. It never reads the file body.
- An empty frontmatter block (`---` immediately followed by `---`) produces
  a Frontmatter record with all fields at their defaults (empty lists, empty string).
- All recognized fields are optional.
- Unknown YAML keys are silently ignored.
- CRLF normalization is handled by `ReadLine` — the parser does not need to
  handle `\r\n` line endings explicitly.
