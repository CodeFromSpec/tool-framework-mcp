---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
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

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpvalidatespecs"`

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
- `"missing"` — file does not exist on disk, or no
  manifest entry exists for this node.
- `"stale"` — chain hash in the manifest does not
  match the current chain hash.
- `"modified"` — checksum in the manifest does not
  match the hash of the file on disk.
- `"orphan"` — manifest entry exists but no
  corresponding node in the spec tree.

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
   "code-from-spec/" using ListAllFiles or equivalent.
   Store as all_dirs for use in Step 3.

### Step 2 — Parse all nodes

3. For each discovered LogicalName (`ln`):
     a. Call `FrontmatterParse(CfsPath(ln.Path))`.
        If it fails, add FormatError(node=ln.Name,
        rule="parse", detail=<error message>) to
        format_errors. Mark node as parse-failed.
        Continue to next node.
     b. Call `NodeParse(ln.Name)`. If it fails,
        add FormatError(node=ln.Name,
        rule="parse", detail=<error message>) to
        format_errors. Mark node as parse-failed.
        Continue to next node.
     c. Cache (ln, frontmatter, parsed_node) keyed by
        ln.Name.

### Step 3 — Format validation

4. Build a list of SpecTreeValidateInput from
   successfully parsed nodes: each entry =
   (ln.Name, frontmatter, parsed_node).
   Call `SpecTreeValidate(entries, all_dirs)`. Append
   all returned FormatError entries to format_errors.

### Step 4 — Ranking and cycle detection

5. If format_errors is non-empty (from Steps 2 or 3):
     Skip ranking step. ranked_entries = empty.
     cycles = [].
   Else:
     Build a list of NodeRankInput from successfully
     parsed nodes: each entry =
     (ln.Name, ln.Parent, frontmatter).
     Call `NodeRankCompute(entries)`.
     If NodeRankCompute returns UnresolvableReference
     error:
       Append FormatError(node="", rule="ranking",
       detail=<error message>) to format_errors.
       ranked_entries = empty. cycles = [].
     Else:
       Store ranked_entries and cycles from the result.

### Step 5 — Read manifest

6. Call `ManifestOpen("read")`. If it fails, treat
   as empty manifest (no entries). Store the result
   as `manifest_handle`.

### Step 6 — Staleness detection

7. Determine processing order for staleness checks:
   If ranked_entries is non-empty:
     Order nodes by rank ascending, then by
     logical_name ascending within equal rank.
   Else:
     Order nodes alphabetically by logical_name.

   For each node that has a non-empty output in its
   frontmatter, in the above order:

     a. Derive the artifact logical name: strip "SPEC/"
        prefix from node.logical_name and prepend
        "ARTIFACT/".

     b. Call `ChainResolve(node.logical_name)`. If it
        fails: Append StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     c. Call `ChainHashCompute(chain)` using the result
        from step (b). If it fails: Append
        StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     d. Look up the artifact logical name in
        manifest_handle.Entries.

        If no entry exists: Append StalenessEntry with
        status="missing", detail="no manifest entry".

        If entry exists:
          Compare entry.ChainHash with computed chain
          hash. If they differ: Append StalenessEntry
          with status="stale", detail="manifest chain
          hash <entry.ChainHash> does not match
          expected hash <chain hash>".

          If chain hashes match: check the file on
          disk. Construct CfsPath from
          frontmatter.output. Call
          `OpenFile(path, "read", 30000)`. If it
          fails (file does not exist): Append
          StalenessEntry with status="missing".
          Else: read the full file content, compute
          its SHA-1 hash (base64url, 27 chars).
          Call `handle.Close()`. Compare with
          entry.Checksum. If they differ: Append
          StalenessEntry with status="modified",
          detail="file checksum does not match
          manifest checksum".

        If chain hash matches and checksum matches:
        skip (artifact is up to date).

        Set artifact_path from frontmatter.output.
        Set rank from the node's rank (from Step 4,
        or 0 if no ranking available).

### Step 7 — Orphan detection

8. For each entry in manifest_handle.Entries:
     Derive the generating node's logical name: strip
     "ARTIFACT/" prefix and prepend "SPEC/".
     If no successfully parsed node has that logical
     name, or if the node's frontmatter.output is
     empty: Append StalenessEntry(
       node=entry key (artifact logical name),
       artifact_path=entry.Path,
       status="orphan",
       detail="manifest entry has no corresponding
       spec node",
       rank=0).

### Step 8 — Assemble report

9. Return ValidationReport with:
     format_errors = all FormatError entries from
       Steps 2, 3, 4
     cycles = cycle list from Step 4 (empty list if
       ranking was skipped or no cycles)
     staleness = all StalenessEntry entries from
       Steps 6 and 7, ordered by rank ascending then
       node logical_name ascending

## Go-specific guidance

- Use the `spectree` package for `SpecTreeScan`.
- Use the `logicalnames` package for `LogicalName`.
- Use the `spectreevalidate` package for
  `SpecTreeValidate` and `SpecTreeValidateInput`,
  `FormatError`.
- Use the `noderanking` package for `NodeRankCompute`,
  `NodeRankInput`, `NodeRankEntry`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `ManifestOpen`,
  `ManifestHandle`, `ManifestEntry`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `parsenode` package for `NodeParse`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, `ValidateCfsPath`, and
  `CfsPath`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- The package name should be `mcpvalidatespecs`.
- `StalenessEntry`, `ValidationReport` are exported
  structs.
- The function never returns an error — all problems
  are collected in the report.
