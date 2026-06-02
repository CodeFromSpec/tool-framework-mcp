<!-- code-from-spec: ROOT/functional/logic/mcp_tools/validate_specs@6KRmZ0BNke1rfsHrp3e_16fL5J4 -->

# Public

## Namespace

    namespace: mcpvalidatespecs

## Interface

```
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

function MCPValidateSpecs() -> ValidationReport
```

MCPValidateSpecs scans the full spec tree, validates
format, detects cycles, and checks staleness. Always
returns a report — never raises an error.

StalenessEntry.status is one of:
- "missing" — the artifact file does not exist.
- "stale" — the artifact file exists but the hash
  does not match.
- "malformed tag" — the artifact file exists but has
  no artifact tag or the tag cannot be parsed.

Entries where the hash matches are not included in
staleness.

StalenessEntry.rank is the rank from NodeRankCompute.
Entries with equal rank have no dependency between them
and may be processed in parallel.

cycles is a flat list of logical names involved in
non-convergence during ranking.

### Step 1 — Discover nodes

1. Call SpecTreeScan().
   If it raises an error, return a ValidationReport with:
   - format_errors = [FormatError(node="", rule="scan",
     detail=<error message>)]
   - cycles = []
   - staleness = []

### Step 2 — Parse all nodes

2. For each discovered node in the scan result:
   a. Call FrontmatterParse with the node's file_path.
   b. Call NodeParse with the node's logical_name.
   c. If either call raises an error, record a
      FormatError(node=logical_name, rule="parse",
      detail=<error message>) and exclude this node
      from subsequent steps.
   d. Otherwise, cache (frontmatter, node) for this
      logical name.

### Step 3 — Format validation

3. Build a list of SpecTreeValidateInput from successfully
   parsed nodes:
   - logical_name: the node's logical name
   - frontmatter: the cached Frontmatter
   - node: the cached Node

4. Call SpecTreeValidate(entries).
   Collect all returned FormatError entries.

### Step 4 — Ranking and cycle detection

5. If Step 2 or Step 3 produced any format errors,
   skip this step (ranking results would be unreliable).

6. Otherwise, build a list of NodeRankInput from
   successfully parsed nodes:
   - logical_name: the node's logical name
   - frontmatter: the cached Frontmatter

7. Call NodeRankCompute(entries).

8. If NodeRankCompute raises UnresolvableReference:
   Record a FormatError(node="", rule="ranking",
   detail=<error message>).
   Staleness entries will fall back to alphabetical
   order by node logical name.

9. Otherwise, store the ranked entries and cycle
   participants.

### Step 5 — Staleness detection

10. For each successfully parsed node whose frontmatter
    has a non-empty output field, ordered by rank
    ascending then logical name ascending (or
    alphabetical by logical name if no ranking
    available):

    a. Call ChainResolve(logical_name).
       If it raises an error, record StalenessEntry(
         node=logical_name,
         artifact_path=frontmatter.output,
         status="missing",
         detail=<error message>,
         rank=<node rank or 0>)
       and continue to next node.

    b. Call ChainHashCompute(chain).
       If it raises an error, record StalenessEntry(
         node=logical_name,
         artifact_path=frontmatter.output,
         status="missing",
         detail=<error message>,
         rank=<node rank or 0>)
       and continue to next node.

    c. Construct a PathCfs from frontmatter.output.
       Call ArtifactTagExtract with the path.

       If FileUnreadable:
         Record StalenessEntry(status="missing",
           detail=<reason>).
       If NoTagFound or MalformedTag:
         Record StalenessEntry(status="malformed tag",
           detail=<reason>).
       If tag hash does not match chain hash:
         Record StalenessEntry(status="stale",
           detail="file hash <tag.hash> does not match
           expected hash <chain_hash>").
       If tag hash matches chain hash:
         Skip (not included in staleness).

       Set artifact_path = frontmatter.output.
       Set rank = node's rank from Step 4, or 0 if
       no ranking available.

### Step 6 — Assemble report

11. Return ValidationReport with:
    - format_errors: all FormatErrors from Steps 2, 3, 4.
    - cycles: cycle participant logical names from Step 4,
      or empty list if ranking was skipped.
    - staleness: all StalenessEntries from Step 5,
      ordered by rank ascending then node logical name
      ascending.
