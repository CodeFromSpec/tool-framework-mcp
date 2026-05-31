<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@jNJiLPGSDiDr_RxnQ3oLokFJvm4 -->

# MCPValidateSpecs

## Records

```
record StalenessEntry
  node: string
  output_id: string
  artifact_path: string
  status: string        -- "missing", "stale", or "malformed tag"
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of FormatError
  cycles: list of string
  staleness: list of StalenessEntry
```

## Function

```
function MCPValidateSpecs() -> ValidationReport
```

No parameters. Scans the entire spec tree starting from `code-from-spec/`.
Always returns a report — never raises an error. Problems are collected
in the report.

---

### Step 1 — Discover nodes

1. Call `SpecTreeScan()`.
   If it fails:
     - Return a `ValidationReport` with:
         format_errors = [ FormatError(node = "", rule = "scan", detail = <error message>) ]
         cycles = []
         staleness = []

2. Store the returned list of `SpecTreeNode` entries as <all_nodes>.

---

### Step 2 — Parse all nodes

Initialize:
- <parsed_nodes>    = empty map from logical_name to (frontmatter, node)
- <format_errors>   = empty list of FormatError

For each entry in <all_nodes>:

  1. Call `FrontmatterParse(entry.file_path)`.
     If it fails:
       - Append FormatError(node = entry.logical_name, rule = "parse", detail = <error message>)
         to <format_errors>.
       - Skip this entry (do not add to <parsed_nodes>).
       - Continue to next entry.
     Store result as <frontmatter>.

  2. Call `NodeParse(entry.logical_name)`.
     If it fails:
       - Append FormatError(node = entry.logical_name, rule = "parse", detail = <error message>)
         to <format_errors>.
       - Skip this entry (do not add to <parsed_nodes>).
       - Continue to next entry.
     Store result as <node>.

  3. Add entry to <parsed_nodes>:
       key = entry.logical_name
       value = (frontmatter = <frontmatter>, node = <node>)

---

### Step 3 — Format validation

1. Build a list of `SpecTreeValidateInput` from all entries in <parsed_nodes>:
     For each (logical_name, (frontmatter, node)) in <parsed_nodes>:
       - Create SpecTreeValidateInput(logical_name = logical_name, frontmatter = frontmatter, node = node)

2. Call `SpecTreeValidate(entries)`.
   Append all returned FormatError entries to <format_errors>.

---

### Step 4 — Ranking and cycle detection

Initialize:
- <rank_map>        = empty map from logical_name to integer
- <cycles>          = empty list of string

If <format_errors> is non-empty:
  - Skip this step entirely.
  - (Ranking results would be unreliable — leave <rank_map> empty and <cycles> empty.)

Otherwise:

  1. Build a list of `NodeRankInput` from all entries in <parsed_nodes>:
       For each (logical_name, (frontmatter, _)) in <parsed_nodes>:
         - Create NodeRankInput(logical_name = logical_name, frontmatter = frontmatter)

  2. Call `NodeRankCompute(entries)`.

     If it returns an UnresolvableReference error:
       - Append FormatError(node = "", rule = "ranking", detail = <error message>)
         to <format_errors>.
       - Leave <rank_map> empty and <cycles> empty.
       - (Staleness entries will fall back to alphabetical order by node logical name.)

     Otherwise:
       - For each entry in the returned <ranked> list:
           Store <rank_map>[entry.logical_name] = entry.rank
       - Set <cycles> = the returned <cycles> list.

---

### Step 5 — Staleness detection

Initialize:
- <staleness_entries> = empty list of StalenessEntry

Identify nodes to check: all entries in <parsed_nodes> where
`frontmatter.outputs` is non-empty.

Determine processing order:
  - If <rank_map> is non-empty:
      Sort by rank ascending, then by logical_name ascending for ties.
  - Otherwise:
      Sort alphabetically by logical_name ascending.

For each such node in the determined order:

  Let <logical_name> = the node's logical name.
  Let <outputs>      = frontmatter.outputs for this node.
  Let <node_rank>    = <rank_map>[<logical_name>] if present, else 0.

  1. Call `ChainResolve(<logical_name>)`.
     If it fails:
       - For each output in <outputs>:
           Append StalenessEntry(
             node          = <logical_name>,
             output_id     = output.id,
             artifact_path = output.path,
             status        = "missing",
             detail        = <error message>,
             rank          = <node_rank>
           ) to <staleness_entries>.
       - Continue to next node.
     Store result as <chain>.

  2. Call `ChainHashCompute(<chain>)`.
     If it fails:
       - For each output in <outputs>:
           Append StalenessEntry(
             node          = <logical_name>,
             output_id     = output.id,
             artifact_path = output.path,
             status        = "missing",
             detail        = <error message>,
             rank          = <node_rank>
           ) to <staleness_entries>.
       - Continue to next node.
     Store result as <expected_hash>.

  3. For each output in <outputs>:

     a. Construct a PathCfs from output.path.

     b. Call `ArtifactTagExtract(<path_cfs>)`.

        If error is FileUnreadable:
          - Append StalenessEntry(
              node          = <logical_name>,
              output_id     = output.id,
              artifact_path = output.path,
              status        = "missing",
              detail        = <error message>,
              rank          = <node_rank>
            ) to <staleness_entries>.
          - Continue to next output.

        If error is NoTagFound or MalformedTag:
          - Append StalenessEntry(
              node          = <logical_name>,
              output_id     = output.id,
              artifact_path = output.path,
              status        = "malformed tag",
              detail        = <error message>,
              rank          = <node_rank>
            ) to <staleness_entries>.
          - Continue to next output.

        Otherwise (tag extracted successfully):
          - Let <file_hash> = tag.hash.
          - If <file_hash> does not equal <expected_hash>:
              Append StalenessEntry(
                node          = <logical_name>,
                output_id     = output.id,
                artifact_path = output.path,
                status        = "stale",
                detail        = "file hash <file_hash> does not match expected hash <expected_hash>",
                rank          = <node_rank>
              ) to <staleness_entries>.
          - If <file_hash> equals <expected_hash>:
              Skip — do not add an entry.

---

### Step 6 — Assemble and return report

Sort <staleness_entries> by rank ascending, then by node logical name ascending.

Return ValidationReport(
  format_errors = <format_errors>,
  cycles        = <cycles>,
  staleness     = <staleness_entries>
)

---

## Contracts

- Always returns a report — never raises an error to the caller.
- Reports all problems found — does not stop at the first.
- Staleness check runs for all nodes with outputs, even if format
  errors exist for other nodes.
- Staleness entries include rank for parallel processing.
- Each node is parsed at most once (results cached from Step 2).
