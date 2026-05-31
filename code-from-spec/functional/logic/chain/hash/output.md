<!-- code-from-spec: ROOT/functional/logic/chain/hash@52_WIlO3-KYupfvM8y3ikzfTFkM -->

# Chain Hash

## Functions

---

### ChainHashCompute(chain: Chain) -> string

Receives a resolved `Chain` record and returns a 27-character base64url-encoded SHA-1 string.

**Errors:**
- `FileUnreadable`: a file in the chain cannot be read or opened.
- `ParseFailure`: a node file cannot be parsed.
- `(FileReader.*)`: propagated from `FileOpen`.
- `(NodeParsing.*)`: propagated from `NodeParse`.

---

#### Step 1 — Collect content hashes

Maintain an ordered list of raw SHA-1 digests (20 bytes each).
Process positions in the following order:

**1a. Ancestors**

For each item in `chain.ancestors`:
1. Call `NodeParse(ancestor.logical_name)`.
   If it fails, raise error `"parse failure"`.
2. Hash the `# Public` section using the full-section hashing procedure (see below).
3. If the section is absent or empty (no content and no subsections), skip — do not add a digest.
4. Otherwise append the resulting SHA-1 digest to the list.

**1b. Dependencies**

For each item in `chain.dependencies`:

- If `LogicalNameIsArtifact(dep.logical_name)` is true:
  1. Hash the artifact file at `dep.file_path` using the artifact-file hashing procedure (see below).
  2. Append the resulting digest to the list.

- Else if `dep.qualifier` is absent:
  1. Call `NodeParse(dep.logical_name)`.
     If it fails, raise error `"parse failure"`.
  2. Hash the `# Public` section (full section).
  3. If absent or empty, skip.
  4. Otherwise append the digest to the list.

- Else (`dep.qualifier` is present):
  1. Call `NodeParse(dep.logical_name)`.
     If it fails, raise error `"parse failure"`.
  2. Hash the `## <dep.qualifier>` subsection within `# Public` using the subsection hashing procedure (see below).
  3. If the subsection is not found, skip.
  4. Otherwise append the digest to the list.

**1c. External entries**

For each item in `chain.external`:

- If the entry has no fragments:
  1. Create a `PathCfs` from `entry.path`.
  2. Call `FileOpen(path_cfs)`. If it fails, raise error `"file unreadable"`.
  3. Read all lines with `FileReadLine` until `EndOfFile`.
  4. Call `FileClose`.
  5. For each line, append the line text followed by `\n`.
  6. Compute SHA-1 of the concatenated bytes.
  7. Append the digest to the list.

- If the entry has fragments:
  1. Create a `PathCfs` from `entry.path`.
  2. Initialize an empty byte buffer for the concatenation.
  3. For each fragment in declaration order:
     a. Parse `fragment.lines` as `<start>-<end>` (1-based, inclusive integers).
     b. Call `FileOpen(path_cfs)`. If it fails, raise error `"file unreadable"`.
     c. Call `FileSkipLines(reader, start - 1)` to skip lines before the range.
     d. Read `end - start + 1` lines with `FileReadLine`. For each line, append the line text followed by `\n` to the buffer.
     e. Call `FileClose`.
  4. Compute a single SHA-1 of the entire concatenated buffer.
  5. Append the digest to the list.

**1d. Target — `# Public` section**

1. Call `NodeParse(chain.target.logical_name)`.
   If it fails, raise error `"parse failure"`.
2. Hash the `# Public` section (full section).
3. If absent or empty, skip.
4. Otherwise append the digest to the list.

**1e. Target — `# Agent` section**

Using the same `NodeParse` result from step 1d:
1. Hash the `# Agent` section (full section).
2. If absent or empty, skip.
3. Otherwise append the digest to the list.

**1f. Input**

If `chain.input` is present:
1. Hash the artifact file at `chain.input.file_path` using the artifact-file hashing procedure (see below).
2. Append the resulting digest to the list.

---

#### Step 2 — Compute final hash

1. Concatenate all digests collected in step 1 as raw bytes (20 bytes each), in the order they were appended.
2. Compute SHA-1 of the concatenated bytes.
3. Encode the resulting 20 bytes as base64url (RFC 4648 §5, no padding).
4. Return the resulting 27-character string.

---

## Hashing procedures

### Full-section hashing

Used for `# Public` and `# Agent` sections of a parsed node.

Given a `NodeSection`:
1. If the section is absent, return absent (skip).
2. Initialize an empty byte buffer.
3. Append the section's `raw_heading` line followed by `\n`.
4. For each line in `section.content`, append the line followed by `\n`.
5. For each subsection in `section.subsections`:
   a. Append the subsection's `raw_heading` line followed by `\n`.
   b. For each line in `subsection.content`, append the line followed by `\n`.
6. If the buffer contains only the heading line (content is empty and subsections list is empty), return absent (skip).
7. Compute SHA-1 of the buffer.
8. Return the digest.

### Subsection hashing

Used for `## <qualifier>` entries within `# Public`.

Given a parsed node and a qualifier string:
1. If `node.public` is absent, return absent (skip).
2. Compute `normalized_qualifier` = `NormalizeText(qualifier)`.
3. Find the subsection in `node.public.subsections` whose `heading` equals `normalized_qualifier`.
   If not found, return absent (skip).
4. Initialize an empty byte buffer.
5. Append the subsection's `raw_heading` line followed by `\n`.
6. For each line in `subsection.content`, append the line followed by `\n`.
7. Compute SHA-1 of the buffer.
8. Return the digest.

### Artifact-file hashing

Used for `ARTIFACT/` dependencies and `input`.

Given a `file_path` (`PathCfs`):
1. Call `FileOpen(file_path)`. If it fails, raise error `"file unreadable"`.
2. Read the first line with `FileReadLine`.
   - If `EndOfFile`, call `FileClose` and return the SHA-1 of an empty byte sequence.
3. If the first line is exactly `"---"`:
   a. Read and discard lines with `FileReadLine` until a line that is exactly `"---"` is found (closing frontmatter delimiter). Discard that line too.
   b. Continue to step 4.
   If the first line is not `"---"`:
   a. Treat the first line as the first content line. Include it in the buffer (step 4).
4. Initialize an empty byte buffer.
   If skipping frontmatter, the first content line has not yet been read — proceed to the loop.
   If not skipping, include the first line already read.
5. Read all remaining lines with `FileReadLine` until `EndOfFile`. For each line, append the line followed by `\n` to the buffer.
6. Call `FileClose`.
7. Compute SHA-1 of the buffer.
8. Return the digest.

Note: Call `FileClose` in all cases, including error paths, to avoid leaking the file handle.
