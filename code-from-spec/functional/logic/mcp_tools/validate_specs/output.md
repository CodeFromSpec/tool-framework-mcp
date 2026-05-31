<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@oXnoZBtPgby3_SDCXrccUwD2EDw -->

# MCPValidateSpecs

## Records

```
record StalenessEntry
  node:          string
  output_id:     string
  artifact_path: string
  status:        string    -- "missing" | "stale" | "malformed tag"
  detail:        string
  rank:          integer

record ValidationReport
  format_errors: list of FormatError
  cycles:        list of string
  staleness:     list of StalenessEntry
```

## Function

```
function MCPValidateSpecs() -> ValidationReport
```

No parameters. Scans the entire spec tree starting from `code-from-spec/`.
Always returns a report — never raises an error. All problems are collected
in the report.

---

### Step 1 — Discover nodes

1. Call `SpecTreeScan()`.
2. If `SpecTreeScan` fails:
   - Construct a single FormatError:
     - node   = ""
     - rule   = "scan"
     - detail = error message from the failure
   - Return a ValidationReport with:
     - format_errors = [that FormatError]
     - cycles        = []
     - staleness     = []

---

### Step 2 — Parse all nodes

Initialize:
- `parsed_nodes`  = empty list of records { logical_name, file_path, frontmatter, node }
- `format_errors` = empty list of FormatError

For each entry in the list returned by `SpecTreeScan`:

1. Call `FrontmatterParse(entry.file_path)`.
   If it fails:
   - Append a FormatError:
     - node   = entry.logical_name
     - rule   = "parse"
     - detail = error message
   - Skip this node (do not add to `parsed_nodes`).
   - Continue to next entry.

2. Call `NodeParse(entry.logical_name)`.
   If it fails:
   - Append a FormatError:
     - node   = entry.logical_name
     - rule   = "parse"
     - detail = error message
   - Skip this node (do not add to `parsed_nodes`).
   - Continue to next entry.

3. Append a record to `parsed_nodes`:
   - logical_name = entry.logical_name
   - file_path    = entry.file_path
   - frontmatter  = result of FrontmatterParse
   - node         = result of NodeParse

---

### Step 3 — Format validation

1. Build a list of `SpecTreeValidateInput` from `parsed_nodes`:
   - For each record in `parsed_nodes`:
     - logical_name = record.logical_name
     - frontmatter  = record.frontmatter
     - node         = record.node

2. Call `SpecTreeValidate(entries)`.

3. Append every FormatError returned by `SpecTreeValidate` to `format_errors`.

---

### Step 4 — Ranking and cycle detection

Initialize:
- `ranked_entries` = empty list   -- will hold NodeRankEntry records
- `cycles`         = []
- `ranking_failed` = false

If `format_errors` is non-empty (from Steps 2 or 3):
- Skip this step entirely.
- Leave `ranked_entries` empty and `cycles` empty.

Otherwise:

1. Build a list of `NodeRankInput` from `parsed_nodes`:
   - For each record in `parsed_nodes`:
     - logical_name = record.logical_name
     - frontmatter  = record.frontmatter

2. Call `NodeRankCompute(entries)`.

3. If `NodeRankCompute` returns an UnresolvableReference error:
   - Append a FormatError:
     - node   = ""
     - rule   = "ranking"
     - detail = error message
   - Set `ranking_failed` = true
   - Leave `ranked_entries` empty, `cycles` empty.

4. Otherwise:
   - Store the returned `ranked` list in `ranked_entries`.
   - Store the returned `cycles` list in `cycles`.

---

### Step 5 — Staleness detection

Initialize:
- `staleness_entries` = empty list of StalenessEntry

Determine processing order:

- If `ranked_entries` is non-empty:
  - Filter `parsed_nodes` to only those nodes that have at least one
    entry in `frontmatter.outputs`.
  - Sort this subset by rank ascending (using the rank from
    `ranked_entries` keyed on logical_name), then by logical_name
    ascending as a tiebreaker.
  - For each node not found in `ranked_entries`, use rank = 0.

- Else (ranking skipped or failed):
  - Filter `parsed_nodes` to only those nodes that have at least one
    entry in `frontmatter.outputs`.
  - Sort this subset alphabetically by logical_name ascending.

For each node record in the ordered subset:

  a. Determine `node_rank`:
     - Look up the node's logical_name in `ranked_entries`.
     - If found, use that entry's rank.
     - If not found, use 0.

  b. Call `ChainResolve(record.logical_name)`.
     If it fails:
     - For each output in `record.frontmatter.outputs`:
       - Append a StalenessEntry:
         - node          = record.logical_name
         - output_id     = output.id
         - artifact_path = output.path
         - status        = "missing"
         - detail        = error message from ChainResolve failure
         - rank          = node_rank
     - Continue to next node.

  c. Call `ChainHashCompute(chain)` with the resolved chain.
     If it fails:
     - For each output in `record.frontmatter.outputs`:
       - Append a StalenessEntry:
         - node          = record.logical_name
         - output_id     = output.id
         - artifact_path = output.path
         - status        = "missing"
         - detail        = error message from ChainHashCompute failure
         - rank          = node_rank
     - Continue to next node.

  d. Store `expected_hash` = result of ChainHashCompute.

  e. For each output in `record.frontmatter.outputs`:

     1. Construct a PathCfs from output.path.

     2. Call `ArtifactTagExtract(path)`.

     3. If `ArtifactTagExtract` returns a FileUnreadable error:
        - Append a StalenessEntry:
          - node          = record.logical_name
          - output_id     = output.id
          - artifact_path = output.path
          - status        = "missing"
          - detail        = error message
          - rank          = node_rank
        - Continue to next output.

     4. If `ArtifactTagExtract` returns a NoTagFound or MalformedTag error:
        - Append a StalenessEntry:
          - node          = record.logical_name
          - output_id     = output.id
          - artifact_path = output.path
          - status        = "malformed tag"
          - detail        = error message
          - rank          = node_rank
        - Continue to next output.

     5. Compare the extracted tag's hash to `expected_hash`.
        - If they differ:
          - Append a StalenessEntry:
            - node          = record.logical_name
            - output_id     = output.id
            - artifact_path = output.path
            - status        = "stale"
            - detail        = "file hash <extracted hash>, expected <expected_hash>"
            - rank          = node_rank
        - If they match:
          - Do not add any StalenessEntry (skip this output).

---

### Step 6 — Assemble report

Sort `staleness_entries` by rank ascending, then by node logical_name
ascending as a tiebreaker.

Return ValidationReport:
- format_errors = `format_errors` (all FormatErrors from Steps 2, 3, 4)
- cycles        = `cycles` (empty list if ranking was skipped or failed)
- staleness     = sorted `staleness_entries`

---

## Contracts and invariants

- Always returns a ValidationReport — never raises an error to the caller.
- Collects all problems; does not stop at the first error.
- Staleness checking runs for all nodes with outputs, even when format
  errors exist for other nodes. Nodes that failed to parse are excluded
  from staleness checking (their parse errors are already reported).
- Each node file is parsed at most once (results are cached in `parsed_nodes`
  and reused across Steps 3, 4, and 5).
- `StalenessEntry.status` is always one of: `"missing"`, `"stale"`,
  `"malformed tag"`.
- Entries where the artifact hash matches the chain hash are not included
  in the staleness list.
- `StalenessEntry.rank` reflects the node's computed rank, enabling the
  caller to identify which staleness entries can be regenerated in parallel.
