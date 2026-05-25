<!-- code-from-spec: ROOT/functional/utils/format_validation@PENDING -->

## Data structures

```
record FormatError
  node: string
  rule: string
  detail: string
```

## Functions

### ValidateFormat(discovered_nodes) -> list of FormatError

1. Create an empty errors list.

2. Build a set of all known logical names from discovered_nodes.

3. For each discovered node, determine whether it has children:
   a node has children if any other discovered node's logical name
   starts with the node's logical name followed by "/".

4. For each discovered node:
   a. Open the file using file_reader.
      If unreadable, add a FormatError with rule "unreadable node"
      and continue to the next node.
   b. Parse frontmatter using frontmatter.
   c. Parse body using node_parsing.
   d. Run all validation rules below, collecting every error.

**Rule: Name verification**

5. The first heading in the parsed body (the name section heading)
   must match the logical name derived from the file path using
   logical_names reverse resolution.
   Comparison uses name_normalization.
   If they do not match, add a FormatError with rule
   "name verification" and a detail describing the mismatch.

**Rule: Frontmatter field restrictions**

6. If the node has children and the frontmatter contains any of
   the fields depends_on, external, input, or outputs, add a
   FormatError with rule "frontmatter field restrictions" for
   each offending field.

**Rule: Agent section restrictions**

7. If the node has children and the parsed body contains an
   agent section, add a FormatError with rule
   "agent section restrictions".

**Rule: Dependency targets**

8. For each entry in the frontmatter depends_on list:
   a. Resolve the entry to a file path using logical_names.
      If resolution fails, add a FormatError with rule
      "dependency targets" and detail "unresolvable".
   b. If the entry is an ancestor of the current node (the
      current node's logical name starts with the entry
      followed by "/"), add a FormatError with rule
      "dependency targets" and detail "ancestor dependency".
   c. If the entry is a descendant of the current node (the
      entry starts with the current node's logical name
      followed by "/"), add a FormatError with rule
      "dependency targets" and detail "descendant dependency".

**Rule: External file existence**

9. For each entry in the frontmatter external list:
   a. Check that the file at the entry's path exists.
      If not, add a FormatError with rule "external file existence".
   b. If the entry has fragments declared:
      - Open the file using file_reader.
      - Extract the declared line range.
      - Compute SHA-1 of the extracted content, encode as
        base64url.
      - Compare with the declared hash.
      - If they do not match, add a FormatError with rule
        "external file existence" and detail "hash mismatch".

**Rule: Output path validation**

10. For each entry in the frontmatter outputs list:
    a. Pass the entry's path through path_validation.
    b. If validation fails (traversal, absolute path, or outside
       project root), add a FormatError with rule
       "output path validation" and the validation detail.

**Rule: Duplicate public subsections**

11. If the parsed body has a public section, collect all level-2
    subsection headings within it.
    Normalize each heading using name_normalization.
    If any two normalized headings are equal, add a FormatError
    with rule "duplicate public subsections" for each duplicate.

12. Return the errors list.
