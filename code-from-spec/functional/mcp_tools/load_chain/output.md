<!-- code-from-spec: ROOT/functional/mcp_tools/load_chain@_lo5s-OsOb8GCKwZfrXpMIWRbYs -->

# LoadChain

## Data Structures

```
record ChainResult
  chain_hash: string           -- 27-character base64url SHA-1 hash
  context:    string           -- all chain content concatenated
  input:      optional string  -- content of the input artifact, if declared
```

## Function Signature

```
function LoadChain(logical_name) -> ChainResult
  errors:
    - "invalid logical name": logical_name is not a recognized ROOT/ reference.
    - "no outputs": target node has no outputs field in its frontmatter.
    - "invalid output path": an output path fails path validation.
    - "chain resolution failure": a dependency cannot be resolved.
    - "unreadable file": a file in the chain cannot be read or parsed.
```

## Step-by-Step Logic

### Phase 1 — Validation

1. Validate the logical name.
   Call ResolvePath(logical_name).
   If it raises "unsupported reference" or any error,
     raise error "invalid logical name".

2. Parse the target node's frontmatter.
   Call ParseFrontmatter(ResolvePath(logical_name)).
   If parsing fails, raise error "unreadable file".

3. Check that the frontmatter has an `outputs` field with at least one entry.
   If `outputs` is absent or empty, raise error "no outputs".

4. Validate each output path.
   For each output in frontmatter.outputs:
     Call ValidatePath(output.path, project_root).
     If validation fails, raise error "invalid output path".

### Phase 2 — Assemble the Context Stream

Initialize:
  - `context_parts` as an empty list of strings
  - `hash_inputs`   as an empty list of raw SHA-1 byte sequences

The context stream is built by appending to `context_parts`.
Each piece that contributes to the hash is also appended to `hash_inputs`
(see "Hash contribution" notes below).

---

**Step 1 — Ancestors (root down to target's direct parent)**

Collect all ancestor logical names of the target node.
  Start from ROOT and walk down to the target's direct parent.
  (Use GetParent repeatedly to build the ancestor list, then reverse it
   so iteration goes root-first.)

For each ancestor in root-to-parent order:
  1. Parse the ancestor node: call ParseNode(ancestor_logical_name).
     If parsing fails, raise error "unreadable file".

  2. If the parsed node has no `# Public` section, or the section content
     is empty after stripping whitespace, skip this ancestor entirely.

  3. Otherwise:
     Append the `# Public` section content (without the "# Public" heading)
       to `context_parts`.
     Compute SHA-1 of the full `# Public` section content
       (with the "# Public" heading included).
     Append the raw SHA-1 bytes to `hash_inputs`.

---

**Step 2 — Dependencies (depends_on)**

Retrieve `depends_on` from the target node's frontmatter.
Sort entries alphabetically by their logical name string.

For each entry in sorted order:

  Case A: entry is a ROOT/ reference without a qualifier
    -- e.g. ROOT/x/y
    1. Parse the referenced node: call ParseNode(entry).
       If parsing fails, raise error "chain resolution failure".
    2. If the parsed node has no `# Public` section or it is empty,
       skip this entry.
    3. Append the `# Public` content (without heading) to `context_parts`.
    4. Compute SHA-1 of the full `# Public` content (with heading).
       Append raw bytes to `hash_inputs`.

  Case B: entry is a ROOT/ reference with a qualifier
    -- e.g. ROOT/x/y(z)
    1. Extract the qualifier using ExtractQualifier(entry).
    2. Parse the referenced node using the path without the qualifier.
       Call ParseNode(entry).
       If parsing fails, raise error "chain resolution failure".
    3. Find the subsection inside `# Public` whose heading normalizes
       to the qualifier (use NormalizeName to compare).
       If no matching subsection is found, raise error "chain resolution failure".
    4. Append the subsection content (without the "## <qualifier>" heading)
       to `context_parts`.
    5. Compute SHA-1 of the subsection content (with its heading included).
       Append raw bytes to `hash_inputs`.

  Case C: entry is an ARTIFACT/ reference with a qualifier
    -- e.g. ARTIFACT/x/y(id)
    1. Call ResolveArtifactReference(entry) to get node_path and artifact_id.
       If it fails, raise error "chain resolution failure".
    2. Parse the frontmatter of the node at node_path.
       Find the output whose id matches artifact_id.
       If not found, raise error "chain resolution failure".
    3. Read the full content of the artifact file at that output path.
       If unreadable, raise error "unreadable file".
    4. Strip any frontmatter (content between the opening --- and closing ---
       delimiters, inclusive) from the artifact content.
    5. Append the stripped content to `context_parts`.
    6. Compute SHA-1 of the stripped content.
       Append raw bytes to `hash_inputs`.

---

**Step 3 — External files (external)**

Retrieve `external` from the target node's frontmatter.
Sort entries alphabetically by their `path` field.

For each entry in sorted order:

  Case A: no fragments declared (entry.fragments is absent or empty)
    1. Read the full file at entry.path.
       If unreadable, raise error "unreadable file".
    2. Append the full content to `context_parts`.
    3. Compute SHA-1 of the full content.
       Append raw bytes to `hash_inputs`.

  Case B: fragments declared
    1. Open a FileReader for entry.path using OpenFileReader(entry.path).
       If unreadable, raise error "unreadable file".
    2. Initialize `fragment_content` as empty string.
    3. For each fragment in entry.fragments (in declaration order):
         a. The fragment specifies a `lines` range (e.g. "10-20") and
            optionally a `description` and `hash`.
         b. Parse the start and end line numbers from fragment.lines.
         c. Read lines from the FileReader up to and including the end line.
            Collect only lines within [start, end].
         d. Append the collected lines to `fragment_content`.
    4. Append `fragment_content` to `context_parts`.
    5. Compute SHA-1 of `fragment_content`.
       Append raw bytes to `hash_inputs`.

---

**Step 4 — Target's reduced frontmatter and `# Public` section**

1. Parse the target node: call ParseNode(logical_name).
   If parsing fails, raise error "unreadable file".

2. Build the reduced frontmatter block:
   Construct a YAML block containing only the `outputs` field,
   wrapped in --- delimiters:
     ---
     outputs:
       - id: <output.id>
         path: <output.path>
       ...
     ---

3. Append the reduced frontmatter block to `context_parts`.
   (The reduced frontmatter is NOT included in `hash_inputs`.)

4. If the target has a `# Public` section and it is not empty:
     Append the `# Public` content (without heading) to `context_parts`.
     Compute SHA-1 of the full `# Public` content (with heading).
     Append raw bytes to `hash_inputs`.

---

**Step 5 — Target's `# Agent` section**

1. If the target node has a `# Agent` section and it is not empty:
     Append the `# Agent` content (without heading) to `context_parts`.
     Compute SHA-1 of the full `# Agent` content (with heading).
     Append raw bytes to `hash_inputs`.
   Otherwise, skip.

---

### Phase 3 — Input artifact (if declared)

1. Check whether the target node's frontmatter has an `input` field
   with a non-empty value.

2. If the `input` field is present:
     a. Resolve the input reference (it is an ARTIFACT/ logical name).
        Call ResolveArtifactReference(frontmatter.input).
        If it fails, raise error "chain resolution failure".
     b. Parse the frontmatter of that node to locate the artifact file path.
     c. Read the artifact file content.
        If unreadable, raise error "unreadable file".
     d. Strip any frontmatter block from the content.
     e. Compute SHA-1 of the stripped content.
        Append raw bytes to `hash_inputs`.
     f. Store the stripped content as `input_content`.
        (It is NOT appended to `context_parts`.)

3. If no `input` field, `input_content` remains absent.

---

### Phase 4 — Compute the chain hash

1. Concatenate all raw SHA-1 byte sequences in `hash_inputs`
   (in the order they were appended) into a single byte sequence.

2. Compute the SHA-1 of that concatenated byte sequence.

3. Encode the resulting 20-byte SHA-1 as base64url
   (RFC 4648 §5, no padding), yielding exactly 27 characters.

4. This is the `chain_hash`.

---

### Phase 5 — Assemble and return the result

1. Join all strings in `context_parts` into a single `context` string
   by concatenation (no separator).

2. Build the ChainResult:
   - chain_hash: the 27-character hash computed above.
   - context:    the concatenated context string.
   - input:      `input_content` if present, otherwise absent.

3. Return the ChainResult.

---

## Error Conditions

| Error                      | Condition                                                              |
|----------------------------|------------------------------------------------------------------------|
| "invalid logical name"     | logical_name does not start with ROOT/ or fails ResolvePath.           |
| "no outputs"               | Target node frontmatter has no outputs field or it is empty.           |
| "invalid output path"      | Any output path fails ValidatePath.                                    |
| "chain resolution failure" | A depends_on entry cannot be resolved or a required subsection/artifact is missing. |
| "unreadable file"          | Any file in the chain cannot be opened, read, or parsed.               |

## Contracts

- Returns all chain content in one call — no pagination.
- If any file in the chain is unreadable, returns an error — no partial results.
- The context stream contains no metadata or structural markers — only spec content.
- The reduced frontmatter block (Step 4) appears in the context stream but does
  NOT contribute to the chain hash.
- Section headings (# Public, # Agent) are stripped from the context stream
  but ARE included in the hashed content.
- Tools are stateless — each call resolves its own inputs independently.
