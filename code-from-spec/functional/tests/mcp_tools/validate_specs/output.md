<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@QQUT06e0n7X-MBuFHgztj6S0XiM -->

## Test cases for MCPValidateSpecs

Each test creates a spec tree on disk with `_node.md` files, then calls
`MCPValidateSpecs`. The function always returns a `ValidationReport` —
it never raises an error.

---

### Happy path

#### Clean tree — no errors

Setup:
- Create a spec tree with ROOT (with public section) and ROOT/a (leaf
  with output = "out/a.go").
- Compute the chain hash for ROOT/a using `ChainHashCompute`.
- Create "out/a.go" with a valid artifact tag containing that hash.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` is empty.
- `ValidationReport.cycles` is empty.
- `ValidationReport.staleness` is empty.

---

#### Stale artifact detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with output).
- Create the output file with an artifact tag containing a 27-character
  base64url string that differs from the current chain hash.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains one `StalenessEntry` for ROOT/a.
- `StalenessEntry.status` = `"stale"`.
- `StalenessEntry.rank` is present (integer value).

---

#### Missing artifact detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with output).
- Do not create the output file.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains one `StalenessEntry` for ROOT/a.
- `StalenessEntry.status` = `"missing"`.

---

#### Malformed tag detected

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with output).
- Create the output file with content that has no artifact tag or a
  malformed artifact tag.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains one `StalenessEntry` for ROOT/a.
- `StalenessEntry.status` = `"malformed tag"`.

---

#### Staleness entries include rank

Setup:
- Create a spec tree with ROOT, ROOT/a (leaf with output), ROOT/b (leaf
  with output, depends_on = ["ROOT/a"]).
- Create both output files with outdated artifact tags (stale hashes).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains `StalenessEntry` records for
  both ROOT/a and ROOT/b.
- Each entry has a `rank` value.
- ROOT/a's `rank` is lower than ROOT/b's `rank`.

---

#### Staleness ordered by rank then name

Setup:
- Create a spec tree with ROOT, ROOT/z (leaf with output), ROOT/a (leaf
  with output).
- Both output files have stale artifact tags.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` has two entries.
- ROOT/a appears before ROOT/z (same rank, alphabetical ordering).

---

### Format errors

#### Format error from invalid depends_on

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with depends_on =
  ["ROOT/missing"]).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains a
  `spectreevalidate.FormatError` for ROOT/a with rule = "dependency_targets".

---

#### Format error from parse failure

Setup:
- Create a spec tree with ROOT and ROOT/a whose `_node.md` has invalid
  content (e.g. text before any heading).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains a `spectreevalidate.FormatError`
  with rule = `"parse"` for ROOT/a.
- Other nodes are still validated (validation continues).

---

#### Continues after parse failure

Setup:
- Create a spec tree with ROOT, ROOT/a (invalid `_node.md` content),
  ROOT/b (valid leaf with output and a stale artifact).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains a `spectreevalidate.FormatError`
  for ROOT/a.
- `ValidationReport.staleness` contains a `StalenessEntry` for ROOT/b.
- Both are reported in the same `ValidationReport`.

---

### Cycle detection

#### Simple cycle detected

Setup:
- Create a spec tree with ROOT, ROOT/a (leaf, depends_on = ["ROOT/b"]),
  ROOT/b (leaf, depends_on = ["ROOT/a"]).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.cycles` is not empty.
- `ValidationReport.cycles` contains at least one of ROOT/a or ROOT/b.

---

#### Ranking skipped when format errors exist

Setup:
- Create a spec tree with ROOT, ROOT/a (leaf with invalid depends_on
  target), ROOT/b (valid leaf with output).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` is not empty.
- Ranking is skipped — any `StalenessEntry` for ROOT/b has `rank` = 0
  (the default when no ranking is available).

---

### Edge cases

#### Empty spec tree — scan fails

Setup:
- Do not create a `code-from-spec/` directory.

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.format_errors` contains a `spectreevalidate.FormatError`
  with rule = `"scan"`.
- `ValidationReport.cycles` is empty.
- `ValidationReport.staleness` is empty.

---

#### Node with no output — not in staleness

Setup:
- Create a spec tree with ROOT and ROOT/a (leaf with no output field).

Action: Call `MCPValidateSpecs`.

Expected outcome:
- `ValidationReport.staleness` contains no `StalenessEntry` for ROOT/a.
- Staleness check only runs for nodes with an output field.
