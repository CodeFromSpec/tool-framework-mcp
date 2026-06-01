<!-- code-from-spec: ROOT/functional/logic/chain/hash@ojUxjMTkYUktO3iuIsQ4C_kiZL8 -->

# Interface

```
function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
    - FileUnreadable: a file in the chain cannot be read or opened.
    - ParseFailure: a node file cannot be parsed.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.
```

---

# ChainHashCompute

**Parameters:** `chain` — a `chainresolver.Chain` record  
**Returns:** 27-character base64url encoded SHA-1 string

---

## Helper: HashSpecSection(node, section_type) -> optional raw bytes (20)

`section_type` is one of: `"public"`, `"agent"`

1. If `section_type` is `"public"`, let `section` = `node.public`.
   Else let `section` = `node.agent`.

2. If `section` is absent, return absent.

3. If `section.content` is empty and `section.subsections` is empty,
   return absent.

4. Let `buf` = empty byte buffer.

5. Append `section.raw_heading` + `"\n"` to `buf`.

6. For each line in `section.content`:
   Append line + `"\n"` to `buf`.

7. For each subsection in `section.subsections`:
   Append `subsection.raw_heading` + `"\n"` to `buf`.
   For each line in `subsection.content`:
     Append line + `"\n"` to `buf`.

8. Return SHA-1(`buf`) as 20 raw bytes.

---

## Helper: HashSpecSubsection(node, qualifier) -> optional raw bytes (20)

1. Let `normalized_qualifier` = `NormalizeText(qualifier)`.

2. If `node.public` is absent, return absent.

3. Find the subsection in `node.public.subsections` whose `heading`
   equals `normalized_qualifier`.
   If not found, return absent.

4. Let `buf` = empty byte buffer.

5. Append `subsection.raw_heading` + `"\n"` to `buf`.

6. For each line in `subsection.content`:
   Append line + `"\n"` to `buf`.

7. Return SHA-1(`buf`) as 20 raw bytes.

---

## Helper: HashArtifactFile(file_path: pathutils.PathCfs) -> raw bytes (20)

1. Call `FileOpen(file_path)`.
   If it fails, raise error "file unreadable".

2. Let `buf` = empty byte buffer.

3. Call `FileReadLine(reader)` to read the first line.
   If `EndOfFile`, go to step 6.

4. If the first line equals `"---"`:
   a. Read lines with `FileReadLine` until a line equals `"---"`.
      Discard all lines read (frontmatter body and closing delimiter).
      If `EndOfFile` is raised before finding the closing `"---"`,
      go to step 6.
   b. Read remaining lines: call `FileReadLine` in a loop until
      `EndOfFile`. For each line, append line + `"\n"` to `buf`.
   Go to step 6.

5. (First line is not `"---"`) Append first_line + `"\n"` to `buf`.
   Read remaining lines with `FileReadLine` in a loop until `EndOfFile`.
   For each line, append line + `"\n"` to `buf`.

6. Call `FileClose(reader)`.

7. Return SHA-1(`buf`) as 20 raw bytes.

   If any error is raised in steps 2–5, call `FileClose(reader)`
   before propagating the error.

---

## Helper: HashExternalFile(path_string: string) -> raw bytes (20)

1. Create `PathCfs` with `value` = `path_string`.

2. Call `FileOpen(cfs_path)`.
   If it fails, raise error "file unreadable".

3. Let `buf` = empty byte buffer.

4. Read lines with `FileReadLine` in a loop until `EndOfFile`.
   For each line, append line + `"\n"` to `buf`.

5. Call `FileClose(reader)`.

6. Return SHA-1(`buf`) as 20 raw bytes.

   If any error is raised in steps 2–4, call `FileClose(reader)`
   before propagating the error.

---

## Step 1 — Collect content hashes

Let `hashes` = empty list of raw byte arrays (each 20 bytes).

### 1a. Ancestors

For each `ancestor` in `chain.ancestors` (in order):

1. Call `NodeParse(ancestor.logical_name)`.
   If it fails, raise error "parse failure".

2. Call `HashSpecSection(node, "public")`.
   If the result is present, append to `hashes`.

### 1b. Dependencies

For each `dep` in `chain.dependencies` (in order, already sorted):

1. If `LogicalNameIsArtifact(dep.logical_name)` is true:
   Call `HashArtifactFile(dep.file_path)`.
   Append result to `hashes`.

2. Else if `dep.qualifier` is absent:
   Call `NodeParse(dep.logical_name)`.
   If it fails, raise error "parse failure".
   Call `HashSpecSection(node, "public")`.
   If present, append to `hashes`.

3. Else (`dep.qualifier` is present):
   Call `NodeParse(dep.logical_name)`.
   If it fails, raise error "parse failure".
   Call `HashSpecSubsection(node, dep.qualifier)`.
   If present, append to `hashes`.

### 1c. External entries

For each `ext` in `chain.external` (in order, already sorted):

1. Call `HashExternalFile(ext.path)`.
   Append result to `hashes`.

### 1d. Target — `# Public`

1. Call `NodeParse(chain.target.logical_name)`.
   If it fails, raise error "parse failure".
   Save the result as `target_node`.

2. Call `HashSpecSection(target_node, "public")`.
   If present, append to `hashes`.

### 1e. Target — `# Agent`

1. Using `target_node` from step 1d:
   Call `HashSpecSection(target_node, "agent")`.
   If present, append to `hashes`.

### 1f. Input

1. If `chain.input` is absent, skip.

2. Call `HashArtifactFile(chain.input.file_path)`.
   Append result to `hashes`.

---

## Step 2 — Compute final hash

1. Concatenate all byte arrays in `hashes` in order → `raw_concat`.

2. Compute SHA-1(`raw_concat`) → 20-byte result.

3. Encode the 20-byte result as base64url (RFC 4648 §5, no padding).

4. Return the resulting 27-character string.
