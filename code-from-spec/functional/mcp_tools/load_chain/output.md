<!-- code-from-spec: ROOT/functional/mcp_tools/load_chain@WqwpiA5knFV_2erXXhxVzoYflfE -->

# LoadChain

## Overview

Assembles all spec context needed to generate an artifact for a
given target node. Returns a chain hash, a continuous context
stream, and optionally a separate input artifact.


---


## Data structures

```
record HashAccumulator
  raw_hashes: list of byte arrays  -- each entry is 20 raw SHA-1 bytes

record ChainResult
  chain_hash:  string              -- 27-character base64url SHA-1
  context:     string              -- full concatenated context stream
  input:       optional string     -- input artifact content, if present
```


---


## Functions


### function LoadChain(logical_name) -> ChainResult

  Parameters:
    logical_name: string  -- the target node, must be a ROOT/ reference

  Returns:
    ChainResult record

  Errors:
    - "invalid logical name": logical_name is not a ROOT/ reference.
    - "no outputs": target node has no outputs field.
    - "invalid output path": an output path fails path validation.
    - "chain resolution failure": a dependency cannot be resolved.
    - "unreadable file": a file in the chain cannot be read or parsed.

  Steps:

  1. Validate the logical name.
     Call ResolvePath(logical_name).
     If ResolvePath raises "unsupported reference", raise error
     "invalid logical name".

  2. Parse the target node's frontmatter.
     Call ParseFrontmatter on the file path resolved in step 1.
     If it fails, raise error "unreadable file".
     If the parsed frontmatter has no outputs (empty list), raise
     error "no outputs".

  3. Validate all output paths.
     For each output in frontmatter.outputs:
       Call ValidatePath(output.path, project_root).
       If it fails, raise error "invalid output path".

  4. Initialize state.
     Set context_buffer to an empty string.
     Initialize a HashAccumulator with an empty raw_hashes list.

  5. Build the ancestor list.
     Collect all ancestors of logical_name from root down to the
     target's direct parent, in tree-depth order.
     (e.g. for ROOT/a/b/c the order is: ROOT, ROOT/a, ROOT/a/b)
     If the target is ROOT itself, the ancestor list is empty.

  6. Step 1 — Ancestors.
     For each ancestor in the ancestor list (root to direct parent):
       a. Call ParseNode(ancestor) to get its parsed sections.
          If ParseNode fails, raise error "unreadable file".
       b. If the node has no public section, or the public section
          content is empty, skip this ancestor entirely.
       c. Otherwise:
          - Append the public section content (without the "# Public"
            heading) to context_buffer.
          - Compute SHA-1 of the public section content INCLUDING the
            "# Public" heading (raw bytes, UTF-8 encoded).
          - Append the resulting 20-byte raw hash to
            accumulator.raw_hashes.

  7. Step 2 — Dependencies.
     Sort target frontmatter.depends_on alphabetically by logical name.
     For each dependency in sorted order:

       Case A: Dependency is a ROOT/ reference with no qualifier
         (e.g. ROOT/x/y):
           - Call ParseNode(dependency).
             If it fails, raise error "chain resolution failure".
           - If the node has no public section or the public section
             content is empty, skip.
           - Append the public section content (without heading) to
             context_buffer.
           - Compute SHA-1 of the full public section content
             INCLUDING its "# Public" heading.
           - Append raw hash to accumulator.raw_hashes.

       Case B: Dependency is a ROOT/ reference WITH a qualifier
         (e.g. ROOT/x/y(z)):
           - Call ExtractQualifier(dependency) to get qualifier z.
           - Call ParseNode on the base path (logical name stripped
             of the qualifier).
             If it fails, raise error "chain resolution failure".
           - Find the subsection within the public section whose
             normalized heading matches NormalizeName(z).
             If not found, raise error "chain resolution failure".
           - Append the subsection content (without the "## z"
             heading) to context_buffer.
           - Compute SHA-1 of the subsection content INCLUDING the
             "## z" heading.
           - Append raw hash to accumulator.raw_hashes.

       Case C: Dependency is an ARTIFACT/ reference
         (e.g. ARTIFACT/x/y(id)):
           - Call ResolveArtifactReference(dependency) to get
             ArtifactReference (node_path, artifact_id).
             If it fails, raise error "chain resolution failure".
           - Call ParseFrontmatter on node_path to get the node's
             frontmatter.
             If it fails, raise error "chain resolution failure".
           - Find the output in the frontmatter whose id matches
             artifact_id.
             If not found, raise error "chain resolution failure".
           - Open a FileReader for that output's file path.
             If it fails, raise error "unreadable file".
           - Read all lines. Skip any frontmatter block at the top
             (content between the first "---" and the closing "---").
             Collect the remaining lines as artifact_content.
           - Close the reader.
           - Append artifact_content to context_buffer.
           - Compute SHA-1 of artifact_content.
           - Append raw hash to accumulator.raw_hashes.

  8. Step 3 — External files.
     Sort target frontmatter.external alphabetically by path.
     For each external entry in sorted order:

       If external entry has no fragments declared:
         - Open a FileReader for external.path.
           If it fails, raise error "unreadable file".
         - Read all content into external_content.
         - Close the reader.
         - Append external_content to context_buffer.
         - Compute SHA-1 of external_content.
         - Append raw hash to accumulator.raw_hashes.

       If external entry has fragments declared:
         - Open a FileReader for external.path.
           If it fails, raise error "unreadable file".
         - Read all lines into a line array (1-based index).
         - Close the reader.
         - Initialize fragment_content as an empty string.
         - For each fragment in external.fragments (in declaration
           order):
             Parse fragment.lines as a line range (e.g. "10-20" or
             "42").
             Extract those lines from the line array.
             If the range is out of bounds, raise error
             "unreadable file".
             Append the extracted lines to fragment_content.
         - Append fragment_content to context_buffer.
         - Compute SHA-1 of fragment_content.
         - Append raw hash to accumulator.raw_hashes.

  9. Step 4 — Target's # Public section.
     Call ParseNode(logical_name).
     If it fails, raise error "unreadable file".

     Build a reduced frontmatter block:
       Format as YAML fenced with "---" lines containing only the
       outputs field from the target's frontmatter.
       Example:
         ---
         outputs:
           - id: <id>
             path: <path>
         ---

     Append reduced_frontmatter to context_buffer.
     (The reduced frontmatter is NOT included in the hash.)

     If the target has a public section and its content is not empty:
       - Append the public section content (without the "# Public"
         heading) to context_buffer.
       - Compute SHA-1 of the public section content INCLUDING the
         "# Public" heading.
       - Append raw hash to accumulator.raw_hashes.

  10. Step 5 — Target's # Agent section.
      If the target node has an agent section and its content is
      not empty:
        - Append the agent section content (without the "# Agent"
          heading) to context_buffer.
        - Compute SHA-1 of the agent section content INCLUDING the
          "# Agent" heading.
        - Append raw hash to accumulator.raw_hashes.

  11. Handle input artifact (if present).
      If the target frontmatter has a non-empty input field:
        - Call ResolveArtifactReference(frontmatter.input) to get
          ArtifactReference.
          If it fails, raise error "chain resolution failure".
        - Call ParseFrontmatter on the resolved node_path to find
          the output matching artifact_id.
          If not found, raise error "chain resolution failure".
        - Open a FileReader for that output's file path.
          If it fails, raise error "unreadable file".
        - Read all lines. Strip any frontmatter block at the top.
          Collect remaining lines as input_content.
        - Close the reader.
        - Compute SHA-1 of input_content.
        - Append raw hash to accumulator.raw_hashes.
        - Store input_content separately (do NOT append to
          context_buffer).

  12. Compute the final chain hash.
      Concatenate all entries in accumulator.raw_hashes in order
      into a single byte array (each entry is 20 bytes).
      Compute SHA-1 of that concatenated byte array.
      Encode the result using base64url (RFC 4648 §5, no padding).
      The result is exactly 27 characters.

  13. Assemble and return ChainResult:
      - chain_hash:  the 27-character base64url string from step 12.
      - context:     context_buffer.
      - input:       input_content from step 11, or absent if no
                     input field was present.


---


## Error conditions summary

| Error                    | Trigger                                                  |
|--------------------------|----------------------------------------------------------|
| "invalid logical name"   | logical_name is not a ROOT/ reference.                   |
| "no outputs"             | Target node frontmatter has no outputs list.             |
| "invalid output path"    | An output path fails ValidatePath.                       |
| "chain resolution failure" | A dependency or input cannot be resolved or found.     |
| "unreadable file"        | Any file in the chain cannot be opened, read, or parsed. |


---


## Contracts and invariants

- The function is stateless; it resolves all inputs independently
  on each call.
- If any file in the chain is unreadable, the function raises an
  error immediately — no partial results are returned.
- The context stream contains no metadata, headers, or structural
  markers — only raw spec content plus the reduced frontmatter block
  immediately before the target's public section.
- The chain hash includes section headings even though headings are
  stripped from the context stream.
- The input artifact content is returned as a separate item, not
  included in the context stream.
