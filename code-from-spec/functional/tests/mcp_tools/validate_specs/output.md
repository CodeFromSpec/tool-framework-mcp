<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@TIFpsxuRKtakhvrRjpeO5D-DAWw -->

# Test Specification: MCPValidateSpecs

Each test case creates a spec tree on disk with `_node.md` files, then calls
`MCPValidateSpecs`. The function always returns a `ValidationReport` and never
raises an error. Problems are collected in the report.

---

## Happy Path

---

### TC-HP-01: Clean tree — no errors

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a node — leaf, outputs = [{id: "code", path: "out/a.go"}].
- Create "out/a.go" containing a valid artifact tag whose hash matches the
  current chain hash for ROOT/a.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` is empty.
- `cycles` is empty.
- `staleness` is empty.

---

### TC-HP-02: Stale artifact detected

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, outputs = [{id: "code", path: "out/a.go"}].
- Create "out/a.go" containing an artifact tag with an outdated (non-matching)
  hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `status` = `"stale"`.
  - `output_id` matching `"code"`.
  - `rank` is present (a non-negative integer).

---

### TC-HP-03: Missing artifact detected

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, outputs = [{id: "code", path: "out/a.go"}].
- Do not create the output file "out/a.go".

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `status` = `"missing"`.
  - `output_id` matching `"code"`.

---

### TC-HP-04: Malformed tag detected

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, outputs = [{id: "code", path: "out/a.go"}].
- Create "out/a.go" with arbitrary content that contains no artifact tag (or
  contains a tag that cannot be parsed).

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `status` = `"malformed tag"`.
  - `output_id` matching `"code"`.

---

### TC-HP-05: Multiple outputs — each checked independently

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, outputs = [
    {id: "x", path: "out/x.go"},
    {id: "y", path: "out/y.go"}
  ].
- Create "out/x.go" with a valid artifact tag whose hash matches the current
  chain hash for ROOT/a.
- Do not create "out/y.go".

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` with:
  - `output_id` = `"y"`.
  - `status` = `"missing"`.
- No `StalenessEntry` exists for `output_id` = `"x"` (hash matches, so it is
  not included).

---

### TC-HP-06: Staleness entries include rank

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, outputs = [{id: "code-a", path: "out/a.go"}].
- ROOT/b node — leaf, outputs = [{id: "code-b", path: "out/b.go"}],
  depends_on = ["ROOT/a"].
- Create "out/a.go" and "out/b.go" each with an artifact tag containing an
  outdated hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains two `StalenessEntry` records, one for ROOT/a and one for
  ROOT/b.
- Both entries have a `rank` value present (a non-negative integer).
- The `rank` of the ROOT/a entry is strictly less than the `rank` of the ROOT/b
  entry.

---

### TC-HP-07: Staleness ordered by rank then name

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/z node — leaf, outputs = [{id: "code-z", path: "out/z.go"}]. Stale.
- ROOT/a node — leaf, outputs = [{id: "code-a", path: "out/a.go"}]. Stale.
- Neither ROOT/z nor ROOT/a depends on the other (same rank).
- Create both output files with outdated artifact tags.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` has two entries.
- The entry for ROOT/a appears before the entry for ROOT/z (same rank,
  alphabetical ordering by node name).

---

## Format Errors

---

### TC-FE-01: Format error from invalid depends_on

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, depends_on = ["ROOT/missing"] (the target does not exist
  in the tree).

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a with
  `rule` = `"dependency_targets"`.

---

### TC-FE-02: Format error from parse failure

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node whose `_node.md` has invalid content (e.g., body text that
  appears before any heading, making it unparseable).

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a with `rule` = `"parse"`.
- Other valid nodes in the tree are still validated (validation continues after
  the parse failure).

---

### TC-FE-03: Continues after parse failure

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node whose `_node.md` has invalid content (unparseable).
- ROOT/b node — valid leaf, outputs = [{id: "code-b", path: "out/b.go"}].
  Create "out/b.go" with a stale artifact tag.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a.
- `staleness` contains a `StalenessEntry` for ROOT/b.
- Both are present in the same report.

---

## Cycle Detection

---

### TC-CY-01: Simple cycle detected

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, depends_on = ["ROOT/b"].
- ROOT/b node — leaf, depends_on = ["ROOT/a"].

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `cycles` is not empty.
- `cycles` contains at least one of the logical names `"ROOT/a"` or `"ROOT/b"`.

---

### TC-CY-02: Ranking skipped when format errors exist

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf, depends_on = ["ROOT/missing"] (invalid target, causes a
  format error).
- ROOT/b node — valid leaf, outputs = [{id: "code-b", path: "out/b.go"}].
  Create "out/b.go" with a stale artifact tag.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` is not empty (contains the error for ROOT/a).
- Ranking is skipped because format errors exist — any `StalenessEntry` for
  ROOT/b has `rank` = `0` (the default when no ranking is available).

---

## Edge Cases

---

### TC-EC-01: Empty spec tree — scan fails

**Setup**

Do not create a `code-from-spec/` directory.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` with `rule` = `"scan"`.
- `cycles` is empty.
- `staleness` is empty.

---

### TC-EC-02: Node with no outputs — not in staleness

**Setup**

Create a spec tree:
- ROOT node.
- ROOT/a node — leaf with no outputs declared.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains no `StalenessEntry` for ROOT/a.
- The staleness check only runs for nodes that declare outputs.
