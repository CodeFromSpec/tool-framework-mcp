<!-- code-from-spec: ROOT/functional/mcp_tools/load_chain@PENDING -->

## Data structures

```
record Output
  id: string
  path: string

record Frontmatter
  depends_on: optional list of strings
  external: optional list of external records
  input: optional string
  outputs: optional list of Output

record ExternalEntry
  path: string
  fragments: optional list of line range strings
```

## Functions

### function LoadChain(logical_name) -> list of text items

Loads the complete spec chain for a node and returns the chain
hash, context stream, and (optionally) input as separate items.

**Validation**

1. Validate logical_name as a ROOT/ reference using logical_names.
   If invalid, raise error "invalid logical name".

2. Resolve logical_name to a file path using logical_names.
   Read the frontmatter of that file using frontmatter parser.
   If the file cannot be read, raise error "unreadable file".

3. If frontmatter has no outputs field, raise error "no outputs".

4. For each output in outputs:
   Validate the output path using path_validation.
   If any path fails, raise error "invalid output path".

**Context stream assembly**

Initialize an empty list called hash_parts.
Initialize an empty string called context.

**Step 1 -- Ancestors (root to target's parent)**

5. Determine the ancestor chain from the root node down to the
   target's direct parent, in tree depth order.

6. For each ancestor node, from root toward the target's parent:
   a. Read the ancestor file using file_reader.
   b. Parse the file using node_parsing.
   c. Extract the "# Public" section content.
   d. If "# Public" is absent or empty, skip this ancestor entirely.
   e. Compute SHA-1 of the "# Public" content including the
      "# Public" heading. Append the raw 20-byte hash to hash_parts.
   f. Append the "# Public" content (without the heading) to context.

**Step 2 -- Dependencies (depends_on)**

7. If the target has a depends_on field, sort entries alphabetically
   by logical name.

8. For each depends_on entry, in alphabetical order:
   a. If the entry is "ROOT/x/y" (plain node reference):
      - Resolve to a file path using logical_names.
      - Read the file using file_reader and parse using node_parsing.
      - Extract the "# Public" section content.
      - Compute SHA-1 of the content. Append raw hash to hash_parts.
      - Append the content (without the heading) to context.
   b. If the entry is "ROOT/x/y(z)" (subsection reference):
      - Resolve "ROOT/x/y" to a file path.
      - Read and parse the file.
      - Extract the "## z" subsection within "# Public".
      - Compute SHA-1 of the subsection content.
        Append raw hash to hash_parts.
      - Append the subsection content (without the heading) to context.
   c. If the entry is "ARTIFACT/x/y(id)" (artifact reference):
      - Resolve to the artifact file path.
      - Read the full content using file_reader, excluding frontmatter.
      - Compute SHA-1 of the content. Append raw hash to hash_parts.
      - Append the content to context.
   d. If any dependency cannot be resolved,
      raise error "chain resolution failure".
   e. If any file cannot be read,
      raise error "unreadable file".

**Step 3 -- External files**

9. If the target has an external field, sort entries alphabetically
   by path.

10. For each external entry, in alphabetical order by path:
    a. Read the file at the declared path using file_reader.
       If unreadable, raise error "unreadable file".
    b. If the entry has no fragments declared:
       - Use the full file content.
    c. If the entry has fragments declared:
       - For each fragment line range, in declaration order,
         extract the content at that range.
       - Concatenate the extracted fragments.
    d. Compute SHA-1 of the resulting content.
       Append raw hash to hash_parts.
    e. Append the content to context.

**Step 4 -- Target "# Public"**

11. Parse the target node using node_parsing.
    Extract the "# Public" section content.

12. Build a reduced frontmatter block containing only the outputs field.

13. Compute SHA-1 of the "# Public" content including the heading.
    Append raw hash to hash_parts.

14. Prepend the reduced frontmatter block to the "# Public" content
    (without the heading). Append the combined text to context.

**Step 5 -- Target "# Agent"**

15. Extract the "# Agent" section content from the target node.

16. If "# Agent" is present:
    a. Compute SHA-1 of the "# Agent" content including the heading.
       Append raw hash to hash_parts.
    b. Append the content (without the heading) to context.

**Input separation**

17. If the target has an input field:
    a. Resolve the input reference to an artifact file path.
    b. Read the artifact content using file_reader, excluding
       frontmatter.
    c. Compute SHA-1 of the input content.
       Append raw hash to hash_parts.
    d. Store the input content as a separate text item (do not
       append to context).

**Chain hash computation**

18. Concatenate all raw 20-byte hashes in hash_parts in order.

19. Compute SHA-1 of the concatenated bytes.

20. Encode the resulting 20-byte hash as base64url
    (RFC 4648 section 5, no padding) producing a 27-character string.

**Return**

21. Build the result as a list of text items:
    - Item 1: the 27-character chain hash.
    - Item 2: the context string.
    - Item 3 (only if input field exists): the input content.

22. Return the list.
