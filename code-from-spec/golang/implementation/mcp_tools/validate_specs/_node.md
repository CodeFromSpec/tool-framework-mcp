---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/artifact_tag
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/parsing/node_parsing
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/spec_tree/validate
  - SPEC/golang/implementation/utils/node_ranking
output: internal/mcpvalidatespecs/mcpvalidatespecs.go
---

# SPEC/golang/implementation/mcp_tools/validate_specs

Validates the spec tree for format errors, circular
references, and artifact staleness.

# Public

## Package

`package mcpvalidatespecs`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpvalidatespecs"`

## Interface

```go
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
}

type ValidationReport struct {
	FormatErrors []spectreevalidate.FormatError
	Cycles       []string
	Staleness    []StalenessEntry
}

func MCPValidateSpecs() ValidationReport
```

No parameters. Scans the entire spec tree starting from
`code-from-spec/`. Always returns a report — never
returns an error. Problems are collected in the report.

`StalenessEntry.Status` is one of:
- `"missing"` — file does not exist.
- `"stale"` — hash mismatch.
- `"malformed tag"` — file exists but has no artifact
  tag or the tag cannot be parsed.

`StalenessEntry.Rank` is the rank from `NodeRankCompute`.

# Agent

Implement the validate specs tool as a Go package.

## Logic

### Step 1 — Discover nodes

1. Call `SpecTreeScan()` to discover all spec nodes.
   If SpecTreeScan fails: return ValidationReport with
     format_errors = [ FormatError(node="", rule="scan",
     detail=<error message>) ],
     cycles = [], staleness = [].

2. Discover all subdirectory paths under
   "code-from-spec/" using ListFiles or equivalent.
   Store as all_dirs for use in Step 3.

### Step 2 — Parse all nodes

3. For each discovered SpecTreeNode:
     a. Call `FrontmatterParse(node.file_path)`. If it
        fails, add FormatError(node=node.logical_name,
        rule="parse", detail=<error message>) to
        format_errors. Mark node as parse-failed.
        Continue to next node.
     b. Call `NodeParse(node.logical_name)`. If it fails,
        add FormatError(node=node.logical_name,
        rule="parse", detail=<error message>) to
        format_errors. Mark node as parse-failed.
        Continue to next node.
     c. Cache (frontmatter, parsed_node) keyed by
        logical_name.

### Step 3 — Format validation

4. Build a list of SpecTreeValidateInput from
   successfully parsed nodes: each entry =
   (logical_name, frontmatter, parsed_node).
   Call `SpecTreeValidate(entries, all_dirs)`. Append
   all returned FormatError entries to format_errors.

### Step 4 — Ranking and cycle detection

5. If format_errors is non-empty (from Steps 2 or 3):
     Skip ranking step. ranked_entries = empty.
     cycles = [].
   Else:
     Build a list of NodeRankInput from successfully
     parsed nodes: each entry =
     (logical_name, frontmatter).
     Call `NodeRankCompute(entries)`.
     If NodeRankCompute returns UnresolvableReference
     error:
       Append FormatError(node="", rule="ranking",
       detail=<error message>) to format_errors.
       ranked_entries = empty. cycles = [].
     Else:
       Store ranked_entries and cycles from the result.

### Step 5 — Staleness detection

6. Determine processing order for staleness checks:
   If ranked_entries is non-empty:
     Order nodes by rank ascending, then by
     logical_name ascending within equal rank.
   Else:
     Order nodes alphabetically by logical_name.

   For each node that has a non-empty output in its
   frontmatter, in the above order:

     a. Call `ChainResolve(node.logical_name)`. If it
        fails: Append StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     b. Call `ChainHashCompute(chain)` using the result
        from step (a). If it fails: Append
        StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     c. Construct PathCfs from frontmatter.output. Call
        `ArtifactTagExtract(path)`.

        If error is FileUnreadable: Append
        StalenessEntry with status="missing".

        If error is NoTagFound or MalformedTag: Append
        StalenessEntry with status="malformed tag".

        If tag is successfully extracted and tag.hash
        does not match chain hash: Append
        StalenessEntry with status="stale", detail=
        "file hash <tag.hash> does not match expected
        hash <chain hash>".

        If tag.hash matches chain hash: skip (not
        included).

        Set artifact_path from frontmatter.output.
        Set rank from the node's rank (from Step 4,
        or 0 if no ranking available).

### Step 6 — Assemble report

7. Return ValidationReport with:
     format_errors = all FormatError entries from
       Steps 2, 3, 4
     cycles = cycle list from Step 4 (empty list if
       ranking was skipped or no cycles)
     staleness = all StalenessEntry entries from
       Step 5, ordered by rank ascending then
       node logical_name ascending

## Go-specific guidance

- Use the `spectree` package for `SpecTreeScan`.
- Use the `spectreevalidate` package for
  `SpecTreeValidate` and `SpecTreeValidateInput`,
  `FormatError`.
- Use the `noderanking` package for `NodeRankCompute`,
  `NodeRankInput`, `NodeRankEntry`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `artifacttag` package for `ArtifactTagExtract`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `parsenode` package for `NodeParse`.
- Use the `pathutils` package for `PathValidateCfs`,
  `PathCfs`.
- The package name should be `mcpvalidatespecs`.
- `StalenessEntry`, `ValidationReport` are exported
  structs.
- The function never returns an error — all problems
  are collected in the report.
