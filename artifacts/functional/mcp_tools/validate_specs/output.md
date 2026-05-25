<!-- code-from-spec: ROOT/functional/mcp_tools/validate_specs@PENDING -->

## Data structures

```
record FormatError
  node: string
  message: string

record StalenessEntry
  node: string
  artifact_path: string
  status: string ("missing" or "stale")

record ValidationReport
  format_errors: list of FormatError
  circular_references: list of list of string
  staleness: list of StalenessEntry
```

## Functions

### function ValidateSpecs() -> ValidationReport

Validates the entire spec tree for format errors, circular
references, and artifact staleness. Takes no parameters.

**Step 1 -- Discover nodes**

1. Use node_discovery to find all _node.md files under the
   "code-from-spec/" directory.

2. For each discovered file, derive the node's logical name
   using logical_names reverse resolution.

**Step 2 -- Parse all nodes**

3. For each discovered node:
   a. Use frontmatter to parse the YAML frontmatter.
   b. Use node_parsing to parse the body into sections.
   c. Cache the parsed result (frontmatter and sections) so
      each node is parsed only once and reused by later steps.
   d. If a file cannot be read, raise error "unreadable file".
   e. If a file has invalid structure, raise error "parse failure".

**Step 3 -- Format validation**

4. For each discovered node, use format_validation to check it
   against the structural rules. Format validation uses:
   - logical_names to verify that depends_on targets resolve.
   - name_normalization to compare headings with logical names
     derived from filesystem paths.
   - path_validation to verify that outputs paths are safe.

5. Collect all FormatError entries into a list.

**Step 4 -- Ranking and cycle detection**

6. Use node_ranking to rank all nodes and artifacts and detect
   circular references. Pass the full set of discovered nodes
   with their parsed frontmatter.

7. The ranking assigns each node and artifact a numeric rank
   that determines processing order for staleness resolution
   (lower rank first).

8. If circular references are detected, record the cycle
   participants as a list of lists of logical name strings.

**Step 5 -- Staleness detection**

9. Collect all nodes that have an outputs field. Sort them by
   rank (lowest rank first).

10. For each such node, in rank order:
    a. Compute the chain hash using the same algorithm as
       load_chain: SHA-1 of concatenated per-position raw
       hashes, then base64url encoded (no padding), producing
       a 27-character string.
    b. For each output declared in the node's outputs:
       - Use artifact_tag to extract the hash from the generated
         file at the output path.
       - If the file does not exist, add a StalenessEntry with
         status "missing".
       - If the file exists but has no artifact tag, add a
         StalenessEntry with status "missing".
       - If the file exists and has an artifact tag but the hash
         does not match the computed chain hash, add a
         StalenessEntry with status "stale".
       - If the hash matches, do not add an entry (skip).

**Output assembly**

11. Build the ValidationReport:
    - format_errors: all FormatError entries from step 5.
    - circular_references: all cycles from step 8.
    - staleness: all StalenessEntry entries from step 10,
      ordered by rank (lowest first).

12. Return the ValidationReport.
