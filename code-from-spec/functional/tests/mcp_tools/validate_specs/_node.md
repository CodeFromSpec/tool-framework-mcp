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

Create a spec tree with ROOT (with `# Public` containing
a `## Context` subsection) and ROOT/a (leaf with
output = "out/a.go"). Create "out/a.go" with a valid
artifact tag whose hash matches the current chain hash.
Call MCPValidateSpecs.

Expect report with empty format_errors, empty cycles,
and empty staleness.

#### Stale artifact detected

Create a spec tree with ROOT and ROOT/a (leaf with
output). Create the output file with an artifact tag
containing an outdated hash. Call MCPValidateSpecs.

Expect report contains a StalenessEntry for ROOT/a
with status = "stale" and rank present.

#### Missing artifact detected

Create a spec tree with ROOT and ROOT/a (leaf with
output). Do not create the output file. Call
MCPValidateSpecs.

Expect report contains a StalenessEntry for ROOT/a
with status = "missing".

#### Malformed tag detected

Create a spec tree with ROOT and ROOT/a (leaf with
output). Create the output file with content that
has no artifact tag (or a malformed one). Call
MCPValidateSpecs.

Expect report contains a StalenessEntry for ROOT/a
with status = "malformed tag".

#### Staleness entries include rank

Create a spec tree with ROOT, ROOT/a (leaf with
output), ROOT/b (leaf with output, depends_on =
["ROOT/a"]). Create output files with outdated hashes.
Call MCPValidateSpecs.

Expect both StalenessEntries have rank values. ROOT/a's
rank is lower than ROOT/b's rank.

#### Staleness ordered by rank then name

Create a spec tree with ROOT, ROOT/z (leaf with
output), ROOT/a (leaf with output). Both stale. Call
MCPValidateSpecs.

Expect staleness entries ordered: ROOT/a before ROOT/z
(same rank, alphabetical).

### Format errors

#### Format error from invalid depends_on

Create a spec tree with ROOT and ROOT/a (leaf with
depends_on = ["ROOT/missing"]). Call MCPValidateSpecs.

Expect report contains a FormatError for ROOT/a with
rule = "dependency_targets".

#### Format error from parse failure

Create a spec tree with ROOT and ROOT/a whose _node.md
has invalid content (e.g. text before any heading). Call
MCPValidateSpecs.

Expect report contains a FormatError with rule = "parse"
for ROOT/a. Other nodes are still validated.

#### Continues after parse failure

Create a spec tree with ROOT, ROOT/a (invalid content),
ROOT/b (valid leaf with output, stale artifact). Call
MCPValidateSpecs.

Expect report contains a FormatError for ROOT/a AND a
StalenessEntry for ROOT/b. Both are reported.

### Cycle detection

#### Simple cycle detected

Create a spec tree with ROOT, ROOT/a (leaf, depends_on
= ["ROOT/b"]), ROOT/b (leaf, depends_on = ["ROOT/a"]).
Call MCPValidateSpecs.

Expect report.cycles is not empty and contains at least
one of ROOT/a or ROOT/b.

#### Ranking skipped when format errors exist

Create a spec tree with ROOT, ROOT/a (leaf with invalid
depends_on target), ROOT/b (valid leaf with output).
Call MCPValidateSpecs.

Expect report contains format errors. Ranking is
skipped — staleness entries for ROOT/b have rank = 0
(default when no ranking available).

### Edge cases

#### Empty spec tree — scan fails

Do not create a code-from-spec/ directory. Call
MCPValidateSpecs.

Expect report contains a FormatError with rule = "scan".
Cycles and staleness are empty.

#### Node with no output — not in staleness

Create a spec tree with ROOT and ROOT/a (leaf with no
output). Call MCPValidateSpecs.

Expect no StalenessEntry for ROOT/a — staleness check
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
