<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@4inVRvMKuYkOC4YdJkZH1vhsyRs -->

## Namespace

    namespace: mcpvalidatespecs

## Records

```
record StalenessEntry
  node: string
  output_id: string
  artifact_path: string
  status: string
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of spectreevalidate.FormatError
  cycles: list of string
  staleness: list of StalenessEntry
```

## Functions

```
function MCPValidateSpecs() -> ValidationReport
```

### MCPValidateSpecs

No parameters. Scans the entire spec tree starting from `code-from-spec/`.
Always returns a report — never raises an error. Problems are collected
in the report.

---

#### Step 1 — Discover nodes

1. Call `SpecTreeScan()`.
   If it fails, return a ValidationReport with:
   - format_errors = [ FormatError { node = "", rule = "scan", detail = error message } ]
   - cycles = []
   - staleness = []

2. Store the resulting list of SpecTreeNode entries.

---

#### Step 2 — Parse all nodes

1. Initialize:
   - parsed_nodes = empty map from logical_name to record { frontmatter, node }
   - parse_errors = empty list of FormatError

2. For each SpecTreeNode in the discovered list:
   a. Call `FrontmatterParse(node.file_path)`.
      If it fails, append FormatError { node = node.logical_name, rule = "parse", detail = error message }
      to parse_errors, and skip to the next node.
   b. Call `NodeParse(node.logical_name)`.
      If it fails, append FormatError { node = node.logical_name, rule = "parse", detail = error message }
      to parse_errors, and skip to the next node.
   c. Store { frontmatter, node } in parsed_nodes keyed by node.logical_name.

---

#### Step 3 — Format validation

1. Build a list of SpecTreeValidateInput from parsed_nodes:
   - For each entry in parsed_nodes: { logical_name, frontmatter, node }

2. Call `SpecTreeValidate(entries)`.
   Append all returned FormatError entries to a format_errors list.

3. Collect all errors: format_errors = parse_errors + SpecTreeValidate errors.

---

#### Step 4 — Ranking and cycle detection

1. If format_errors (from Steps 2 and 3 combined) is non-empty, skip ranking.
   Set ranked_entries = empty map.
   Set cycles = [].
   Set ranking_available = false.

2. Otherwise:
   a. Build a list of NodeRankInput from parsed_nodes:
      - For each entry: { logical_name, frontmatter }
   b. Call `NodeRankCompute(entries)`.
      If it returns an UnresolvableReference error:
      - Append FormatError { node = "", rule = "ranking", detail = error message } to format_errors.
      - Set ranked_entries = empty map.
      - Set cycles = [].
      - Set ranking_available = false.
      Otherwise:
      - Store ranked entries in ranked_entries as a map from logical_name to rank integer.
      - Store cycle participant logical names in cycles.
      - Set ranking_available = true.

---

#### Step 5 — Staleness detection

1. Build the ordered work list of nodes that have non-empty `outputs` in their frontmatter.
   If ranking_available:
   - Sort by rank ascending, then by logical_name ascending for ties.
   Otherwise:
   - Sort by logical_name ascending.

2. Initialize staleness = empty list of StalenessEntry.

3. For each node_logical_name in the work list:
   a. Retrieve frontmatter from parsed_nodes.
   b. Determine rank:
      - If ranking_available and node_logical_name is in ranked_entries, use that rank.
      - Otherwise use 0.
   c. Call `ChainResolve(node_logical_name)`.
      If it fails:
      - For each output in frontmatter.outputs, append StalenessEntry:
          node = node_logical_name
          output_id = output.id
          artifact_path = output.path
          status = "missing"
          detail = error message
          rank = rank
      - Continue to next node.
   d. Call `ChainHashCompute(chain)`.
      If it fails:
      - For each output in frontmatter.outputs, append StalenessEntry:
          node = node_logical_name
          output_id = output.id
          artifact_path = output.path
          status = "missing"
          detail = error message
          rank = rank
      - Continue to next node.
   e. For each output in frontmatter.outputs:
      - Construct a PathCfs from output.path.
      - Call `ArtifactTagExtract(path)`.
        If FileUnreadable:
          Append StalenessEntry { node = node_logical_name, output_id = output.id,
            artifact_path = output.path, status = "missing",
            detail = error message, rank = rank }.
        If NoTagFound or MalformedTag:
          Append StalenessEntry { node = node_logical_name, output_id = output.id,
            artifact_path = output.path, status = "malformed tag",
            detail = error message, rank = rank }.
        If tag is extracted successfully and tag.hash does not equal chain hash:
          Append StalenessEntry { node = node_logical_name, output_id = output.id,
            artifact_path = output.path, status = "stale",
            detail = "file hash: <tag.hash>, expected: <chain hash>", rank = rank }.
        If tag is extracted successfully and tag.hash equals chain hash:
          Skip — do not add an entry.

---

#### Step 6 — Assemble report

1. Sort staleness by rank ascending, then by node logical_name ascending.

2. Return ValidationReport:
   - format_errors: all FormatErrors from Steps 2, 3, and 4
   - cycles: cycle participant logical names from Step 4 (empty list if skipped)
   - staleness: sorted staleness list from Step 5
