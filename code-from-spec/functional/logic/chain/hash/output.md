<!-- code-from-spec: ROOT/functional/logic/chain/hash@OLmwFyPww5GsRZqRDDjkYnRSE2w -->

# Chain Hash

Computes a 27-character base64url-encoded SHA-1 hash representing
the chain for a given resolved `Chain` record. Used for artifact
staleness detection.

---

## Records

Uses `chainresolver.Chain`, `chainresolver.ChainItem`,
`frontmatter.FrontmatterExternal`, `frontmatter.FrontmatterExternalFragment`,
`pathutils.PathCfs`, `parsenode.Node`, `parsenode.NodeSection`,
`parsenode.NodeSubsection` as defined in their respective modules.

---

## Functions

```
function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
    - FileUnreadable: a file in the chain cannot be read or opened.
    - ParseFailure: a node file cannot be parsed.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.
```

Receives a `Chain` (as returned by `ChainResolve`) and returns a
27-character base64url-encoded SHA-1 hash.

Steps:

  1. Initialize an empty list `content_hashes` of raw SHA-1 byte
     sequences (20 bytes each).

  2. **Process ancestors.**
     For each ancestor in `chain.ancestors` (in order, root first):
       a. Call `NodeParse(ancestor.logical_name)`.
          If it fails, raise error "parse failure".
       b. Call `HashFullSection(node.public)`.
          If the result is present, append it to `content_hashes`.

  3. **Process dependencies.**
     For each dep in `chain.dependencies` (in the order provided
     — already sorted alphabetically by file path, then qualifier):
       a. If `LogicalNameIsArtifact(dep.logical_name)` is true:
            Call `HashArtifactFile(dep.file_path)`.
            If the result is present, append it to `content_hashes`.
       b. Else if `dep.qualifier` is absent:
            Call `NodeParse(dep.logical_name)`.
            If it fails, raise error "parse failure".
            Call `HashFullSection(node.public)`.
            If the result is present, append it to `content_hashes`.
       c. Else (qualifier is present):
            Call `NodeParse(dep.logical_name)`.
            If it fails, raise error "parse failure".
            Call `HashSubsection(node.public, dep.qualifier)`.
            If the result is present, append it to `content_hashes`.

  4. **Process external entries.**
     For each ext in `chain.external` (in the order provided —
     already sorted alphabetically by path):
       a. If `ext.fragments` is absent:
            Call `HashExternalFile(ext.path)`.
            If the result is present, append it to `content_hashes`.
       b. Else:
            Call `HashExternalFragments(ext.path, ext.fragments)`.
            If the result is present, append it to `content_hashes`.

  5. **Process target.**
     Call `NodeParse(chain.target.logical_name)`.
     If it fails, raise error "parse failure".
       a. Call `HashFullSection(node.public)`.
          If the result is present, append it to `content_hashes`.
       b. Call `HashFullSection(node.agent)`.
          If the result is present, append it to `content_hashes`.

  6. **Process input.**
     If `chain.input` is present:
       Call `HashArtifactFile(chain.input.file_path)`.
       If the result is present, append it to `content_hashes`.

  7. **Compute final hash.**
     Concatenate all raw byte sequences in `content_hashes` in order.
     Compute SHA-1 of the concatenation.
     Encode the 20-byte result as base64url (RFC 4648 §5, no padding).
     Return the resulting 27-character string.

---

```
function HashFullSection(section: optional parsenode.NodeSection) -> optional raw-bytes
```

Hashes a full spec section (`# Public` or `# Agent`).

Steps:

  1. If `section` is absent, return absent.

  2. If `section.content` is empty and `section.subsections` is empty,
     return absent.

  3. Initialize an empty byte buffer `buf`.

  4. Append the bytes of `section.raw_heading`, then `\n`, to `buf`.

  5. For each line in `section.content`:
       Append the bytes of the line, then `\n`, to `buf`.

  6. For each subsection in `section.subsections` (in order):
       a. Append the bytes of `subsection.raw_heading`, then `\n`, to `buf`.
       b. For each line in `subsection.content`:
            Append the bytes of the line, then `\n`, to `buf`.

  7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

```
function HashSubsection(public_section: optional parsenode.NodeSection, qualifier: string) -> optional raw-bytes
```

Hashes a specific `##` subsection within `# Public`, matched by
normalized heading.

Steps:

  1. If `public_section` is absent, return absent.

  2. Compute `target_heading` = `NormalizeText(qualifier)`.

  3. Find the subsection in `public_section.subsections` whose
     `heading` equals `target_heading`.
     If not found, return absent.

  4. Initialize an empty byte buffer `buf`.

  5. Append the bytes of `subsection.raw_heading`, then `\n`, to `buf`.

  6. For each line in `subsection.content`:
       Append the bytes of the line, then `\n`, to `buf`.

  7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

```
function HashArtifactFile(file_path: pathutils.PathCfs) -> optional raw-bytes
```

Hashes an artifact file, stripping any frontmatter block at the top.

Steps:

  1. Call `FileOpen(file_path)`.
     If it fails, call `FileClose` on any open handle and raise
     error "file unreadable".

  2. Read the first line with `FileReadLine`.
     If `EndOfFile` is raised, call `FileClose` and return absent.

  3. Initialize an empty byte buffer `buf`.

  4. If the first line equals `"---"` (frontmatter start):
       a. Read lines with `FileReadLine` until a line equals `"---"`
          (closing delimiter) or `EndOfFile` is raised.
          Discard all these lines (they are frontmatter).
       b. If `EndOfFile` is raised before the closing `"---"`,
          call `FileClose` and return absent.
       c. Continue reading the remaining lines of the file with
          `FileReadLine`. For each line (until `EndOfFile`):
            Append the bytes of the line, then `\n`, to `buf`.
     Else (no frontmatter):
       a. Append the bytes of the first line, then `\n`, to `buf`.
       b. Read remaining lines with `FileReadLine`. For each line
          (until `EndOfFile`):
            Append the bytes of the line, then `\n`, to `buf`.

  5. Call `FileClose`.

  6. If `buf` is empty, return absent.

  7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

```
function HashExternalFile(path: string) -> optional raw-bytes
```

Hashes the full content of an external file.

Steps:

  1. Create `cfs_path` as a `PathCfs` with `value` = `path`.

  2. Call `FileOpen(cfs_path)`.
     If it fails, raise error "file unreadable".

  3. Initialize an empty byte buffer `buf`.

  4. Read all lines with `FileReadLine` until `EndOfFile`:
       For each line, append the bytes of the line, then `\n`, to `buf`.

  5. Call `FileClose`.

  6. If `buf` is empty, return absent.

  7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

```
function HashExternalFragments(path: string, fragments: list of frontmatter.FrontmatterExternalFragment) -> optional raw-bytes
```

Hashes selected line ranges from an external file, concatenated in
declaration order.

Steps:

  1. Create `cfs_path` as a `PathCfs` with `value` = `path`.

  2. Initialize an empty byte buffer `buf`.

  3. For each fragment in `fragments` (in declaration order):
       a. Parse `fragment.lines` as `"<start>-<end>"` where `start`
          and `end` are 1-based line numbers (inclusive).
          If parsing fails, raise error "file unreadable".
       b. Call `FileOpen(cfs_path)`.
          If it fails, raise error "file unreadable".
       c. Call `FileSkipLines(reader, start - 1)` to skip lines
          before the range.
       d. Read `end - start + 1` lines with `FileReadLine`.
          For each line, append the bytes of the line, then `\n`,
          to `buf`.
          If `EndOfFile` is raised before reading all expected lines,
          call `FileClose` and stop reading (use what was read).
       e. Call `FileClose`.

  4. If `buf` is empty, return absent.

  5. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

## Error Conditions

- "file unreadable" — a required file could not be opened or read.
  Raised when `FileOpen` fails for any reason (file does not exist,
  permission denied, or other OS error).
- "parse failure" — a `_node.md` file could not be parsed by
  `NodeParse`. Raised when `NodeParse` returns any error.
- `FileReader.*` errors — propagated directly from `FileOpen`
  and related calls.
- `NodeParsing.*` errors — propagated directly from `NodeParse`.

---

## Contracts and Invariants

- CRLF normalization is handled by `FileReadLine` and `NodeParse`;
  no additional normalization is applied here.
- Every line appended to a buffer is always followed by `\n`,
  regardless of the original file's line endings or trailing
  newline presence.
- Positions with absent or empty content contribute no bytes and
  no hash to the final concatenation (they are skipped entirely).
- `FileClose` is always called — including on error paths — to
  avoid leaking file handles.
- The final hash is deterministic: the same files on disk always
  produce the same 27-character string.
- The 27-character output is SHA-1 encoded as base64url
  (RFC 4648 §5, no padding).
