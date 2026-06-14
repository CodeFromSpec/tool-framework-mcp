<!-- code-from-spec: ROOT/functional/logic/chain/hash@4cUqP9l_QMC-Vxl7WAvcKI46-KQ -->

## Interface

```
function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
    - ParseFailure: a node file cannot be parsed.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.
```

Receives a `Chain` (as returned by `ChainResolve`) and returns a 27-character
base64url encoded SHA-1 hash (RFC 4648 §5, no padding).

---

## Helper: ExtractBlock(content: list of string) -> string

1. Remove leading blank lines from `content` (lines that are empty or contain
   only U+0020 and U+0009).
2. Remove trailing blank lines from the result of step 1.
3. If nothing remains, return empty string.
4. Join the remaining lines with `\n`, then append one `\n` at the end.
5. Return the joined string.

---

## Helper: ConcatenateSubsections(subsections: list of NodeSubsection) -> string

1. Initialize `result` as an empty string.
2. For each subsection in `subsections`, in document order:
   a. Compute `heading_line` = subsection.raw_heading with trailing whitespace
      removed, followed by `\n`.
   b. Compute `body` = ExtractBlock(subsection.content).
   c. If `result` is not empty, append one `\n` to `result` (blank line
      separator between consecutive blocks).
   d. Append `heading_line` to `result`.
   e. Append `body` to `result`.
3. Return `result`.

---

## Helper: ExtractAgentSection(section: NodeSection) -> string

1. Compute `heading_line` = section.raw_heading with trailing whitespace
   removed, followed by `\n`.
2. Compute `body` = ExtractBlock(section.content).
3. Initialize `result` = `heading_line` + `body`.
4. For each subsection in `section.subsections`, in document order:
   a. Compute `sub_heading_line` = subsection.raw_heading with trailing
      whitespace removed, followed by `\n`.
   b. Compute `sub_body` = ExtractBlock(subsection.content).
   c. Append one `\n` to `result` (blank line separator).
   d. Append `sub_heading_line` to `result`.
   e. Append `sub_body` to `result`.
5. Return `result`.

---

## Helper: NeutralizeArtifactTag(line: string) -> string

1. Search `line` for the pattern:
   `code-from-spec: <anything>@<27 base64url characters>`
   where base64url characters are A-Z, a-z, 0-9, `-`, `_`.
2. If a match is found, replace the 27-character hash portion with
   `---------------------------` (27 hyphens). Preserve everything else on
   the line, including the logical name.
3. Return the (possibly modified) line.

---

## Helper: HashPublicSubsections(node: parsenode.Node) -> optional raw bytes (20 bytes)

1. If `node.public` is absent, return absent.
2. If `node.public.subsections` is empty, return absent.
3. Compute `content` = ConcatenateSubsections(node.public.subsections).
4. Compute SHA-1 of `content` (as UTF-8 bytes).
5. Return the raw 20-byte SHA-1 digest.

---

## Helper: HashQualifiedSubsection(node: parsenode.Node, qualifier: string) -> optional raw bytes (20 bytes)

1. Compute `normalized_qualifier` = NormalizeText(qualifier).
2. Search `node.public.subsections` for an entry whose `heading` field equals
   `normalized_qualifier`.
3. If not found, return absent.
4. Let `sub` be the found subsection.
5. Compute `heading_line` = sub.raw_heading with trailing whitespace removed,
   followed by `\n`.
6. Compute `body` = ExtractBlock(sub.content).
7. Compute `content` = `heading_line` + `body`.
8. Compute SHA-1 of `content` (as UTF-8 bytes).
9. Return the raw 20-byte SHA-1 digest.

---

## Helper: HashAgentSection(node: parsenode.Node) -> optional raw bytes (20 bytes)

1. If `node.agent` is absent, return absent.
2. If `node.agent.content` is empty and `node.agent.subsections` is empty,
   return absent.
3. Compute `content` = ExtractAgentSection(node.agent).
4. Compute SHA-1 of `content` (as UTF-8 bytes).
5. Return the raw 20-byte SHA-1 digest.

---

## Helper: HashArtifactFile(file_path: pathutils.PathCfs) -> raw bytes (20 bytes)

1. Call `FileOpen(file_path)`.
   If `FileOpen` raises an error, raise "file unreadable".
2. Initialize `content` as an empty string.
3. Loop:
   a. Call `FileReadLine(reader)`.
   b. If `FileReadLine` raises "end of file", exit the loop.
   c. Compute `neutralized` = NeutralizeArtifactTag(line).
   d. Append `neutralized` + `\n` to `content`.
4. Call `FileClose(reader)`.
   (Call `FileClose` even if an error occurs.)
5. Compute SHA-1 of `content` (as UTF-8 bytes).
6. Return the raw 20-byte SHA-1 digest.

---

## Helper: HashExternalFile(file_path: pathutils.PathCfs) -> raw bytes (20 bytes)

1. Call `FileOpen(file_path)`.
   If `FileOpen` raises an error, raise "file unreadable".
2. Initialize `content` as an empty string.
3. Loop:
   a. Call `FileReadLine(reader)`.
   b. If `FileReadLine` raises "end of file", exit the loop.
   c. Append `line` + `\n` to `content`.
4. Call `FileClose(reader)`.
   (Call `FileClose` even if an error occurs.)
5. Compute SHA-1 of `content` (as UTF-8 bytes).
6. Return the raw 20-byte SHA-1 digest.

---

## function ChainHashCompute(chain: chainresolver.Chain) -> string

1. Initialize `raw_hashes` as an empty list of byte sequences (each 20 bytes).

**Step 1 — Ancestors**

2. For each `ancestor` in `chain.ancestors` (in order, root to target's parent):
   a. Call `NodeParse(ancestor.unqualified_logical_name)`.
      If `NodeParse` fails, raise "parse failure".
   b. Compute `h` = HashPublicSubsections(node).
   c. If `h` is not absent, append `h` to `raw_hashes`.

**Step 2 — Dependencies**

3. For each `dep` in `chain.dependencies` (in alphabetical order by logical name):
   a. If `LogicalNameIsArtifact(dep.unqualified_logical_name)`:
      - Compute `h` = HashArtifactFile(dep.file_path).
      - Append `h` to `raw_hashes`.
   b. Else if `LogicalNameIsExternal(dep.unqualified_logical_name)`:
      - Compute `h` = HashExternalFile(dep.file_path).
      - Append `h` to `raw_hashes`.
   c. Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` and
      `dep.qualifier` is absent:
      - Call `NodeParse(dep.unqualified_logical_name)`.
        If `NodeParse` fails, raise "parse failure".
      - Compute `h` = HashPublicSubsections(node).
      - If `h` is not absent, append `h` to `raw_hashes`.
   d. Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` and
      `dep.qualifier` is present:
      - Call `NodeParse(dep.unqualified_logical_name)`.
        If `NodeParse` fails, raise "parse failure".
      - Compute `h` = HashQualifiedSubsection(node, dep.qualifier).
      - If `h` is not absent, append `h` to `raw_hashes`.

**Step 3 — Target `# Public`**

4. Call `NodeParse(chain.target.unqualified_logical_name)`.
   If `NodeParse` fails, raise "parse failure".
   Store the result as `target_node`.
5. Compute `h_public` = HashPublicSubsections(target_node).
   If `h_public` is not absent, append `h_public` to `raw_hashes`.

**Step 4 — Target `# Agent`**

6. Compute `h_agent` = HashAgentSection(target_node).
   If `h_agent` is not absent, append `h_agent` to `raw_hashes`.

**Step 5 — Input**

7. If `chain.input` is present:
   a. Let `input` = `chain.input`.
   b. If `LogicalNameIsArtifact(input.unqualified_logical_name)`:
      - Compute `h` = HashArtifactFile(input.file_path).
      - Append `h` to `raw_hashes`.
   c. Else if `LogicalNameIsExternal(input.unqualified_logical_name)`:
      - Compute `h` = HashExternalFile(input.file_path).
      - Append `h` to `raw_hashes`.

**Step 6 — Final hash**

8. Concatenate all entries in `raw_hashes` as a single byte sequence
   (each is 20 bytes, in the order they were appended).
9. Compute SHA-1 of the concatenated bytes.
10. Encode the 20-byte SHA-1 result as base64url (RFC 4648 §5, no padding),
    producing a 27-character string.
11. Return the 27-character string.

---

## Contracts

- `FileReadLine` normalizes CRLF to LF before returning — no additional CRLF
  normalization is needed when reading files.
- `NodeParse` also normalizes CRLF to LF — spec node content is always LF by
  the time it reaches this module.
- Positions with missing or empty content (absent sections, no subsections,
  empty extracted text) contribute no hash to `raw_hashes` — they are skipped,
  not hashed as empty.
- Deterministic: same files on disk always produce the same hash.
- Dependencies are deduplicated by `ChainResolve` before being passed to
  `ChainHashCompute` — this function processes them as given.
