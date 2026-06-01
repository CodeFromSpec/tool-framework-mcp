<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@djiquT-0z_dvf3TD00Fypr57ET4 -->

# validate_specs

namespace: mcpvalidatespecs

## Records

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

## Functions

function MCPValidateSpecs() -> ValidationReport

  No parameters. Scans the entire spec tree starting from "code-from-spec/".
  Always returns a ValidationReport — never raises an error.
  Problems are collected in the report.

  StalenessEntry.status is one of:
    - "missing"       — file does not exist.
    - "stale"         — hash mismatch.
    - "malformed tag" — file exists but has no artifact tag or the tag cannot be parsed.

  Entries where the hash matches are not included in staleness.

  StalenessEntry.rank is the rank from NodeRankCompute.
  Entries with equal rank have no dependency between them and can be processed in parallel.

  cycles is a flat list of logical names involved in non-convergence during ranking
  (as returned by NodeRankCompute).

  ### Step 1 — Discover nodes

  1. Call SpecTreeScan().
     If it fails:
       Create a single format_error with node = "", rule = "scan", detail = error message.
       Return ValidationReport with that single format_errors entry, empty cycles, empty staleness.

  ### Step 2 — Parse all nodes

  2. Initialize an empty cache of parsed results keyed by logical name.
     Initialize an empty list of format_errors (parse_errors).

  3. For each discovered node from Step 1:
     a. Call FrontmatterParse with the node's file_path.
        If it fails:
          Record a FormatError with node = logical_name, rule = "parse", detail = error message.
          Skip to the next discovered node.
     b. Call NodeParse with the node's logical_name.
        If it fails:
          Record a FormatError with node = logical_name, rule = "parse", detail = error message.
          Skip to the next discovered node.
     c. Store the (frontmatter, node) pair in the cache under the node's logical_name.

  ### Step 3 — Format validation

  4. Build a list of SpecTreeValidateInput from successfully cached nodes:
       For each entry in cache: logical_name + frontmatter + node.

  5. Call SpecTreeValidate(entries).
     Collect all returned FormatError entries into format_errors.

  ### Step 4 — Ranking and cycle detection

  6. If any format_errors exist (from Steps 2 or 3):
       Skip this step.
       ranked_entries = empty list.
       cycle_list = empty list.
     Else:
       a. Build a list of NodeRankInput from successfully cached nodes:
            For each entry in cache: logical_name + frontmatter.
       b. Call NodeRankCompute(entries).
          If NodeRankCompute returns an UnresolvableReference error:
            Record a FormatError with node = "", rule = "ranking", detail = error message.
            ranked_entries = empty list.
            cycle_list = empty list.
          Else:
            Store ranked_entries (list of NodeRankEntry with logical_name and rank).
            Store cycle_list (list of logical names involved in cycles).

  ### Step 5 — Staleness detection

  7. Build a lookup of rank by logical_name from ranked_entries.
     If ranked_entries is empty, rank lookup returns 0 for every node.

  8. Collect all nodes from the cache that have one or more outputs in their frontmatter.
     Sort them: first by rank ascending (from rank lookup), then by logical_name ascending.

  9. Initialize an empty list of staleness_entries.

  10. For each such node in sorted order:
      a. Call ChainResolve(logical_name) to get the resolved Chain.
         If it fails:
           For each output in frontmatter.outputs:
             Record StalenessEntry with:
               node         = logical_name
               output_id    = output.id
               artifact_path = output.path
               status       = "missing"
               detail       = error message from ChainResolve
               rank         = rank lookup for logical_name (0 if not found)
           Continue to next node.

      b. Call ChainHashCompute(chain) to get the current chain hash string.
         If it fails:
           For each output in frontmatter.outputs:
             Record StalenessEntry with:
               node         = logical_name
               output_id    = output.id
               artifact_path = output.path
               status       = "missing"
               detail       = error message from ChainHashCompute
               rank         = rank lookup for logical_name (0 if not found)
           Continue to next node.

      c. For each output in frontmatter.outputs:
           i.  Construct a PathCfs from output.path.
           ii. Call ArtifactTagExtract with the PathCfs.
               If FileUnreadable error:
                 Record StalenessEntry with:
                   node         = logical_name
                   output_id    = output.id
                   artifact_path = output.path
                   status       = "missing"
                   detail       = error message describing the file cannot be read
                   rank         = rank lookup for logical_name (0 if not found)
               If NoTagFound or MalformedTag error:
                 Record StalenessEntry with:
                   node         = logical_name
                   output_id    = output.id
                   artifact_path = output.path
                   status       = "malformed tag"
                   detail       = error message describing the tag issue
                   rank         = rank lookup for logical_name (0 if not found)
               If tag extracted successfully:
                 If tag.hash does not match chain hash:
                   Record StalenessEntry with:
                     node         = logical_name
                     output_id    = output.id
                     artifact_path = output.path
                     status       = "stale"
                     detail       = "file hash <tag.hash> does not match expected hash <chain hash>"
                     rank         = rank lookup for logical_name (0 if not found)
                 If tag.hash matches chain hash:
                   Skip — do not add an entry.

  ### Step 6 — Assemble report

  11. Sort staleness_entries by rank ascending, then by node logical_name ascending.

  12. Return ValidationReport with:
        format_errors = all FormatErrors collected in Steps 2, 3, and 4
        cycles        = cycle_list (empty list if no cycles or ranking was skipped)
        staleness     = staleness_entries sorted as above

## Contracts

  - Always returns a ValidationReport — never raises an error.
  - Reports all problems found — does not stop at the first.
  - Staleness check runs for all nodes with outputs, even if format errors exist for other nodes.
  - Staleness entries include rank for parallel processing guidance.
  - Each node is parsed at most once (cache reused across all steps).
