<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@186VSFnhRiCMTjD_6SWT31Thg74 -->

## Test suite: MCPValidateSpecs

Each test case creates a spec tree on disk, then calls `MCPValidateSpecs`.
The function always returns a `ValidationReport` and never raises an error.

---

### Happy path

#### TC-HP-1: Clean tree — no errors

Setup:
- Create a spec tree with ROOT (`_node.md` containing a public section)
  and ROOT/a (leaf node with `output = "out/a.go"`).
- Compute the current chain hash for ROOT/a.
- Create `"out/a.go"` containing a valid artifact tag whose hash matches
  the current chain hash.

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` is empty.
- `ValidationReport.cycles` is empty.
- `ValidationReport.staleness` is empty.

---

#### TC-HP-2: Stale artifact detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with `output = "out/a.go"`).
- Create `"out/a.go"` containing an artifact tag with an outdated hash
  (any hash that does not match the current chain hash for ROOT/a).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains exactly one `StalenessEntry`.
- That entry has `node = "ROOT/a"`, `status = "stale"`, and `rank` is present
  (an integer value).

---

#### TC-HP-3: Missing artifact detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with `output = "out/a.go"`).
- Do not create the file `"out/a.go"`.

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains exactly one `StalenessEntry`.
- That entry has `node = "ROOT/a"` and `status = "missing"`.

---

#### TC-HP-4: Malformed tag detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with `output = "out/a.go"`).
- Create `"out/a.go"` with content that contains no artifact tag
  (or contains a line that cannot be parsed as an artifact tag).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains exactly one `StalenessEntry`.
- That entry has `node = "ROOT/a"` and `status = "malformed tag"`.

---

#### TC-HP-5: Staleness entries include rank

Setup:
- Create a spec tree with:
  - ROOT (public section)
  - ROOT/a (leaf with `output = "out/a.go"`)
  - ROOT/b (leaf with `output = "out/b.go"`, `depends_on = ["ROOT/a"]`)
- Create both output files with outdated hashes.

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains two `StalenessEntry` records,
  one for ROOT/a and one for ROOT/b.
- Both entries have a `rank` field that is an integer.
- ROOT/a's `rank` is strictly less than ROOT/b's `rank`.

---

#### TC-HP-6: Staleness ordered by rank then name

Setup:
- Create a spec tree with ROOT, ROOT/z (leaf with `output = "out/z.go"`),
  and ROOT/a (leaf with `output = "out/a.go"`).
- Create both output files with outdated hashes.
- Neither node depends on the other.

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains two entries.
- The entry for ROOT/a appears before the entry for ROOT/z
  (same rank, ordered alphabetically by node name).

---

### Format errors

#### TC-FE-1: Format error from invalid depends_on

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with
  `depends_on = ["ROOT/missing"]`).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains at least one
  `spectreevalidate.FormatError`.
- That error has `node = "ROOT/a"` and `rule = "dependency_targets"`.

---

#### TC-FE-2: Format error from parse failure

Setup:
- Create a spec tree with ROOT and ROOT/a whose `_node.md` has invalid
  content (e.g. text appearing before any heading, making it unparseable).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains at least one
  `spectreevalidate.FormatError` with `node = "ROOT/a"` and
  `rule = "parse"`.
- Other valid nodes in the tree are still validated (validation continues).

---

#### TC-FE-3: Continues after parse failure

Setup:
- Create a spec tree with ROOT, ROOT/a (invalid `_node.md` content),
  and ROOT/b (valid leaf with `output = "out/b.go"` and an outdated
  artifact hash).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains a `spectreevalidate.FormatError`
  for ROOT/a with `rule = "parse"`.
- `ValidationReport.staleness` contains a `StalenessEntry` for ROOT/b.
- Both are present in the same report.

---

### Cycle detection

#### TC-CD-1: Simple cycle detected

Setup:
- Create a spec tree with ROOT, ROOT/a (leaf with
  `depends_on = ["ROOT/b"]`), and ROOT/b (leaf with
  `depends_on = ["ROOT/a"]`).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.cycles` is not empty.
- `ValidationReport.cycles` contains at least one of `"ROOT/a"` or `"ROOT/b"`.

---

#### TC-CD-2: Ranking skipped when format errors exist

Setup:
- Create a spec tree with ROOT, ROOT/a (leaf with an invalid `depends_on`
  target, e.g. `depends_on = ["ROOT/missing"]`), and ROOT/b (valid leaf
  with `output = "out/b.go"` and an outdated artifact hash).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` is not empty.
- Ranking is skipped — ROOT/b's `StalenessEntry` has `rank = 0`
  (default when no ranking is available).

---

### Edge cases

#### TC-EC-1: Empty spec tree — scan fails

Setup:
- Do not create a `code-from-spec/` directory.

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains at least one
  `spectreevalidate.FormatError` with `rule = "scan"`.
- `ValidationReport.cycles` is empty.
- `ValidationReport.staleness` is empty.

---

#### TC-EC-2: Node with no output — not in staleness

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with no `output` field).

Actions:
- Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains no entry for ROOT/a.
- Staleness checks only apply to nodes that declare an output.
