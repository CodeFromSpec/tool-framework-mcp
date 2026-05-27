<!-- code-from-spec: ROOT/functional/utils/chain_hash@YRLj-9uV__mZdA8VdWgDJVg0tUo -->

# chain_hash

Computes the chain hash for a spec node. The chain hash is a 27-character
base64url-encoded SHA-1 that captures the full set of inputs for a given
node — ancestors, dependencies, external files, and the target itself.

---

## Dependencies

Uses the following utilities:

- `file_reader` — sequential line-by-line file reading (OpenFileReader,
  ReadLine, Close).
- `frontmatter` — ParseFrontmatter for reading depends_on, external, input,
  and outputs fields.
- `logical_names` — ResolvePath, ResolveArtifactReference, GetParent,
  ExtractQualifier.

---

## Helper: ExtractSection

```
function ExtractSection(file_path, heading) -> string
```

Reads `file_path` and returns the raw (CRLF-normalized to LF) text of the
section that begins with `heading`, up to (but not including) the next line
that starts with `#` at the same or higher level, or end of file.

The heading line itself is included in the returned text.

If the heading is not found in the file, returns an empty string.

Parameters:
- file_path: string — path to the spec node file.
- heading: string — exact heading text, e.g. `"# Public"` or `"## Interface"`.

Steps:

  1. Open a FileReader for file_path.
     If the file cannot be opened, raise error "unreadable file: <file_path>".

  2. Determine heading_level: count the leading `#` characters in heading.

  3. Read lines one at a time, normalizing CRLF to LF, until end of file.
     Keep a flag `inside` (initially false) and a buffer `lines` (initially
     empty).

     For each line:
       If `inside` is false:
         If the line equals heading, set `inside` to true and append the
         line to `lines`.
       Else (`inside` is true):
         If the line starts with one or more `#` characters and the number
         of leading `#` characters is less than or equal to heading_level:
           Stop reading — the section has ended.
         Else:
           Append the line to `lines`.

  4. Close the reader.

  5. Join `lines` with LF and return the result.
     If `lines` is empty, return an empty string.
```

---

## Helper: ExtractSubsection

```
function ExtractSubsection(file_path, parent_heading, sub_heading) -> string
```

Reads `file_path`, first locates the section beginning with `parent_heading`
(a `#`-level heading), and within it locates the subsection beginning with
`sub_heading` (a `##`-level heading). Returns the raw (CRLF-normalized to LF)
text of that subsection.

Used for `depends_on: ROOT/x/y(z)` entries — the qualifier `z` is a `##`
subsection inside `# Public`.

Parameters:
- file_path: string — path to the spec node file.
- parent_heading: string — e.g. `"# Public"`.
- sub_heading: string — e.g. `"## Interface"`.

Steps:

  1. Open a FileReader for file_path.
     If the file cannot be opened, raise error "unreadable file: <file_path>".

  2. Determine parent_level: count leading `#` in parent_heading.
     Determine sub_level: count leading `#` in sub_heading.

  3. Read lines one at a time (CRLF → LF).
     Keep flags `in_parent` (false), `in_sub` (false), and buffer `lines`
     (empty).

     For each line:
       If `in_sub` is true:
         If line starts with `#` and leading `#` count <= sub_level:
           Stop reading.
         Else:
           Append line to `lines`.
       Else if `in_parent` is true:
         If line starts with `#` and leading `#` count <= parent_level:
           Stop reading — left parent without finding subsection.
         Else if line equals sub_heading:
           Set `in_sub` to true, append line to `lines`.
       Else:
         If line equals parent_heading:
           Set `in_parent` to true.

  4. Close the reader.

  5. Join `lines` with LF and return. If `lines` is empty, return "".
```

---

## Helper: ReadFileStripFrontmatter

```
function ReadFileStripFrontmatter(file_path) -> string
```

Reads a file, normalizes CRLF to LF, and removes any leading YAML frontmatter
block (content between the first `---` delimiter pair, inclusive).

Used for `ARTIFACT/` references in `depends_on` and `input`.

Parameters:
- file_path: string — path to an artifact file.

Steps:

  1. Open a FileReader for file_path.
     If the file cannot be opened, raise error "unreadable file: <file_path>".

  2. Read all lines (CRLF → LF) into a buffer `all_lines`.

  3. Close the reader.

  4. Check for frontmatter:
     If the first non-empty line is exactly `---`:
       Find the next line (after the first `---`) that is also exactly `---`.
       If found, discard all lines up to and including that second `---`.
       The remaining lines form the content.
     Else:
       All lines form the content.

  5. Join the content lines with LF and return.
```

---

## Helper: ReadFileRaw

```
function ReadFileRaw(file_path) -> string
```

Reads the entire file, normalizing CRLF to LF. No other processing.

Parameters:
- file_path: string.

Steps:

  1. Open a FileReader for file_path.
     If the file cannot be opened, raise error "unreadable file: <file_path>".

  2. Read all lines (CRLF → LF) into `all_lines`.

  3. Close the reader.

  4. Return all lines joined with LF.
```

---

## Helper: SHA1Digest

```
function SHA1Digest(text) -> bytes (20 bytes)
```

Computes the SHA-1 digest of `text` encoded as UTF-8.
Returns the raw 20-byte digest.

---

## Helper: ResolveArtifactFilePath

```
function ResolveArtifactFilePath(logical_name) -> string
```

Resolves an `ARTIFACT/x/y(id)` logical name to the file path of the
generated artifact.

Parameters:
- logical_name: string — must start with `ARTIFACT/` and have a qualifier.

Steps:

  1. Call ResolveArtifactReference(logical_name) to get
     record { node_path, artifact_id }.
     If it raises "unrecognized prefix" or "missing qualifier",
     propagate as error "invalid logical name: <logical_name>".

  2. Call ParseFrontmatter(node_path) to get the node's frontmatter.
     For each entry in frontmatter.outputs:
       If entry.id equals artifact_id, return entry.path.

  3. If no matching output found, raise error
     "invalid logical name: <logical_name> — artifact id not found".
```

---

## Main Function

```
function ComputeChainHash(logical_name) -> string
  errors:
    - invalid logical name: cannot resolve the logical name.
    - unreadable file: a file in the chain cannot be read.
```

Returns a 27-character base64url-encoded SHA-1 string.

### Steps

**Step 1 — Validate and resolve target**

  1. If logical_name does not start with `"ROOT/"` and is not `"ROOT"`,
     raise error "invalid logical name: <logical_name>".

  2. Call ResolvePath(logical_name) to get target_file_path.
     If it raises an error, propagate as "invalid logical name: <logical_name>".

  3. Read the target's frontmatter:
     Call ParseFrontmatter(target_file_path) -> target_fm.
     If it raises an error, propagate as
     "unreadable file: <target_file_path>".

**Step 2 — Collect ancestor content hashes**

  Initialize `digest_list` as an empty list of byte sequences.

  Build the ancestor chain:
    Start with current = logical_name (stripped of any qualifier).
    Repeatedly call GetParent(current) until "no parent" is raised.
    Collect all returned values in a list — these are the ancestors from
    root down to the target's parent.

  For each ancestor in root-to-parent order:
    a. Call ResolvePath(ancestor) -> ancestor_file_path.
    b. Call ExtractSection(ancestor_file_path, "# Public") -> section_text.
    c. If section_text is not empty:
         Compute SHA1Digest(section_text) and append to digest_list.
    d. If section_text is empty, skip (do not append).

**Step 3 — Collect depends_on content hashes**

  Sort target_fm.depends_on alphabetically by the logical name string.

  For each entry in sorted order:

    Case A — entry starts with `"ROOT/"` and has no qualifier
    (ExtractQualifier returns absent):
      a. Call ResolvePath(entry) -> dep_file_path.
      b. Call ExtractSection(dep_file_path, "# Public") -> section_text.
      c. Compute SHA1Digest(section_text) and append to digest_list.
         (If section_text is empty, SHA1Digest of "" is still appended.)

    Case B — entry starts with `"ROOT/"` and has a qualifier q:
      a. Call ResolvePath(entry) -> dep_file_path.
         (ResolvePath strips the qualifier.)
      b. Call ExtractSubsection(dep_file_path, "# Public", "## <q>")
         -> sub_text.
      c. Compute SHA1Digest(sub_text) and append to digest_list.

    Case C — entry starts with `"ARTIFACT/"`:
      a. Call ResolveArtifactFilePath(entry) -> artifact_file_path.
      b. Call ReadFileStripFrontmatter(artifact_file_path) -> content.
      c. Compute SHA1Digest(content) and append to digest_list.

    If any resolution or read raises an error, propagate it unchanged.

**Step 4 — Collect external content hashes**

  Sort target_fm.external alphabetically by the `path` field.

  For each external entry in sorted order:

    If external entry has no fragments (fragments is absent or empty):
      a. Call ReadFileRaw(external.path) -> content.
      b. Compute SHA1Digest(content) and append to digest_list.

    If external entry has fragments:
      a. Initialize `combined` as an empty string.
      b. For each fragment in declaration order:
           Parse fragment.lines as a line range "<start>-<end>" (1-based,
           inclusive).
           Open a FileReader for external.path.
           Skip (start - 1) lines using SkipLines.
           Read (end - start + 1) lines. Normalize CRLF → LF.
           Append these lines (joined with LF) to `combined`.
           Close the reader.
      c. Compute SHA1Digest(combined) and append to digest_list.

**Step 5 — Target # Public content hash**

  a. Call ExtractSection(target_file_path, "# Public") -> pub_text.
  b. If pub_text is not empty:
       Compute SHA1Digest(pub_text) and append to digest_list.

**Step 6 — Target # Agent content hash**

  a. Call ExtractSection(target_file_path, "# Agent") -> agent_text.
  b. If agent_text is not empty:
       Compute SHA1Digest(agent_text) and append to digest_list.

**Step 7 — Input content hash**

  If target_fm.input is not empty:
    a. Call ResolveArtifactFilePath(target_fm.input) -> input_file_path.
    b. Call ReadFileStripFrontmatter(input_file_path) -> content.
    c. Compute SHA1Digest(content) and append to digest_list.

**Step 8 — Final hash**

  a. Concatenate all byte sequences in digest_list in order.
     Each SHA-1 digest is 20 raw bytes; the concatenation is
     (N × 20) bytes where N is the length of digest_list.
  b. Compute SHA1Digest of the concatenation -> final_bytes.
  c. Encode final_bytes as base64url (RFC 4648 §5, no padding).
     The result is 27 characters.
  d. Return the 27-character string.

---

## Error Summary

| Error | Trigger |
|---|---|
| `"invalid logical name: <name>"` | Logical name cannot be resolved, or ARTIFACT id not found |
| `"unreadable file: <path>"` | Any file in the chain cannot be opened or read |

---

## Invariants

- All content is read raw from disk, never from parsed or reconstructed data.
- The only normalization applied before hashing is CRLF → LF.
- Identical files on disk always produce the identical hash (deterministic).
- The output is always exactly 27 characters (base64url, no padding, SHA-1).
