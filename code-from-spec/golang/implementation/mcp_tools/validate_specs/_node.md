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
references, and artifact and verdict staleness.

# Public

## Package

`package mcpvalidatespecs`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpvalidatespecs"`

## Interface

```go
type StalenessEntry struct {
	Node         string
	ArtifactPath string
	Status       string
	Detail       string
	Rank         int
	Result       string
	Blocked      bool
	BlockedBy    string
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

`StalenessEntry.Blocked` is true when the entry has a
dependency the session cannot satisfy. `BlockedBy` is
a single string identifying the first blocker found.
Blocked is orthogonal to Status — a node can be stale
and blocked at the same time.

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
   Build `knownSpecNodes` as a `[]string`: for each
   node, append node.Reference.LogicalName.
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

### Step 6 — Staleness and blocking detection

7. Initialize `blockedSet` as an empty map from string
   to string (manifestKey → reason). This propagates
   blocking transitively.

   Determine processing order for staleness checks:
   If ranked_entries is non-empty:
     Order nodes by rank ascending, then by
     logical_name ascending within equal rank.
   Else:
     Order nodes alphabetically by logical_name.

   For each node whose frontmatter Type is not nil,
   in the above order:

     a. Derive the manifest key: strip "SPEC/" prefix
        from node.logical_name. If
        *node.Frontmatter.Type is "verdict", prepend
        "VERDICT/"; otherwise prepend "ARTIFACT/".
        Let `resolved_output` =
        `parsing.ResolvedOutput(node)` (explicit output
        or default path).

     b. Call `chainresolver.ChainResolve(node.logical_name, knownSpecNodes)`.
        If it
        fails: Append StalenessEntry(
          node=node.logical_name,
          artifact_path=*resolved_output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     c. Call `chainhash.ChainHashCompute(chain)` using the result
        from step (b). It returns
        `(chain_hash, positions, err)`. Ignore
        `positions`. If it fails: Append
        StalenessEntry(
          node=node.logical_name,
          artifact_path=*resolved_output,
          status="missing", detail=<error message>,
          rank=<node rank or 0 if unavailable>)
        to staleness. Continue to next node.

     d. Look up the manifest key in m.Entries.

        If no entry exists: Append StalenessEntry with
        status="missing", detail="no manifest entry".

        If entry exists:
          Compare entry.ChainHash with computed chain
          hash. If they differ: Append StalenessEntry
          with status="stale", detail="manifest chain
          hash <entry.ChainHash> does not match
          expected hash <chain hash>",
          result=entry.Result.

          If chain hashes match: check the file on
          disk. Construct oslayer.CfsPath from
          *resolved_output. Call
          `oslayer.OpenFile(path, "read", 30000)`. If it
          fails (file does not exist): Append
          StalenessEntry with status="missing".
          Else: read the full file content, compute
          its SHA-1 hash (base64url, 27 chars).
          Call `handle.Close()`. Compare with
          entry.Checksum. If they differ: Append
          StalenessEntry with status="modified",
          detail="file checksum does not match
          manifest checksum",
          result=entry.Result.

        If chain hash matches and checksum matches:
        skip (up to date).

        Set artifact_path from *resolved_output.
        Set rank from the node's rank (from Step 4,
        or 0 if no ranking available).

     e. **Blocking check** — after computing the node's
        own status (stale/missing/modified/up-to-date),
        check whether the entry is blocked. Check these
        conditions in order; stop at the first match:

        1. **wait_on targets**: expand globs in
           node.Frontmatter.WaitOn using
           parsing.ExpandGlob. For each target:
           - Derive its manifest key (the target name
             itself — it is already ARTIFACT/ or
             VERDICT/).
           - Look up in manifest. If no entry exists,
             or chain hash does not match the current
             computed hash, or checksum does not match
             the file on disk: not satisfied.
           - For VERDICT/ targets: additionally check
             that entry.Result is "pass" or "accepted".
           - If not satisfied: blocked. Set
             reason = "wait_on target not satisfied:
             <target>".

        2. **Modified ARTIFACT/ dependencies**: expand
           globs in node.Frontmatter.Imports and
           node.Frontmatter.Input. For each entry that
           starts with "ARTIFACT/":
           - Look up its manifest key. If entry exists
             and its checksum does not match the file
             on disk: blocked. Set
             reason = "dependency artifact modified:
             <ref>".

        3. **Transitive blocking**: for each dependency
           checked above (wait_on targets and ARTIFACT/
           imports/input), if its manifest key is in
           `blockedSet`: blocked. Set
           reason = "dependency blocked: <key>".

        If any condition matched: set Blocked = true,
        BlockedBy = reason on the StalenessEntry. Add
        the current node's manifest key to
        `blockedSet` with the same reason.

        Blocking applies even to up-to-date entries
        that were skipped in the status check — an
        up-to-date node whose wait_on target is not
        satisfied is still blocked and must appear in
        the staleness list. In this case, add a
        StalenessEntry with Status = "" (empty),
        Blocked = true, BlockedBy = reason.

### Step 7 — Orphan detection

8. For each entry in m.Entries:
     Derive the generating node's logical name: if the
     key starts with "ARTIFACT/", strip "ARTIFACT/"
     prefix and prepend "SPEC/". If the key starts with
     "VERDICT/", strip "VERDICT/" prefix and prepend
     "SPEC/". Otherwise skip (unknown prefix).
     If no successfully parsed node has that logical
     name, or if the node's frontmatter Type is nil: Append StalenessEntry(
       node=entry key,
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
