---
depends_on:
  - ROOT/functional/logic/mcp_tools/validate_specs(interface)
  - ROOT/functional/logic/chain/hash(interface)
output: code-from-spec/functional/tests/mcp_tools/validate_specs/output.md
---

# ROOT/functional/tests/mcp_tools/validate_specs

Test cases for the validate specs tool.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files, then call `MCPValidateSpecs`. The function always
returns a `ValidationReport` — it never raises an error.

### Happy path

#### Clean tree — no errors

Create a spec tree with SPEC (with `# Public` containing
a `## Context` subsection) and SPEC/a (leaf with
output = "out/a.go"). Create "out/a.go" with a valid
artifact tag whose hash matches the current chain hash.
Call MCPValidateSpecs.

Expect report with empty format_errors, empty cycles,
and empty staleness.

#### Stale artifact detected

Create a spec tree with SPEC and SPEC/a (leaf with
output). Create the output file with an artifact tag
containing an outdated hash. Call MCPValidateSpecs.

Expect report contains a StalenessEntry for SPEC/a
with status = "stale" and rank present.

#### Missing artifact detected

Create a spec tree with SPEC and SPEC/a (leaf with
output). Do not create the output file. Call
MCPValidateSpecs.

Expect report contains a StalenessEntry for SPEC/a
with status = "missing".

#### Malformed tag detected

Create a spec tree with SPEC and SPEC/a (leaf with
output). Create the output file with content that
has no artifact tag (or a malformed one). Call
MCPValidateSpecs.

Expect report contains a StalenessEntry for SPEC/a
with status = "malformed tag".

#### Staleness entries include rank

Create a spec tree with SPEC, SPEC/a (leaf with
output), SPEC/b (leaf with output, depends_on =
["SPEC/a"]). Create output files with outdated hashes.
Call MCPValidateSpecs.

Expect both StalenessEntries have rank values. SPEC/a's
rank is lower than SPEC/b's rank.

#### Staleness ordered by rank then name

Create a spec tree with SPEC, SPEC/z (leaf with
output), SPEC/a (leaf with output). Both stale. Call
MCPValidateSpecs.

Expect staleness entries ordered: SPEC/a before SPEC/z
(same rank, alphabetical).

### Format errors

#### Format error from invalid depends_on

Create a spec tree with SPEC and SPEC/a (leaf with
depends_on = ["SPEC/missing"]). Call MCPValidateSpecs.

Expect report contains a FormatError for SPEC/a with
rule = "dependency_targets".

#### Format error from parse failure

Create a spec tree with SPEC and SPEC/a whose _node.md
has invalid content (e.g. text before any heading). Call
MCPValidateSpecs.

Expect report contains a FormatError with rule = "parse"
for SPEC/a. Other nodes are still validated.

#### Continues after parse failure

Create a spec tree with SPEC, SPEC/a (invalid content),
SPEC/b (valid leaf with output, stale artifact). Call
MCPValidateSpecs.

Expect report contains a FormatError for SPEC/a AND a
StalenessEntry for SPEC/b. Both are reported.

#### Subdirectory without _node.md detected

Create a spec tree with SPEC and SPEC/a (leaf). Also
create an empty subdirectory `code-from-spec/b/` with
no `_node.md` inside. Call MCPValidateSpecs.

Expect report contains a FormatError with rule =
"missing_node_md" for the `code-from-spec/b/` directory.

#### _-prefixed dir under code-from-spec not flagged

Create a spec tree with SPEC. Also create a directory
`code-from-spec/_tools/` with no `_node.md`. Call
MCPValidateSpecs.

Expect no FormatError for `_tools/` — `_`-prefixed
directories directly under `code-from-spec/` are
ignored.

### Cycle detection

#### Simple cycle detected

Create a spec tree with SPEC, SPEC/a (leaf, depends_on
= ["SPEC/b"]), SPEC/b (leaf, depends_on = ["SPEC/a"]).
Call MCPValidateSpecs.

Expect report.cycles is not empty and contains at least
one of SPEC/a or SPEC/b.

#### Ranking skipped when format errors exist

Create a spec tree with SPEC, SPEC/a (leaf with invalid
depends_on target), SPEC/b (valid leaf with output).
Call MCPValidateSpecs.

Expect report contains format errors. Ranking is
skipped — staleness entries for SPEC/b have rank = 0
(default when no ranking available).

### Edge cases

#### Empty spec tree — scan fails

Do not create a code-from-spec/ directory. Call
MCPValidateSpecs.

Expect report contains a FormatError with rule = "scan".
Cycles and staleness are empty.

#### Node with no output — not in staleness

Create a spec tree with SPEC and SPEC/a (leaf with no
output). Call MCPValidateSpecs.

Expect no StalenessEntry for SPEC/a — staleness check
only runs for nodes with output.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPValidateSpecs`.
- Use the record names from the interface:
  `ValidationReport`, `StalenessEntry`,
  `spectreevalidate.FormatError` (qualified — it is
  declared in the `spectreevalidate` module).
- Use formal error names and status values as defined
  in the interface.
- Each test case creates a spec tree on disk, then
  calls `MCPValidateSpecs`.
- When a test needs to write an artifact file with a
  valid artifact tag, compute the current chain hash
  using `ChainHashCompute` (from the `chain/hash`
  module) and use it in the tag. When a test needs a
  stale artifact tag, use any 27-character base64url
  string that differs from the current chain hash.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
  Never place content directly under `# Public`
  without a subsection heading — this is a format
  error.
- Logical names map to filesystem paths as follows:
  `SPEC` → `code-from-spec/_node.md`,
  `SPEC/x` → `code-from-spec/x/_node.md`,
  `SPEC/x/y` → `code-from-spec/x/y/_node.md`.
  The `SPEC` prefix is not a directory — it is stripped
  when resolving to a path. When a test says "create a
  spec tree with SPEC and SPEC/a", the files are
  `code-from-spec/_node.md` and
  `code-from-spec/a/_node.md`.
- The first heading in each `_node.md` must be the full
  logical name: `# SPEC` for the root,
  `# SPEC/a` for a child node.
