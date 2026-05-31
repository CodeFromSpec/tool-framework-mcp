<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@fOWQuH1yYA4j9GceTRSlnQSBrYY -->

# Test Specification: MCPValidateSpecs

Each test case creates a spec tree on disk with `_node.md` files, then calls
`MCPValidateSpecs`. The function always returns a `ValidationReport` — it
never raises an error.

---

## Happy Path

### TC-HP-1: Clean tree — no errors

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
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

### TC-HP-2: Stale artifact detected

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
- Create "out/a.go" containing an artifact tag with an outdated (wrong) hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `output_id` matching the output's id ("code")
  - `status` = "stale"
  - `rank` is present (integer value assigned by `NodeRankCompute`).
- `format_errors` is empty.
- `cycles` is empty.

---

### TC-HP-3: Missing artifact detected

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
- Do not create "out/a.go".

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `status` = "missing"
- `format_errors` is empty.
- `cycles` is empty.

---

### TC-HP-4: Malformed tag detected

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
- Create "out/a.go" with content that contains no artifact tag (or a
  malformed one that cannot be parsed).

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `status` = "malformed tag"
- `format_errors` is empty.
- `cycles` is empty.

---

### TC-HP-5: Multiple outputs — each checked independently

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [
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
  - `output_id` = "y"
  - `status` = "missing"
- There is no `StalenessEntry` with `output_id` = "x" (its hash matches).
- `format_errors` is empty.
- `cycles` is empty.

---

### TC-HP-6: Staleness entries include rank

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
- ROOT/b leaf node with outputs = [{id: "code", path: "out/b.go"}] and
  depends_on = ["ROOT/a"].
- Create "out/a.go" and "out/b.go" each with an artifact tag containing an
  outdated hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains two `StalenessEntry` records, one for ROOT/a and one
  for ROOT/b.
- Both entries have a `rank` integer value.
- The `rank` of ROOT/a's entry is strictly lower than the `rank` of ROOT/b's
  entry (ROOT/a has no dependency on ROOT/b, ROOT/b depends on ROOT/a).
- `format_errors` is empty.
- `cycles` is empty.

---

### TC-HP-7: Staleness ordered by rank then name

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/z leaf node with outputs = [{id: "code", path: "out/z.go"}].
- ROOT/a leaf node with outputs = [{id: "code", path: "out/a.go"}].
- Neither ROOT/z nor ROOT/a depends on the other.
- Create "out/z.go" and "out/a.go" each with an artifact tag containing an
  outdated hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains two `StalenessEntry` records.
- Both entries have the same `rank` value.
- The entry for ROOT/a appears before the entry for ROOT/z (alphabetical
  order by node name when ranks are equal).

---

## Format Errors

### TC-FE-1: Format error from invalid depends_on

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with depends_on = ["ROOT/missing"] where "ROOT/missing"
  does not exist in the tree.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a with
  `rule` = "dependency_targets".

---

### TC-FE-2: Format error from parse failure

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a whose `_node.md` contains invalid content (e.g., text appearing
  before any heading) that causes a parse failure.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a with `rule` = "parse".
- Other nodes in the tree are still validated (the report is not aborted on
  the first error).

---

### TC-FE-3: Continues after parse failure

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a whose `_node.md` has invalid content (causes parse failure).
- ROOT/b leaf node with outputs = [{id: "code", path: "out/b.go"}] whose
  output file exists but contains an artifact tag with an outdated hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a.
- `staleness` contains a `StalenessEntry` for ROOT/b with `status` = "stale".
- Both are reported in the same `ValidationReport`.

---

## Cycle Detection

### TC-CD-1: Simple cycle detected

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with depends_on = ["ROOT/b"].
- ROOT/b leaf node with depends_on = ["ROOT/a"].

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `cycles` is not empty.
- `cycles` contains at least one of the strings "ROOT/a" or "ROOT/b".

---

### TC-CD-2: Ranking skipped when format errors exist

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with depends_on = ["ROOT/missing"] (invalid target, causes
  a format error with rule = "dependency_targets").
- ROOT/b valid leaf node with outputs = [{id: "code", path: "out/b.go"}]
  whose output file contains an artifact tag with an outdated hash.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` is not empty (contains the error for ROOT/a).
- Because format errors are present, ranking is skipped.
- The `StalenessEntry` for ROOT/b has `rank` = 0 (the default when no
  ranking result is available).

---

## Edge Cases

### TC-EC-1: Empty spec tree — scan fails

**Setup**

Do not create a `code-from-spec/` directory.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` with `rule` = "scan".
- `cycles` is empty.
- `staleness` is empty.

---

### TC-EC-2: Node with no outputs — not in staleness

**Setup**

Create a spec tree:
- ROOT node with a public section.
- ROOT/a leaf node with no outputs defined.

**Action**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains no `StalenessEntry` for ROOT/a (staleness check only
  runs for nodes that declare outputs).
- `format_errors` is empty.
- `cycles` is empty.
