<!-- code-from-spec: SPEC/functional/logic/mcp_tools/validate_specs@Y7wJK5sbStvDmZq78IzOc9EKUw4 -->

namespace: mcpvalidatespecs

---

record StalenessEntry
  node: string
  artifact_path: string
  status: string
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of spectreevalidate.FormatError
  cycles: list of string
  staleness: list of StalenessEntry

---

function MCPValidateSpecs() -> ValidationReport

  1. Call SpecTreeScan() to discover all spec nodes.
     If SpecTreeScan fails:
       return ValidationReport with
         format_errors = [ FormatError(node="", rule="scan", detail=<error message>) ]
         cycles = []
         staleness = []

  2. Discover all subdirectory paths under "code-from-spec/" using ListFiles or equivalent.
     Store as all_dirs for use in Step 3.

  3. For each discovered SpecTreeNode:
       a. Call FrontmatterParse(node.file_path).
          If it fails, add FormatError(node=node.logical_name, rule="parse", detail=<error message>)
          to format_errors. Mark node as parse-failed. Continue to next node.
       b. Call NodeParse(node.logical_name).
          If it fails, add FormatError(node=node.logical_name, rule="parse", detail=<error message>)
          to format_errors. Mark node as parse-failed. Continue to next node.
       c. Cache (frontmatter, parsed_node) keyed by logical_name.

  4. Build a list of SpecTreeValidateInput from successfully parsed nodes:
       each entry = (logical_name, frontmatter, parsed_node)
     Call SpecTreeValidate(entries, all_dirs).
     Append all returned FormatError entries to format_errors.

  5. If format_errors is non-empty (from Steps 3 or 4):
       Skip ranking step. ranked_entries = empty. cycles = [].
     Else:
       Build a list of NodeRankInput from successfully parsed nodes:
         each entry = (logical_name, frontmatter)
       Call NodeRankCompute(entries).
       If NodeRankCompute returns UnresolvableReference error:
         Append FormatError(node="", rule="ranking", detail=<error message>) to format_errors.
         ranked_entries = empty.
         cycles = [].
       Else:
         Store ranked_entries and cycles from the result.

  6. Determine processing order for staleness checks:
     If ranked_entries is non-empty:
       Order nodes by rank ascending, then by logical_name ascending within equal rank.
     Else:
       Order nodes alphabetically by logical_name.

     For each node that has a non-empty output in its frontmatter, in the above order:

       a. Call ChainResolve(node.logical_name).
          If it fails:
            Append StalenessEntry(
              node=node.logical_name,
              artifact_path=frontmatter.output,
              status="missing",
              detail=<error message>,
              rank=<node rank from ranked_entries, or 0 if unavailable>
            ) to staleness.
            Continue to next node.

       b. Call ChainHashCompute(chain) using the result from step (a).
          If it fails:
            Append StalenessEntry(
              node=node.logical_name,
              artifact_path=frontmatter.output,
              status="missing",
              detail=<error message>,
              rank=<node rank from ranked_entries, or 0 if unavailable>
            ) to staleness.
            Continue to next node.

       c. Construct PathCfs from frontmatter.output.
          Call ArtifactTagExtract(path).

          If error is FileUnreadable:
            Append StalenessEntry(
              node=node.logical_name,
              artifact_path=frontmatter.output,
              status="missing",
              detail=<error message>,
              rank=<node rank>
            ) to staleness.

          If error is NoTagFound or MalformedTag:
            Append StalenessEntry(
              node=node.logical_name,
              artifact_path=frontmatter.output,
              status="malformed tag",
              detail=<error message>,
              rank=<node rank>
            ) to staleness.

          If tag is successfully extracted and tag.hash does not match chain hash:
            Append StalenessEntry(
              node=node.logical_name,
              artifact_path=frontmatter.output,
              status="stale",
              detail="file hash <tag.hash> does not match expected hash <chain hash>",
              rank=<node rank>
            ) to staleness.

          If tag is successfully extracted and tag.hash matches chain hash:
            Do not add an entry. Continue to next node.

  7. Return ValidationReport with:
       format_errors = all FormatError entries collected in Steps 3, 4, 5
       cycles = cycle list from Step 5 (empty list if ranking was skipped or no cycles)
       staleness = all StalenessEntry entries from Step 6,
                   ordered by rank ascending then node logical_name ascending
