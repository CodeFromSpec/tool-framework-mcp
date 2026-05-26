<!-- code-from-spec: ROOT/functional/utils/frontmatter@DuCaQuf5LmW3zCuwC_HVmJAMZWU -->

# ParseFrontmatter

Reads the optional YAML frontmatter block at the top of a file and
returns a structured record of its contents. The parser stops as soon
as the closing `---` delimiter is found and never reads the file body.

---

## Data structures

```
record ExternalFragment
  description: optional string   -- human-readable note for this fragment
  lines: string                  -- line-range selector (e.g. "1-10")
  hash: string                   -- content hash of the selected lines

record External
  path: string                           -- path to the external file
  fragments: optional list of ExternalFragment

record Output
  id: string    -- logical identifier for the output artifact
  path: string  -- relative file path where the artifact is written

record Frontmatter
  depends_on: list of strings   -- logical names this node depends on
  external: list of External    -- external file references
  input: string                 -- logical name of the input artifact
  outputs: list of Output       -- artifacts this node produces
```

---

## Functions

### ParseFrontmatter(file_path) -> Frontmatter

Parses the frontmatter block from the file at `file_path` and returns
a `Frontmatter` record. All fields default to empty (empty list or
empty string) when absent from the YAML.

**Parameters**
- `file_path` — string: path to the file to parse.

**Return value**
- A `Frontmatter` record.

**Errors**
- `"file unreadable"` — the file cannot be opened or read.
- `"malformed YAML"` — the content between `---` delimiters is not
  valid YAML.

**Steps**

1. Open the file at `file_path` using `OpenFileReader`.
   If the file cannot be opened, raise error `"file unreadable"`.

2. Read the first line using `ReadLine`.
   If reading raises "end of file", return an empty `Frontmatter`
   record (all fields empty).
   If the first line is not exactly `"---"`, return an empty
   `Frontmatter` record.
   -- A missing or absent opening delimiter is not an error.

3. Collect lines into a buffer until one of the following:
   a. A line is exactly `"---"` — this is the closing delimiter.
      Stop collecting. Do NOT include this line in the buffer.
   b. `ReadLine` raises "end of file" — the file ended without a
      closing delimiter. Raise error `"malformed YAML"`.

4. Join the collected buffer lines with newline characters to form
   a single YAML string.
   If the buffer is empty (i.e., the block was `---\n---`), proceed
   to step 6 with an empty parsed map.

5. Parse the YAML string.
   If parsing fails for any reason, raise error `"malformed YAML"`.
   The result is a map of field names to values.

6. Extract recognized fields from the parsed map.
   Ignore any fields not listed below.

   - `depends_on`:
     If present, interpret as a list of strings.
     If absent, use an empty list.

   - `external`:
     If present, interpret as a list of `External` records.
     For each entry:
       - `path`: string (required field within each entry).
       - `fragments`: optional list of `ExternalFragment` records.
         For each fragment entry:
           - `description`: optional string.
           - `lines`: string.
           - `hash`: string.
     If absent, use an empty list.

   - `input`:
     If present, interpret as a string.
     If absent, use an empty string.

   - `outputs`:
     If present, interpret as a list of `Output` records.
     For each entry:
       - `id`: string.
       - `path`: string.
     If absent, use an empty list.

7. Return the populated `Frontmatter` record.
