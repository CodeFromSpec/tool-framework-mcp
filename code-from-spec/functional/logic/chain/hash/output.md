<!-- code-from-spec: SPEC/functional/logic/chain/hash@ZuKx2eJMnw-i6pmeAlKUF-YaXmo -->

## Namespace

    namespace: chainhash

## Interface

```
function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
    - ParseFailure: a node file cannot be parsed.
    - FileUnreadable: an artifact or external file cannot be opened.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.
```

Receives a `Chain` (as returned by `ChainResolve`) and returns a
27-character base64url-encoded SHA-1 hash.

---

## Block Extraction

**ExtractBlock(content: list of string) -> string**

1. Remove leading blank lines from `content`
   (lines that are empty or contain only spaces and tabs).
2. Remove trailing blank lines.
3. If nothing remains, return empty string.
4. Join remaining lines with `\n` and append exactly one `\n`.

**FormatSection(raw_heading: string, content: list of string) -> string**

1. Let `head` = `raw_heading` with trailing whitespace removed, followed by `\n`.
2. Let `body` = `ExtractBlock(content)`.
3. Return concatenation of `head` and `body`.

**ConcatenateSubsections(subsections: list of NodeSubsection) -> string**

1. Let `result` = empty string.
2. For each subsection in `subsections`:
   a. Let `block` = `FormatSection(subsection.raw_heading, subsection.content)`.
   b. If `result` is not empty and `block` is not empty, append `\n` to `result`.
   c. Append `block` to `result`.
3. Return `result`.

---

## Content Hash Helpers

**HashPublicSubsections(node: parsenode.Node) -> optional raw bytes (20)**

1. If `node.public` is absent, return absent.
2. If `node.public.subsections` is empty, return absent.
3. Let `text` = `ConcatenateSubsections(node.public.subsections)`.
4. Compute SHA-1 of `text` (UTF-8 bytes). Return the raw 20 bytes.

**HashQualifiedSubsection(node: parsenode.Node, qualifier: string) -> optional raw bytes (20)**

1. Let `normalized_qualifier` = `NormalizeText(qualifier)`.
2. Find the subsection in `node.public.subsections` whose `heading` equals `normalized_qualifier`.
3. If not found, return absent.
4. Let `text` = `FormatSection(subsection.raw_heading, subsection.content)`.
5. Compute SHA-1 of `text` (UTF-8 bytes). Return the raw 20 bytes.

**HashAgentSection(node: parsenode.Node) -> optional raw bytes (20)**

1. If `node.agent` is absent, return absent.
2. If `node.agent.content` is empty (after blank-line removal) and `node.agent.subsections` is empty,
   return absent.
3. Let `text` = `FormatSection(node.agent.raw_heading, node.agent.content)`.
4. For each subsection in `node.agent.subsections`:
   a. Let `sub_block` = `FormatSection(subsection.raw_heading, subsection.content)`.
   b. If `sub_block` is not empty, append `\n` then `sub_block` to `text`.
5. Compute SHA-1 of `text` (UTF-8 bytes). Return the raw 20 bytes.

**HashFileContent(file_path: pathutils.PathCfs, neutralize_artifact_tag: boolean) -> raw bytes (20)**

1. Call `FileOpen(file_path, mode="read", timeout_ms=30000)`.
   If `FileOpen` raises `FileUnreadable`, raise error `"file unreadable"`.
2. Let `lines` = empty list.
3. Loop:
   a. Call `FileReadLine(handle)`.
   b. If `EndOfFile` is raised, exit loop.
   c. If `neutralize_artifact_tag` is true, apply neutralization to the line:
      - If the line matches the pattern
        `code-from-spec: <anything>@<27 base64url characters>`,
        replace the 27-character hash portion with
        `"---------------------------"`.
      - The rest of the line (including the logical name) is unchanged.
   d. Append the line followed by `"\n"` to `lines`.
4. Call `FileClose(handle)`.
   (Call `FileClose` in error paths too before re-raising.)
5. Let `text` = concatenation of all strings in `lines`.
6. Compute SHA-1 of `text` (UTF-8 bytes). Return the raw 20 bytes.

---

## Main Function

**ChainHashCompute(chain: chainresolver.Chain) -> string**

**Step 1 — Collect content hashes**

Let `hashes` = empty list of raw byte sequences (each 20 bytes).

1. For each `ancestor` in `chain.ancestors` (from root to target's parent):
   a. Call `NodeParse(ancestor.unqualified_logical_name)`.
      If it fails, raise error `"parse failure"`.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, append `h` to `hashes`.

2. For each `dep` in `chain.dependencies` (already sorted alphabetically by logical name):
   a. If `LogicalNameIsArtifact(dep.unqualified_logical_name)` is true:
      - Let `h` = `HashFileContent(dep.file_path, neutralize_artifact_tag=true)`.
      - Append `h` to `hashes`.
   b. Else if `LogicalNameIsExternal(dep.unqualified_logical_name)` is true:
      - Let `h` = `HashFileContent(dep.file_path, neutralize_artifact_tag=false)`.
      - Append `h` to `hashes`.
   c. Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` is true:
      - Call `NodeParse(dep.unqualified_logical_name)`.
        If it fails, raise error `"parse failure"`.
      - If `dep.qualifier` is absent:
        - Let `h` = `HashPublicSubsections(node)`.
        - If `h` is present, append `h` to `hashes`.
      - If `dep.qualifier` is present:
        - Let `h` = `HashQualifiedSubsection(node, dep.qualifier)`.
        - If `h` is present, append `h` to `hashes`.

3. Target `# Public`:
   a. Call `NodeParse(chain.target.unqualified_logical_name)`.
      If it fails, raise error `"parse failure"`.
   b. Let `h` = `HashPublicSubsections(node)`.
   c. If `h` is present, append `h` to `hashes`.
   d. Save this `node` result as `target_node`.

4. Target `# Agent`:
   a. Let `h` = `HashAgentSection(target_node)`.
   b. If `h` is present, append `h` to `hashes`.

5. If `chain.input` is present:
   a. Let `input` = `chain.input`.
   b. If `LogicalNameIsArtifact(input.unqualified_logical_name)` is true:
      - Let `h` = `HashFileContent(input.file_path, neutralize_artifact_tag=true)`.
      - Append `h` to `hashes`.
   c. Else if `LogicalNameIsExternal(input.unqualified_logical_name)` is true:
      - Let `h` = `HashFileContent(input.file_path, neutralize_artifact_tag=false)`.
      - Append `h` to `hashes`.

**Step 2 — Compute final hash**

1. Let `concatenated` = concatenation of all byte sequences in `hashes` (20 bytes each, in order).
2. Compute SHA-1 of `concatenated`.
3. Encode the resulting 20 bytes as base64url (RFC 4648 §5, no padding) — producing 27 characters.
4. Return the 27-character string.
