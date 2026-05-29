<!-- code-from-spec: ROOT/functional/logic/chain/hash@LnlHpQD8K08ZDj1G6yTTQx6Or7E -->

# ChainHashCompute

function ChainHashCompute(chain: Chain) -> string
  errors:
    - file unreadable: a file in the chain cannot be read or opened.
    - parse failure: a node file cannot be parsed.

Returns a 27-character base64url-encoded SHA-1 hash representing
the combined content state of all positions in the chain.

---

## Step 1 — Collect content hashes

Process each position in chain assembly order. For each position,
compute a SHA-1 content hash (20 raw bytes). Skip any position
whose relevant content is absent or empty — contribute nothing
to the list.

### 1a. Ancestors

For each item in `chain.ancestors` (root down to target's parent):

  1. Call NodeParse with `ancestor.logical_name`.
     If NodeParse fails, raise error "parse failure".

  2. Hash the `# Public` section using the full-section rule
     (see "Hashing a full section" below).
     If the section is absent or has no content and no subsections, skip.

  3. Append the resulting content hash to the list.

### 1b. Dependencies

For each item in `chain.dependencies` (alphabetical order by file_path,
then qualifier):

  1. If LogicalNameIsArtifact(dep.logical_name) is true:
       Hash the artifact file at `dep.file_path`
       (see "Hashing artifact files" below).
       Append the content hash to the list.

  2. Else if dep.qualifier is absent:
       Call NodeParse with `dep.logical_name`.
       If NodeParse fails, raise error "parse failure".
       Hash the `# Public` section (full section).
       If absent or empty, skip.
       Append the content hash to the list.

  3. Else (qualifier is present):
       Call NodeParse with `dep.logical_name`.
       If NodeParse fails, raise error "parse failure".
       Hash the `## <qualifier>` subsection within `# Public`
       (see "Hashing a subsection" below).
       If not found, skip.
       Append the content hash to the list.

### 1c. External files

For each item in `chain.external` (alphabetical order by path):

  1. If `external.fragments` is absent:
       Hash the full file content
       (see "Hashing external files — no fragments" below).

  2. Else:
       Hash by fragment ranges in declaration order
       (see "Hashing external files — with fragments" below).

  3. Append the content hash to the list.

### 1d. Target — # Public

  1. Call NodeParse with `chain.target.logical_name`.
     If NodeParse fails, raise error "parse failure".
     Store the result as `target_node` for reuse in 1e.

  2. Hash the `# Public` section (full section).
     If absent or empty, skip.
     Append the content hash to the list.

### 1e. Target — # Agent

  1. Using the `target_node` result from step 1d:

  2. Hash the `# Agent` section (full section).
     If absent or empty, skip.
     Append the content hash to the list.

### 1f. Input

  If `chain.input` is present:

  1. Hash the artifact file at `chain.input.file_path`
     (see "Hashing artifact files" below).
     Append the content hash to the list.

---

## Step 2 — Compute final hash

  1. Concatenate all content hashes from Step 1 as raw bytes,
     in the order they were appended (20 bytes each).

  2. Compute SHA-1 of the concatenated bytes.

  3. Encode the 20-byte SHA-1 result as base64url
     (RFC 4648 §5, no padding).

  4. Return the resulting 27-character string.

---

## Hashing a full section

Given a NodeSection (e.g., the `# Public` or `# Agent` section):

  1. If the section is absent, return absent (skip).

  2. If the section has no content lines and no subsections,
     return absent (skip).

  3. Collect lines to hash:
       - The section's `raw_heading` line.
       - All lines in `section.content`.
       - For each subsection in `section.subsections` (in order):
           - The subsection's `raw_heading` line.
           - All lines in `subsection.content`.

  4. After each line collected above, append "\n" (LF).

  5. Compute SHA-1 of the resulting byte sequence.

  6. Return the 20-byte SHA-1.

---

## Hashing a subsection

Given a qualifier string and a NodeSection (`# Public`):

  1. If `node.public` is absent, return absent (skip).

  2. Search `node.public.subsections` for a subsection whose
     `heading` matches the qualifier (case-insensitive).
     If not found, return absent (skip).

  3. Collect lines to hash:
       - The subsection's `raw_heading` line.
       - All lines in `subsection.content`.

  4. After each line collected above, append "\n" (LF).

  5. Compute SHA-1 of the resulting byte sequence.

  6. Return the 20-byte SHA-1.

---

## Hashing artifact files

Given a file_path (PathCfs) for an artifact or input file:

  1. Call FileOpen with `file_path`.
     If the file cannot be opened, raise error "file unreadable".

  2. Attempt to read the first line with FileReadLine.
     If end of file immediately, the file is empty:
       Call FileClose.
       Return absent (skip — no content hash contributed).

  3. If the first line is exactly "---" (frontmatter delimiter):
       Read and discard lines until a line that is exactly "---"
       is encountered (the closing delimiter).
       If end of file is reached before finding the closing "---",
       treat the rest of the file as content (no frontmatter end found).

  4. Read all remaining lines with FileReadLine until end of file.

  5. For each content line, append "\n" after it.

  6. If there are no content lines (file had only frontmatter),
     Call FileClose.
     Return absent (skip).

  7. Compute SHA-1 of the collected byte sequence.

  8. Call FileClose.

  9. Return the 20-byte SHA-1.

---

## Hashing external files — no fragments

Given a file_path (PathCfs) with no fragments declared:

  1. Call FileOpen with `file_path`.
     If the file cannot be opened, raise error "file unreadable".

  2. Read all lines with FileReadLine until end of file.
     After each line, append "\n".

  3. Compute SHA-1 of the collected byte sequence.

  4. Call FileClose.

  5. Return the 20-byte SHA-1.

---

## Hashing external files — with fragments

Given a file_path (PathCfs) and a list of fragments in declaration order:

  1. Collect all fragment byte sequences:

     For each fragment in declaration order:

       a. Parse `fragment.lines` as "<start>-<end>" (1-based, inclusive integers).

       b. Call FileOpen with `file_path`.
          If the file cannot be opened, raise error "file unreadable".

       c. Call FileSkipLines with count = start - 1,
          to skip to the first line of the fragment.

       d. Read (end - start + 1) lines with FileReadLine.
          After each line, append "\n".

       e. Call FileClose.

       f. Append the collected bytes to the fragment accumulator.

  2. Compute SHA-1 of the concatenated fragment bytes (all fragments joined).

  3. Return the 20-byte SHA-1.
