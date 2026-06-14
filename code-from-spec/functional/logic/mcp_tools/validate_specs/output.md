<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@m_Bm5o_RC9rRTHpectAm9OMjECI -->

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

  1. Call SpecTreeScan() to discover all _node.md files.
     If SpecTreeScan fails:
       Return a ValidationReport with:
         format_errors = [ FormatError(node="", rule="scan", detail=<error message>) ]
         cycles = []
         staleness = []

  2. Discover all subdirectory paths under "code-from-spec/" using ListFiles or equivalent.
     Store as all_dirs for use in Step 3.

  3. For each SpecTreeNode in the scan result:
       a. Call FrontmatterParse(node.file_path).
          If parsing fails:
            Append FormatError(node=<logical_name>, rule="parse", detail=<error message>)
            to parse_errors.
            Mark this node as excluded.
            Continue to the next node.
       b. Call NodeParse(node.logical_name).
          If parsing fails:
            Append FormatError(node=<logical_name>, rule="parse", detail=<error message>)
            to parse_errors.
            Mark this node as excluded.
            Continue to the next node.
       c. Cache the (frontmatter, parsed_node) pair for this logical name.

  4. Build a list of SpecTreeValidateInput from successfully parsed nodes:
       Each entry has: logical_name, frontmatter (from cache), node (from cache).
     Call SpecTreeValidate(entries, all_dirs).
     Collect all returned FormatError entries into format_validation_errors.

  5. Set all_format_errors = parse_errors + format_validation_errors.
     Set skip_ranking = (all_format_errors is not empty).
     Set ranked_entries = [].
     Set cycles = [].
     Set ranking_error = absent.

     If skip_ranking is false:
       Build a list of NodeRankInput from successfully parsed nodes:
         Each entry has: logical_name, frontmatter (from cache).
       Call NodeRankCompute(entries).
       If NodeRankCompute returns an UnresolvableReference error:
         Append FormatError(node="", rule="ranking", detail=<error message>)
           to all_format_errors.
         Set ranking_error = present.
       Else:
         Set ranked_entries = the ranked entries returned.
         Set cycles = the cycles list returned.

  6. Build a lookup map from logical_name to rank using ranked_entries.
     If ranking_error is present or skip_ranking is true:
       The fallback ordering is alphabetical by node logical name.

     For each successfully parsed node that has a non-empty output in its frontmatter,
     ordered by rank ascending then logical name ascending
     (use rank from lookup map, or 0 if no mapping available; break ties by logical name):

       a. Call ChainResolve(logical_name).
          If it fails:
            Append StalenessEntry(
              node=<logical_name>,
              artifact_path=<frontmatter.output>,
              status="missing",
              detail=<error message>,
              rank=<rank from map, or 0 if absent>
            ).
            Continue to the next node.

       b. Call ChainHashCompute(chain) where chain is the result from step a.
          If it fails:
            Append StalenessEntry(
              node=<logical_name>,
              artifact_path=<frontmatter.output>,
              status="missing",
              detail=<error message>,
              rank=<rank from map, or 0 if absent>
            ).
            Continue to the next node.

       c. Construct PathCfs from frontmatter.output.
          Call ArtifactTagExtract(path).

          If ArtifactTagExtract returns FileUnreadable:
            Append StalenessEntry(
              node=<logical_name>,
              artifact_path=<frontmatter.output>,
              status="missing",
              detail=<error description>,
              rank=<rank>
            ).

          If ArtifactTagExtract returns NoTagFound or MalformedTag:
            Append StalenessEntry(
              node=<logical_name>,
              artifact_path=<frontmatter.output>,
              status="malformed tag",
              detail=<error description>,
              rank=<rank>
            ).

          If ArtifactTagExtract succeeds and tag.hash does not equal the computed chain hash:
            Append StalenessEntry(
              node=<logical_name>,
              artifact_path=<frontmatter.output>,
              status="stale",
              detail="file hash: <tag.hash>, expected: <chain_hash>",
              rank=<rank>
            ).

          If ArtifactTagExtract succeeds and tag.hash equals the computed chain hash:
            Skip — do not add a StalenessEntry.

  7. Sort all collected StalenessEntries by rank ascending, then by node ascending.

  8. Return ValidationReport with:
       format_errors = all_format_errors
       cycles = cycles
       staleness = sorted StalenessEntries from Step 7
