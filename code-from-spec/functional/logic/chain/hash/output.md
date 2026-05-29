<!-- code-from-spec: ROOT/functional/logic/chain/hash@l4pnqGLJRfmstBcpzIHb-dqT44Y -->

# ChainHashCompute

## Overview

Receives a resolved `Chain` and returns a 27-character base64url-encoded SHA-1
hash representing the combined content of all positions in the chain.

---

## Functions

---

### ChainHashCompute(chain: Chain) -> string

Computes the chain hash for a resolved chain.

Parameters:
- `chain`: a Chain record as returned by ChainResolve

Returns:
- a 27-character base64url-encoded (RFC 4648 §5, no padding) SHA-1 string

Errors:
- "file unreadable": a file in the chain cannot be read or opened
- "parse failure": a node file cannot be parsed

Steps:

1. Initialize an empty list `content_hashes` to collect raw SHA-1 bytes
   (20 bytes each).

2. **Ancestors** — for each item in `chain.ancestors` (in order, root first):

   a. Call `NodeParse(ancestor.logical_name)`.
      If it fails, raise "parse failure".

   b. Call `HashSection(node.public)`.
      If the result is present, append it to `content_hashes`.

3. **Dependencies** — for each item in `chain.dependencies` (in the order
   provided, already alphabetically sorted and deduplicated by ChainResolve):

   a. If `LogicalNameIsArtifact(dep.logical_name)` is true:
      - Call `HashArtifactFile(dep.file_path)`.
      - Append the result to `content_hashes`.

   b. Else if `dep.qualifier` is absent:
      - Call `NodeParse(dep.logical_name)`.
        If it fails, raise "parse failure".
      - Call `HashSection(node.public)`.
        If the result is present, append it to `content_hashes`.

   c. Else (`dep.qualifier` is present):
      - Call `NodeParse(dep.logical_name)`.
        If it fails, raise "parse failure".
      - Call `HashSubsection(node.public, dep.qualifier)`.
        If the result is present, append it to `content_hashes`.

4. **External** — for each item in `chain.external` (in the order provided,
   already alphabetically sorted by ChainResolve):

   a. If `item.fragments` is absent:
      - Call `HashExternalFile(item.path)`.
      - Append the result to `content_hashes`.

   b. Else:
      - Call `HashExternalFileFragments(item.path, item.fragments)`.
      - Append the result to `content_hashes`.

5. **Target `# Public`**:

   a. Call `NodeParse(chain.target.logical_name)`.
      If it fails, raise "parse failure".
      Save the parsed node as `target_node` for reuse in step 6.

   b. Call `HashSection(target_node.public)`.
      If the result is present, append it to `content_hashes`.

6. **Target `# Agent`**:

   a. Call `HashSection(target_node.agent)`.
      If the result is present, append it to `content_hashes`.

7. **Input** — if `chain.input` is present:

   a. Call `HashArtifactFile(chain.input.file_path)`.
   b. Append the result to `content_hashes`.

8. **Final hash**:

   a. Concatenate all entries in `content_hashes` as raw bytes
      (each is 20 bytes; result length = 20 × count).

   b. Compute SHA-1 of the concatenated bytes.

   c. Encode the 20-byte SHA-1 result as base64url
      (RFC 4648 §5, no padding, URL-safe alphabet: A–Z, a–z, 0–9, `-`, `_`).
      The result is exactly 27 characters.

   d. Return the encoded string.

---

### HashSection(section: optional NodeSection) -> optional raw-bytes

Computes a SHA-1 content hash for a full section (e.g. `# Public` or
`# Agent`).

Parameters:
- `section`: an optional NodeSection

Returns:
- 20 raw bytes (SHA-1), or absent if the section contributes no content

Steps:

1. If `section` is absent, return absent.

2. If `section.content` is empty and `section.subsections` is empty,
   return absent.

3. Initialize `buf` as an empty byte buffer.

4. Append `section.raw_heading` + `\n` to `buf`.

5. For each line in `section.content`:
   Append `line` + `\n` to `buf`.

6. For each subsection in `section.subsections` (in order):
   a. Append `subsection.raw_heading` + `\n` to `buf`.
   b. For each line in `subsection.content`:
      Append `line` + `\n` to `buf`.

7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

### HashSubsection(section: optional NodeSection, qualifier: string) -> optional raw-bytes

Computes a SHA-1 content hash for a specific `##` subsection within a section.

Parameters:
- `section`: an optional NodeSection (typically `node.public`)
- `qualifier`: the qualifier string to look up (not yet normalized)

Returns:
- 20 raw bytes (SHA-1), or absent if the subsection is not found

Steps:

1. If `section` is absent, return absent.

2. Compute `target_heading` = `NormalizeText(qualifier)`.

3. Find the subsection in `section.subsections` whose `heading` equals
   `target_heading`. If not found, return absent.

4. Initialize `buf` as an empty byte buffer.

5. Append `subsection.raw_heading` + `\n` to `buf`.

6. For each line in `subsection.content`:
   Append `line` + `\n` to `buf`.

7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

---

### HashArtifactFile(file_path: PathCfs) -> raw-bytes

Computes a SHA-1 hash of an artifact file's content, stripping frontmatter
if present.

Parameters:
- `file_path`: the CFS path of the artifact file

Returns:
- 20 raw bytes (SHA-1)

Errors:
- "file unreadable": the file cannot be opened or read

Steps:

1. Call `FileOpen(file_path)`.
   If it fails, raise "file unreadable".
   Save the reader as `reader`.

2. Attempt to read the first line with `FileReadLine`.
   If "end of file" is raised, call `FileClose(reader)`.
   Compute SHA-1 of an empty buffer. Return the 20 bytes.

3. If the first line equals exactly `"---"`:
   a. Read and discard lines with `FileReadLine` until a line equals
      exactly `"---"` (the closing frontmatter delimiter).
      If "end of file" is raised before finding the closing delimiter,
      call `FileClose(reader)` and compute SHA-1 of empty buffer. Return.
   b. Continue to step 4 to read the remaining content.

   Else (first line is not `"---"`):
   a. Place the first line into a list `lines` — it is content, not frontmatter.
   b. Continue reading in step 4.

4. Initialize `buf` as an empty byte buffer.
   If `lines` already contains the first content line (non-frontmatter case),
   add it: append `line` + `\n` to `buf`.

5. Read remaining lines with `FileReadLine` one at a time.
   For each line: append `line` + `\n` to `buf`.
   Stop when "end of file" is raised.

6. Call `FileClose(reader)`.

7. Compute SHA-1 of `buf`. Return the 20 raw bytes.

Note: `FileClose` must be called in all cases, including error paths.

---

### HashExternalFile(path: string) -> raw-bytes

Computes a SHA-1 hash of a full external file's content.

Parameters:
- `path`: the path string from the `FrontmatterExternal` record

Returns:
- 20 raw bytes (SHA-1)

Errors:
- "file unreadable": the file cannot be opened or read

Steps:

1. Create a `PathCfs` from `path`.

2. Call `FileOpen(cfs_path)`.
   If it fails, raise "file unreadable".
   Save the reader as `reader`.

3. Initialize `buf` as an empty byte buffer.

4. Read all lines with `FileReadLine` one at a time.
   For each line: append `line` + `\n` to `buf`.
   Stop when "end of file" is raised.

5. Call `FileClose(reader)`.

6. Compute SHA-1 of `buf`. Return the 20 raw bytes.

Note: `FileClose` must be called in all cases, including error paths.

---

### HashExternalFileFragments(path: string, fragments: list of FrontmatterExternalFragment) -> raw-bytes

Computes a SHA-1 hash of selected line ranges from an external file,
concatenated in declaration order.

Parameters:
- `path`: the path string from the `FrontmatterExternal` record
- `fragments`: list of `FrontmatterExternalFragment` records, in declaration order

Returns:
- 20 raw bytes (SHA-1)

Errors:
- "file unreadable": the file cannot be opened or read

Steps:

1. Initialize `buf` as an empty byte buffer.

2. For each fragment in `fragments` (in declaration order):

   a. Parse `fragment.lines` as `"<start>-<end>"` where `start` and `end`
      are 1-based line numbers (inclusive).

   b. Create a `PathCfs` from `path`.
      Call `FileOpen(cfs_path)`.
      If it fails, raise "file unreadable".
      Save the reader as `reader`.

   c. Call `FileSkipLines(reader, start - 1)` to advance past lines before
      the fragment.

   d. Read `end - start + 1` lines with `FileReadLine`.
      For each line: append `line` + `\n` to `buf`.
      If "end of file" is raised before all lines are read, stop reading
      and call `FileClose(reader)`.

   e. Call `FileClose(reader)`.

3. Compute SHA-1 of `buf`. Return the 20 raw bytes.

Note: `FileClose` must be called after each fragment's file open, including
on error paths. Each fragment re-opens the file from the beginning.

---

## Contracts and Invariants

- All text normalization (CRLF → LF) is handled upstream by `FileReadLine`
  and `NodeParse`. `ChainHashCompute` does not perform additional normalization.

- Each line has `\n` appended before hashing, regardless of whether the
  original file ended with a newline. This ensures a deterministic
  representation.

- Positions with absent or empty content are skipped (contribute no bytes to
  the concatenation). They are not hashed as empty.

- The chain is assumed to be already deduplicated and sorted (as produced by
  `ChainResolve`). `ChainHashCompute` does not re-sort or deduplicate.

- The function is deterministic: the same files on disk always produce the
  same hash.

- `FileClose` is called in all cases — including error paths — to prevent
  file handle leaks.
