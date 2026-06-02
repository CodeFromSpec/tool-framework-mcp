<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@oVVvediNyF25aTFFlKwIh6S-IsU -->

# Test Specification: MCPValidateSpecs

## Test cases

All tests create a spec tree on disk with `_node.md` files, then call `MCPValidateSpecs`.
The function always returns a `ValidationReport` — it never raises an error.
`ValidationReport` has fields: format_errors, cycles, staleness.
Each `StalenessEntry` has fields: node, artifact_path, status, detail, rank.

---

### Happy path

#### Clean tree — no errors

Setup:
- Create ROOT/_node.md with a public section.
- Create ROOT/a/_node.md as a leaf with output = "out/a.go".
- Create "out/a.go" with a valid artifact tag whose hash matches the current chain hash for ROOT/a.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors is empty.
- report.cycles is empty.
- report.staleness is empty.

---

#### Stale artifact detected

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field.
- Create the output file with an artifact tag containing an outdated hash.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness contains one StalenessEntry for ROOT/a.
- That entry has status = "stale".
- That entry has a rank value present.

---

#### Missing artifact detected

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field.
- Do not create the output file.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness contains one StalenessEntry for ROOT/a with status = "missing".

---

#### Malformed tag detected

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field.
- Create the output file with content that has no artifact tag (or a malformed one).

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness contains one StalenessEntry for ROOT/a with status = "malformed tag".

---

#### Staleness entries include rank

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field.
- Create ROOT/b/_node.md as a leaf with an output field and depends_on = ["ROOT/a"].
- Create both output files with outdated hashes.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness contains StalenessEntries for both ROOT/a and ROOT/b.
- Both entries have rank values.
- ROOT/a's rank is lower than ROOT/b's rank.

---

#### Staleness ordered by rank then name

Setup:
- Create ROOT/_node.md.
- Create ROOT/z/_node.md as a leaf with an output field.
- Create ROOT/a/_node.md as a leaf with an output field.
- Create both output files with outdated hashes.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness has ROOT/a's entry before ROOT/z's entry (same rank, alphabetical order).

---

### Format errors

#### Format error from invalid depends_on

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with depends_on = ["ROOT/missing"] (target does not exist).

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors contains a FormatError for ROOT/a with rule = "dependency_targets".

---

#### Format error from parse failure

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with invalid content (e.g., text appearing before any heading).

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors contains a FormatError with rule = "parse" for ROOT/a.
- Other nodes in the tree are still validated.

---

#### Continues after parse failure

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with invalid content (parse failure).
- Create ROOT/b/_node.md as a valid leaf with an output field, with a stale artifact.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors contains a FormatError for ROOT/a.
- report.staleness contains a StalenessEntry for ROOT/b.
- Both are reported in the same ValidationReport.

---

### Cycle detection

#### Simple cycle detected

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with depends_on = ["ROOT/b"].
- Create ROOT/b/_node.md as a leaf with depends_on = ["ROOT/a"].

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.cycles is not empty.
- report.cycles contains at least one of "ROOT/a" or "ROOT/b".

---

#### Ranking skipped when format errors exist

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an invalid depends_on target.
- Create ROOT/b/_node.md as a valid leaf with an output field.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors is not empty.
- Ranking is skipped — any StalenessEntry for ROOT/b has rank = 0 (default when no ranking is available).

---

### Edge cases

#### Empty spec tree — scan fails

Setup:
- Do not create a code-from-spec/ directory.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.format_errors contains a FormatError with rule = "scan".
- report.cycles is empty.
- report.staleness is empty.

---

#### Node with no output — not in staleness

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with no output field.

Action:
- Call MCPValidateSpecs.

Expected outcome:
- report.staleness contains no StalenessEntry for ROOT/a — staleness check only runs for nodes with an output field.
