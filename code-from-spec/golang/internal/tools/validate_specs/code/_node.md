---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/internal/artifact_tag
  - ROOT/golang/internal/file_reader
  - ROOT/golang/internal/format_validation
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/node_discovery
  - ROOT/golang/internal/node_ranking
  - ROOT/golang/internal/normalizename
  - ROOT/golang/internal/parsenode
  - ROOT/golang/internal/pathvalidation
outputs:
  - id: validate_specs
    path: internal/validate_specs/validate_specs.go
---

# ROOT/golang/internal/tools/validate_specs/code

Implementation of the validate_specs tool handler.

# Agent

## Implementation

1. Use `nodediscovery` to find all `_node.md` files in the
   spec tree.
2. For each discovered node, derive its logical name using
   `logicalnames` reverse resolution.
3. For each node, parse the YAML frontmatter using
   `frontmatter` and parse the body into sections using
   `parsenode`. Cache results so each node is parsed once.
4. Run `formatvalidation` on each node to check structural
   rules. This uses `logicalnames` to verify `depends_on`
   targets resolve, `normalizename` to compare headings with
   logical names, and `pathvalidation` to verify `outputs`
   paths are safe. Collect all format errors.
5. Use `noderanking` to rank all nodes and detect circular
   references. Pass the full set of discovered nodes with
   their parsed frontmatter. If cycles are detected, record
   the cycle participants.
6. For each node with `outputs`, in rank order (lowest first):
   a. Compute the chain hash using SHA-1 of concatenated
      position hashes, base64url encoded.
   b. For each output, use `artifacttag` to extract the hash
      from the generated file.
   c. Compare: file does not exist or has no artifact tag ->
      report `missing`. Hash mismatch -> report `stale`.
      Hash matches -> skip.
7. Assemble the validation report with all collected format
   errors, cycles, and staleness entries. Staleness entries
   are ordered by rank (lowest first).
8. Return the report as a success result.

### Error handling

- Unreadable spec file -> include in format errors, continue
  processing remaining nodes.
- Parse failure -> include in format errors, continue.
- All errors are collected and reported together; the handler
  does not stop at the first error.

## Constraints

- Reports all errors found -- does not stop at the first.
- Staleness check only runs for nodes that have `outputs`.
- Nodes that fail format validation are still checked for
  staleness where possible.
