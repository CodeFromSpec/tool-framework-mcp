[//]: # (code-from-spec: ROOT/functional/mcp_tools/load_chain@O0l5bALQpKtvIfFfFmKd8nFNis8)

# load_chain

## Overview

`load_chain` is an MCP tool that loads and assembles the full specification chain
for a given node. It returns the chain hash on the first line, followed by the
concatenated content of all relevant spec files.

## Interface

```
tool load_chain
  input:
    logical_name: string
  output:
    text: string
  errors:
    - invalid logical name: cannot resolve the provided logical name.
    - unreadable file: a file in the chain cannot be read.
```

## Output Format

The returned text has the following structure:

```
chain_hash: <27-character base64url SHA-1 hash>

<concatenated spec content>

--- input ---
<input content, if any>
```

The first line is always `chain_hash: ` followed by the 27-character hash
produced by `ComputeChainHash`. A blank line follows, then the spec content.

If the node's frontmatter declares an `input` path, the content of that file
is appended after a `--- input ---` separator line. If no input is declared,
the `--- input ---` section is omitted.

## Behavior

### Chain Assembly

Given `logical_name`, the tool assembles the chain by walking from the node up
to the root, collecting spec content at each level, then appending any
`external` files declared in frontmatter, and finally including the node's own
content.

The chain is assembled in this order:

1. Ancestor nodes from root down to the immediate parent (each contributing
   their `# Public` section content, or the specific subsection if a qualifier
   is present in the logical name).
2. The target node's full `_node.md` content.
3. Any files declared under `external` in the target node's frontmatter, in
   declaration order.

### Qualifier Handling

When the logical name includes a parenthetical qualifier (e.g.
`ROOT/x/y(interface)`), the tool extracts only the matching `## interface`
subsection from `# Public` when collecting ancestor content. Without a
qualifier, the full `# Public` section is used.

For the target node itself, the full file content is always included regardless
of any qualifier.

### Ancestor Content Extraction

For each ancestor node (from `ROOT` down to the direct parent):

1. Parse the node using `ParseNode`.
2. If the logical name has a qualifier that matches a subsection heading
   (after normalization via `NormalizeName`), include only that subsection's
   content.
3. Otherwise, include the full content of the `# Public` section.
4. If the node has no `# Public` section, contribute nothing for that ancestor.

### External Files

External files are declared in the target node's frontmatter under `external`.
Each `External` record has a `path` and an optional list of `fragments`.

- If `fragments` is absent or empty, the entire file content is appended.
- If `fragments` is present, only the lines specified by each fragment are
  appended, in declaration order. Each `ExternalFragment` has:
  - `lines`: a line range in the format `"N-M"` (1-based, inclusive).
  - `description`: an optional label (used as a comment header if present).
  - `hash`: a hash of the expected fragment content, used for staleness
    detection (not used during assembly itself).

### Input Material

After all spec content is assembled, if the target node's frontmatter declares
a non-empty `input` path:

1. Validate the path using `ValidatePath`.
2. Read the file at that path.
3. Append the literal line `--- input ---` followed by the file's content.

### Hash Computation

The chain hash is computed by calling `ComputeChainHash(logical_name)`, which
returns a 27-character base64url-encoded SHA-1 hash. This hash covers the
assembled chain content and is placed on the first line of the response.

## Error Handling

- If `logical_name` cannot be resolved (e.g. malformed or unsupported prefix),
  return error: `invalid logical name`.
- If any file in the chain (node files, external files, or the input file)
  cannot be read, return error: `unreadable file`.
- Path validation errors for external or input paths are surfaced as
  `unreadable file`.
