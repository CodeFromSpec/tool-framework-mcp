<!-- code-from-spec: ROOT/functional/utils/chain_hash@1ZciGipwCMwdf-xzAMKkP8P56rQ -->

# chain_hash

Computes the chain hash for a given spec node. The chain hash is a
27-character base64url-encoded SHA-1 digest used for artifact staleness
detection.

---

## Data Structures

```
record ContentHash
  raw_bytes: 20 bytes   -- raw SHA-1 digest, not encoded
```

---

## Helper: ExtractSection

```
function ExtractSection(file_content, heading) -> string

  Parameters:
    file_content  -- raw normalized text of the file (CRLF → LF already applied)
    heading       -- the exact heading line to search for (e.g. "# Public", "## Interface")

  Returns: the raw text from the heading line (inclusive) to the next
           heading at the same or higher level (exclusive), or end of file.
           Returns empty string if heading is not found.

  1. Split file_content into lines.

  2. Scan lines for a line that equals heading exactly.
     If no such line is found, return empty string.

  3. Determine the heading level by counting the leading "#" characters
     in heading (e.g. "# Public" → level 1, "## z" → level 2).

  4. Starting from the found heading line, collect lines until:
     - A line that starts with one to <level> "#" characters followed
       by a space is encountered (a heading at the same or higher level).
     - Or end of file is reached.
     Include the heading line itself; exclude the terminating heading line.

  5. Join the collected lines with LF and return.
```

---

## Helper: ExtractSubsection

```
function ExtractSubsection(file_content, parent_heading, subsection_heading) -> string

  Parameters:
    file_content         -- raw normalized text of the file
    parent_heading       -- e.g. "# Public"
    subsection_heading   -- e.g. "## z"

  Returns: the raw text of the subsection within the parent section.
           Returns empty string if either heading is not found.

  1. Call ExtractSection(file_content, parent_heading) to obtain the
     parent section text.
     If the result is empty, return empty string.

  2. Call ExtractSection(parent_section_text, subsection_heading) to
     obtain the subsection text.
     Return the result (may be empty).
```

---

## Helper: StripFrontmatter

```
function StripFrontmatter(file_content) -> string

  Returns the file content with the leading YAML frontmatter block removed.
  If no frontmatter is present, returns the content unchanged.

  1. Split file_content into lines.

  2. If the first line is not exactly "---", return file_content unchanged.

  3. Scan subsequent lines for the next line that is exactly "---".
     If not found, return file_content unchanged (treat as no frontmatter).

  4. Return the content from the line after the closing "---" to end of
     file, joined with LF.
     The returned string begins with LF if there is a blank line after "---",
     or with the first content character otherwise.
```

---

## Helper: NormalizeCRLF

```
function NormalizeCRLF(raw_bytes) -> string

  1. Replace every occurrence of CR+LF (byte sequence 0x0D 0x0A) with
     a single LF (0x0A).

  2. Return the resulting string.
```

---

## Helper: SHA1Digest

```
function SHA1Digest(content) -> 20 bytes

  1. Compute the SHA-1 hash of content (treated as a byte sequence).

  2. Return the raw 20-byte digest.
```

---

## Helper: ResolveArtifactFilePath

```
function ResolveArtifactFilePath(artifact_logical_name) -> string

  Parameters:
    artifact_logical_name -- an "ARTIFACT/x/y(id)" logical name

  Returns: the file path of the referenced artifact on disk.

  1. Call ResolveArtifactReference(artifact_logical_name) to obtain
     node_path and artifact_id.
     If it raises an error, raise error "invalid logical name: <artifact_logical_name>".

  2. Call ResolvePath(node_path) to obtain the _node.md file path for
     the node.

  3. Call ParseFrontmatter on the _node.md file path to obtain the
     node's frontmatter.
     If it raises an error, raise error "unreadable file: <node_path>".

  4. Search the frontmatter's outputs list for an entry whose id equals
     artifact_id.
     If not found, raise error "invalid logical name: artifact id <artifact_id> not found in <node_path>".

  5. Return the matched output's path.
```

---

## Main Function

```
function ComputeChainHash(logical_name) -> string

  Parameters:
    logical_name -- a ROOT/ logical name identifying the target spec node

  Returns: a 27-character base64url-encoded SHA-1 string (RFC 4648 §5,
           no padding).

  Errors:
    - "invalid logical name: <detail>"  -- cannot resolve the logical name
    - "unreadable file: <path>"         -- a required file cannot be read

  -- ----------------------------------------------------------------
  -- Preparation
  -- ----------------------------------------------------------------

  1. Verify that logical_name starts with "ROOT/".
     If not, raise error "invalid logical name: only ROOT/ names are supported".

  2. Call ResolvePath(logical_name) to get the target node's file path.
     Call OpenFileReader on that path.
     If the file is unreadable, raise error "unreadable file: <path>".
     Read the full file content and normalize CRLF → LF via NormalizeCRLF.
     This is target_content.

  3. Call ParseFrontmatter on the target node file path to obtain
     target_frontmatter.

  -- ----------------------------------------------------------------
  -- Accumulator
  -- ----------------------------------------------------------------

  4. Initialize digest_list as an empty list of 20-byte values.
     Each step below appends raw SHA-1 digests to digest_list.

  -- ----------------------------------------------------------------
  -- Step 1 — Ancestor # Public hashes
  -- ----------------------------------------------------------------

  5. Build the ancestor chain:
     a. Start with current = logical_name.
     b. Repeatedly call GetParent(current) and prepend the result to
        the ancestor list until GetParent raises "no parent".
        (This gives ancestors from ROOT down to the target's parent,
        not including the target itself.)

  6. For each ancestor in root-first order:
     a. Call ResolvePath(ancestor) to get the file path.
     b. Read the file content; normalize CRLF → LF.
        If unreadable, raise error "unreadable file: <path>".
     c. Call ExtractSection(content, "# Public").
        If the result is empty, skip (do not append a digest).
     d. Otherwise, compute SHA1Digest(section_text) and append to
        digest_list.

  -- ----------------------------------------------------------------
  -- Step 2 — depends_on hashes
  -- ----------------------------------------------------------------

  7. Collect all depends_on entries from target_frontmatter.
     Sort them alphabetically by their logical name string.

  8. For each depends_on entry (in sorted order):

     a. If the entry is a ROOT/ name without a qualifier
        (e.g. "ROOT/x/y"):
          i.  Call ResolvePath(entry) to get the file path.
          ii. Read and normalize the file content.
              If unreadable, raise error "unreadable file: <path>".
          iii.Call ExtractSection(content, "# Public").
              If empty, skip.
          iv. Compute SHA1Digest(section_text) and append to digest_list.

     b. If the entry is a ROOT/ name with a qualifier
        (e.g. "ROOT/x/y(z)"):
          i.  Call ExtractQualifier(entry) to get qualifier z.
          ii. Call ResolvePath(entry) to get the file path
              (qualifier is stripped automatically).
          iii.Read and normalize the file content.
              If unreadable, raise error "unreadable file: <path>".
          iv. Call ExtractSubsection(content, "# Public", "## <z>").
              If empty, skip.
          v.  Compute SHA1Digest(subsection_text) and append to
              digest_list.

     c. If the entry is an ARTIFACT/ name (e.g. "ARTIFACT/x/y(id)"):
          i.  Call ResolveArtifactFilePath(entry) to get the file path.
          ii. Read and normalize the file content.
              If unreadable, raise error "unreadable file: <path>".
          iii.Call StripFrontmatter(content).
          iv. Compute SHA1Digest(stripped_content) and append to
              digest_list.

  -- ----------------------------------------------------------------
  -- Step 3 — external file hashes
  -- ----------------------------------------------------------------

  9. Collect all external entries from target_frontmatter.
     Sort them alphabetically by their path field.

  10. For each external entry (in sorted order):

      a. If the entry has no fragments (fragments field is absent or empty):
           i.  Read the file at entry.path; normalize CRLF → LF.
               If unreadable, raise error "unreadable file: <entry.path>".
           ii. Compute SHA1Digest(full_content) and append to digest_list.

      b. If the entry has fragments:
           i.  Read the file at entry.path; normalize CRLF → LF.
               If unreadable, raise error "unreadable file: <entry.path>".
           ii. Split the file content into a list of lines.
           iii.For each fragment in declaration order:
                 - Parse the fragment's lines field as a line range
                   "<start>-<end>" (1-based, inclusive).
                 - Extract lines[start-1 .. end] (inclusive both ends).
                 - Join extracted lines with LF, appending a trailing LF.
                 - Append the fragment text to a concatenation buffer.
           iv. Compute SHA1Digest(concatenated_fragment_text) and append
               to digest_list.

  -- ----------------------------------------------------------------
  -- Step 4 — Target # Public hash
  -- ----------------------------------------------------------------

  11. Call ExtractSection(target_content, "# Public").
      If the result is not empty:
        Compute SHA1Digest(section_text) and append to digest_list.

  -- ----------------------------------------------------------------
  -- Step 5 — Target # Agent hash
  -- ----------------------------------------------------------------

  12. Call ExtractSection(target_content, "# Agent").
      If the result is not empty:
        Compute SHA1Digest(section_text) and append to digest_list.

  -- ----------------------------------------------------------------
  -- Step 6 — Input hash
  -- ----------------------------------------------------------------

  13. If target_frontmatter.input is not empty:
      a. Call ResolveArtifactFilePath(target_frontmatter.input) to get
         the artifact file path.
      b. Read and normalize the file content.
         If unreadable, raise error "unreadable file: <path>".
      c. Call StripFrontmatter(content).
      d. Compute SHA1Digest(stripped_content) and append to digest_list.

  -- ----------------------------------------------------------------
  -- Step 7 — Final hash
  -- ----------------------------------------------------------------

  14. Concatenate all entries in digest_list as raw bytes
      (each entry is exactly 20 bytes; the concatenation is
      20 × len(digest_list) bytes).

  15. Compute SHA1Digest(concatenated_raw_bytes) to get the final
      20-byte digest.

  16. Encode the final digest as base64url:
      - Use the RFC 4648 §5 alphabet
        ("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").
      - Omit padding characters ("=").
      - The result is exactly 27 characters.

  17. Return the 27-character string.
```

---

## Error Summary

| Error | Trigger |
|---|---|
| `"invalid logical name: <detail>"` | logical_name does not start with ROOT/, or an ARTIFACT/ reference cannot be resolved |
| `"unreadable file: <path>"` | any required file on disk cannot be opened or read |
