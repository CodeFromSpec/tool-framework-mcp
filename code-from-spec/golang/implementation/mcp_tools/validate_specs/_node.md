---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/spec_tree/scan
  - SPEC/golang/implementation/spec_tree/validate
  - SPEC/golang/implementation/spec_tree/ranking
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

1. Call `spectree.SpecTreeScan()` to discover all spec nodes.
   If SpecTreeScan fails: return ValidationReport with
     format_errors = [ FormatError(node="", rule="scan",
     detail=<error message>) ],
     cycles = [], staleness = [].

2. Discover all subdirectory paths under
   "code-from-spec/" using ListAllFiles or equivalent.
   Store as all_dirs for use in Step 3.

### Step 2 — Parse all nodes

3. For each discovered CfsReference (`ref`):
     a. Call `parsing.ParseNode(ref.LogicalName)`.
        If it fails, add FormatError(
        node=ref.LogicalName, rule="parse",
        detail=<error message>) to format_errors.
        Mark node as parse-failed. Continue to next
        node.
     b. Cache node keyed by node.Reference.LogicalName.

### Step 3 — Format validation

4. Collect successfully parsed nodes into a list.
   Call `spectreevalidate.SpecTreeValidate(nodes, all_dirs)`. Append
   all returned FormatError entries to format_errors.

### Step 4 — Ranking and cycle detection

5. If format_errors is non-empty (from Steps 2 or 3):
     Skip ranking step. ranked_entries = empty.
     cycles = [].
   Else:
     Call `noderanking.NodeRankCompute(nodes)` with the successfully
     parsed nodes.
     If NodeRankCompute returns UnresolvableReference
     error:
       Append FormatError(node="", rule="ranking",
       detail=<error message>) to format_errors.
       ranked_entries = empty. cycles = [].
     Else:
       Store ranked_entries and cycles from the result.

### Step 5 — Read manifest

6. Call `manifest.OpenManifest(true)`. If it fails,
   treat as empty manifest (no entries). Store the
   result as `m`.

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

     b. Call `chainresolver.ChainResolve(node.logical_name)`. If it
        fails: Append StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     c. Call `chainhash.ChainHashCompute(chain)` using the result
        from step (b). If it fails: Append
        StalenessEntry(
          node=node.logical_name,
          artifact_path=frontmatter.output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     d. Look up the artifact logical name in
        m.Entries.

        If no entry exists: Append StalenessEntry with
        status="missing", detail="no manifest entry".

        If entry exists:
          Compare entry.ChainHash with computed chain
          hash. If they differ: Append StalenessEntry
          with status="stale", detail="manifest chain
          hash <entry.ChainHash> does not match
          expected hash <chain hash>".

          If chain hashes match: check the file on
          disk. Construct oslayer.CfsPath from
          frontmatter.output. Call
          `oslayer.OpenFile(path, "read", 30000)`. If it
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

8. For each entry in m.Entries:
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
- Use the `parsing` package for `ParseNode`,
  `CfsReference`, `NodeFrontmatter`, `Node`.
- Use the `spectreevalidate` package for
  `SpecTreeValidate` and
  `FormatError`.
- Use the `noderanking` package for `NodeRankCompute`
  and `NodeRankEntry`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, `ValidateStringIsCfsPath`, and
  `CfsPath`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- The package name should be `mcpvalidatespecs`.
- `StalenessEntry`, `ValidationReport` are exported
  structs.
- The function never returns an error — all problems
  are collected in the report.

# Private

## TODO

### Empty directories not detected by missing_node_md

`collectAllDirs` derives directory paths from files
returned by `ListAllFiles`. Empty subdirectories (no
files at all) are invisible — `missing_node_md` cannot
detect them. A real directory walk is needed to fix
this. Low priority: empty directories without
`_node.md` are uncommon in practice.
