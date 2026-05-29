<!-- code-from-spec: ROOT/functional/logic/chain/hash@l4pnqGLJRfmstBcpzIHb-dqT44Y -->

# Chain Hash — Pseudocode

## Records

```
record ContentHash
  raw_bytes: 20 bytes (SHA-1 digest)
```

---

## Functions

---

### HashSection(section: NodeSection) -> optional ContentHash

Computes SHA-1 for a full section (# Public or # Agent).

1. If section is absent, return absent.

2. If section has no content lines and no subsections, return absent.

3. Initialize an empty byte buffer.

4. Append section.raw_heading + "\n" to the buffer.

5. For each line in section.content:
     Append line + "\n" to the buffer.

6. For each subsection in section.subsections:
     Append subsection.raw_heading + "\n" to the buffer.
     For each line in subsection.content:
       Append line + "\n" to the buffer.

7. Compute SHA-1 of the buffer. Return the result as a ContentHash.

---

### HashSubsection(section: NodeSection, qualifier: string) -> optional ContentHash

Computes SHA-1 for a specific ## subsection within # Public.

1. If section is absent, return absent.

2. Compute normalized_qualifier = NormalizeText(qualifier).

3. Find the subsection in section.subsections whose heading equals
   normalized_qualifier.
   If not found, return absent.

4. Initialize an empty byte buffer.

5. Append subsection.raw_heading + "\n" to the buffer.

6. For each line in subsection.content:
     Append line + "\n" to the buffer.

7. Compute SHA-1 of the buffer. Return the result as a ContentHash.

---

### HashFileStrippingFrontmatter(file_path: PathCfs) -> ContentHash
  errors:
    - file unreadable: file cannot be opened or read.

1. Call FileOpen(file_path).
   If it fails, raise error "file unreadable".
   Set reader to the result.

2. Call FileReadLine(reader) to read the first line.
   If "end of file" is raised:
     Call FileClose(reader).
     Return SHA-1 of empty bytes as a ContentHash.

3. If the first line equals exactly "---":
     Read lines until a line equals exactly "---".
       Discard each line.
       If "end of file" is raised before the closing "---":
         Call FileClose(reader).
         Raise error "file unreadable".
     Frontmatter has been skipped. Proceed to step 4 to read remaining lines.
   Else:
     The first line is content. Initialize buffer and append first_line + "\n".
     Proceed to step 4 to continue reading remaining lines.

4. Initialize buffer if not already initialized.
   Loop:
     Call FileReadLine(reader).
     If "end of file" is raised, stop the loop.
     Append line + "\n" to the buffer.

5. Call FileClose(reader).

6. Compute SHA-1 of the buffer. Return as a ContentHash.

---

### HashExternalFile(external: FrontmatterExternal) -> ContentHash
  errors:
    - file unreadable: file cannot be opened or read.

1. Create a PathCfs from external.path.

2. If external.fragments is absent or empty:
     Call FileOpen(path).
     If it fails, raise error "file unreadable".
     Set reader to the result.
     Initialize an empty byte buffer.
     Loop:
       Call FileReadLine(reader).
       If "end of file" is raised, stop the loop.
       Append line + "\n" to the buffer.
     Call FileClose(reader).
     Compute SHA-1 of the buffer. Return as a ContentHash.

3. Else (external has fragments):
     Initialize an empty byte buffer.
     For each fragment in external.fragments (in declaration order):
       Parse fragment.lines as "<start>-<end>" where start and end are
       integers (1-based, inclusive).
       Call FileOpen(path).
       If it fails, raise error "file unreadable".
       Set reader to the result.
       Call FileSkipLines(reader, start - 1).
       Read (end - start + 1) lines:
         For i from 1 to (end - start + 1):
           Call FileReadLine(reader).
           If "end of file" is raised:
             Call FileClose(reader).
             Raise error "file unreadable".
           Append line + "\n" to the buffer.
       Call FileClose(reader).
     Compute SHA-1 of the buffer. Return as a ContentHash.

---

### ChainHashCompute(chain: Chain) -> string
  errors:
    - file unreadable: a file in the chain cannot be read or opened.
    - parse failure: a node file cannot be parsed.

**Step 1 — Collect content hashes in chain assembly order**

Initialize an empty list: content_hashes.

**1a. Ancestors**

For each ancestor in chain.ancestors (from root down to target's parent):
  Call NodeParse(ancestor.logical_name).
  If it fails, raise error "parse failure".
  Call HashSection(node.public).
  If the result is not absent, append it to content_hashes.

**1b. Dependencies**

For each dep in chain.dependencies (already in alphabetical order by file path,
then qualifier):
  If LogicalNameIsArtifact(dep.logical_name) is true:
    Call HashFileStrippingFrontmatter(dep.file_path).
    If it fails, propagate error.
    Append the result to content_hashes.
  Else if dep.qualifier is absent:
    Call NodeParse(dep.logical_name).
    If it fails, raise error "parse failure".
    Call HashSection(node.public).
    If the result is not absent, append it to content_hashes.
  Else (dep.qualifier is present):
    Call NodeParse(dep.logical_name).
    If it fails, raise error "parse failure".
    Call HashSubsection(node.public, dep.qualifier).
    If the result is not absent, append it to content_hashes.

**1c. External files**

For each external in chain.external (already in alphabetical order by path):
  Call HashExternalFile(external).
  If it fails, propagate error.
  Append the result to content_hashes.

**1d. Target # Public**

Call NodeParse(chain.target.logical_name).
If it fails, raise error "parse failure".
Set target_node to the result.
Call HashSection(target_node.public).
If the result is not absent, append it to content_hashes.

**1e. Target # Agent**

Call HashSection(target_node.agent).
If the result is not absent, append it to content_hashes.

**1f. Input (if present)**

If chain.input is not absent:
  Call HashFileStrippingFrontmatter(chain.input.file_path).
  If it fails, propagate error.
  Append the result to content_hashes.

**Step 2 — Compute final hash**

1. Initialize an empty byte buffer.

2. For each content_hash in content_hashes (in order):
     Append content_hash.raw_bytes (20 bytes) to the buffer.

3. Compute SHA-1 of the buffer.

4. Encode the 20-byte SHA-1 result as base64url (RFC 4648 §5, no padding).
   The result is exactly 27 characters.

5. Return the 27-character string.
```
