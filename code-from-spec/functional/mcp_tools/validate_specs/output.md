<!-- code-from-spec: ROOT/functional/mcp_tools/validate_specs@mpsHW5elMtFnavuako9qwGSt0iQ -->

# validate_specs

## Data structures

```
record FormatError
  node: string
  rule: string
  detail: string

record StalenessEntry
  node: string
  artifact_path: string
  status: string        -- "missing" or "stale"

record ValidationReport
  format_errors: list of FormatError
  circular_references: list of list of string
  staleness: list of StalenessEntry
```

## Functions

---

### ValidateSpecs() -> ValidationReport

No parameters. Scans the entire spec tree starting from `code-from-spec/`.

Errors:
- `"unreadable file"`: a spec node file cannot be read.
- `"parse failure"`: a spec node file has invalid structure.

---

**Step 1 — Discover nodes**

1. Call `DiscoverNodes()`.
   If `DiscoverNodes` returns a "directory not found" error,
     raise error `"unreadable file: code-from-spec/ does not exist"`.
   If `DiscoverNodes` returns a "walk error",
     raise error `"unreadable file: <error detail>"`.
   If `DiscoverNodes` returns "no nodes found",
     raise error `"unreadable file: no _node.md files found in code-from-spec/"`.

2. The result is a list of DiscoveredNode records, each with:
   - `logical_name`: derived via `ReverseResolve(file_path)`
   - `file_path`: absolute or project-relative path to the `_node.md` file

---

**Step 2 — Parse all nodes**

3. Initialize an empty cache: a map from logical_name -> parsed node record.
   Each parsed node record holds:
   - `frontmatter`: the Frontmatter from `ParseFrontmatter`
   - `parsed_body`: the ParsedNode from `ParseNode`

4. For each discovered node:
   a. Call `ParseFrontmatter(file_path)`.
      If it returns "file unreadable", raise error `"unreadable file: <file_path>"`.
      If it returns "malformed YAML", raise error `"parse failure: <file_path>: malformed YAML"`.
   b. Call `ParseNode(logical_name)`.
      If it returns any error, raise error `"parse failure: <logical_name>: <error detail>"`.
   c. Store the results in the cache under `logical_name`.

---

**Step 3 — Format validation**

5. Call `ValidateFormat(discovered_nodes)`, passing the full list from Step 1.
   If `ValidateFormat` returns "unreadable node" for any node,
     raise error `"unreadable file: <node path>"`.

6. Collect all returned FormatError records into `collected_format_errors`.

---

**Step 4 — Ranking and cycle detection**

7. Call `DetectCycles(nodes_with_frontmatter)`, where
   `nodes_with_frontmatter` is the full set of discovered nodes
   paired with their parsed frontmatter from the cache.

8. If `DetectCycles` returns an "unresolvable reference" error:
   a. Append a FormatError to `collected_format_errors`:
      - node = the node that contains the bad reference
      - rule = `"depends_on"`
      - detail = the error message from `DetectCycles`
   b. Set `ranked_entries` to an empty list.
      (Staleness entries will fall back to alphabetical order.)
   c. Set `cycle_participants` to an empty list.
   Else:
   a. Use the returned `ranked_entries` (list of RankedEntry: logical_name + rank).
   b. Use the returned `cycle_participants` (list of logical names in cycles).
      If non-empty, convert into groups of cycle participants for the report.
      Each group is a list of strings (logical names forming a cycle).

---

**Step 5 — Staleness detection**

9. Determine the ordered list of nodes to check for staleness:
   - If `ranked_entries` is non-empty:
       Sort nodes that have `outputs` by their rank (ascending, lowest first).
   - Else (ranking failed):
       Sort nodes that have `outputs` alphabetically by logical_name.

10. Initialize `staleness_entries` as an empty list.

11. For each node in the staleness order:
    a. Retrieve the node's `frontmatter` from the cache.
       If `frontmatter.outputs` is empty, skip this node.
    b. Compute the chain hash for this node:
       - Use the same algorithm as `load_chain`:
         SHA-1 of the concatenated position hashes of all nodes in the chain,
         then base64url-encode the result.
    c. For each output in `frontmatter.outputs`:
       i.  Let `artifact_path` = `output.path`.
       ii. Validate `artifact_path` using `ValidatePath(artifact_path, project_root)`.
           If validation fails, skip this output (path is unsafe to read).
       iii. Call `ExtractArtifactTag(artifact_path)`.
            If `ExtractArtifactTag` returns "file unreadable":
              Append StalenessEntry:
                node = logical_name
                artifact_path = artifact_path
                status = `"missing"`
              Continue to next output.
            If `ExtractArtifactTag` returns "no tag found":
              Append StalenessEntry:
                node = logical_name
                artifact_path = artifact_path
                status = `"missing"`
              Continue to next output.
            If `ExtractArtifactTag` returns "malformed tag":
              Append StalenessEntry:
                node = logical_name
                artifact_path = artifact_path
                status = `"missing"`
              Continue to next output.
       iv. Compare `extracted_tag.hash` with the computed chain hash.
           If they differ:
             Append StalenessEntry:
               node = logical_name
               artifact_path = artifact_path
               status = `"stale"`
           If they match:
             Do not add an entry for this output.

---

**Step 6 — Assemble and return the report**

12. Assemble the ValidationReport:
    - `format_errors` = `collected_format_errors` (all FormatError records from Step 3 and Step 4)
    - `circular_references` = groups of cycle participant logical names (from Step 4),
      or empty list if no cycles
    - `staleness` = `staleness_entries` ordered by rank (lowest first),
      or alphabetically if ranking was unavailable

13. Return the ValidationReport.

---

## Contracts

- All errors are collected before returning — validation does not stop at the first error.
- Staleness check only runs for nodes whose frontmatter has a non-empty `outputs` field.
- Nodes that fail format validation in Step 3 are still included in staleness checking
  in Step 5 where their outputs can be safely read.
- Each node is parsed exactly once (Step 2 cache), and results are reused in all
  subsequent steps.
- The staleness entries in the report are ordered by dependency rank (lowest first)
  so that the caller can resolve them in the correct order.
```
