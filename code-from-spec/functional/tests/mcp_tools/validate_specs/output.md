<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@szJw-vaJCf5FBjg3NWLmPyxQyNQ -->

# Test Specification: MCPValidateSpecs

## Interface

```
record StalenessEntry
  node: string
  output_id: string
  artifact_path: string
  status: string
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of FormatError
  cycles: list of string
  staleness: list of StalenessEntry

function MCPValidateSpecs() -> ValidationReport
```

---

## Happy Path

### Test: Clean tree — no errors

**Setup**

1. Create a spec tree on disk with:
   - ROOT node containing a public section.
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
2. Create the file "out/a.go" with a valid artifact tag whose hash matches
   the current chain hash for ROOT/a.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` is empty.
- `cycles` is empty.
- `staleness` is empty.

---

### Test: Stale artifact detected

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
2. Create the file "out/a.go" with an artifact tag containing an outdated
   (non-matching) hash.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `output_id` matches the declared output id ("code")
  - `status` = "stale"
  - `rank` is present (an integer value).

---

### Test: Missing artifact detected

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
2. Do not create the file "out/a.go".

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `output_id` = "code"
  - `status` = "missing"

---

### Test: Malformed tag detected

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
2. Create the file "out/a.go" with content that contains no artifact tag
   (or a tag that cannot be parsed).

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` for ROOT/a with:
  - `node` = "ROOT/a"
  - `output_id` = "code"
  - `status` = "malformed tag"

---

### Test: Multiple outputs — each checked independently

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with outputs = [
       {id: "x", path: "out/x.go"},
       {id: "y", path: "out/y.go"}
     ].
2. Create "out/x.go" with a valid artifact tag whose hash matches the
   current chain hash for ROOT/a.
3. Do not create "out/y.go".

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains exactly one `StalenessEntry` with:
  - `output_id` = "y"
  - `status` = "missing"
- No `StalenessEntry` exists for `output_id` = "x" (hash matches, so it
  is not included).

---

### Test: Staleness entries include rank

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
   - ROOT/b node as a leaf with outputs = [{id: "code", path: "out/b.go"}]
     and depends_on = ["ROOT/a"].
2. Create "out/a.go" with an artifact tag containing an outdated hash.
3. Create "out/b.go" with an artifact tag containing an outdated hash.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains `StalenessEntry` records for both ROOT/a and ROOT/b.
- Both entries have a `rank` value (an integer).
- ROOT/a's `rank` is strictly less than ROOT/b's `rank` (ROOT/a has no
  dependents on ROOT/b, so it ranks lower).

---

### Test: Staleness entries ordered by rank then name

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/z node as a leaf with outputs = [{id: "code", path: "out/z.go"}].
   - ROOT/a node as a leaf with outputs = [{id: "code", path: "out/a.go"}].
   - Neither ROOT/z nor ROOT/a depends on the other.
2. Create "out/z.go" with an artifact tag containing an outdated hash.
3. Create "out/a.go" with an artifact tag containing an outdated hash.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains `StalenessEntry` records for both ROOT/a and ROOT/z.
- Because both share the same rank (no dependency between them), the entry
  for ROOT/a appears before the entry for ROOT/z (alphabetical order).

---

## Format Errors

### Test: Format error from invalid depends_on

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a node as a leaf with depends_on = ["ROOT/missing"] (a logical
     name that does not exist in the tree).

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a with:
  - `rule` = "dependency_targets"

---

### Test: Format error from parse failure

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a whose `_node.md` contains invalid content (for example, text
     before any heading, making the node unparseable).

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` with:
  - `rule` = "parse"
  - referring to ROOT/a.
- Other nodes are still validated (the scan continues past the parse
  failure).

---

### Test: Continues after parse failure

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a whose `_node.md` has invalid content (unparseable).
   - ROOT/b as a valid leaf with outputs = [{id: "code", path: "out/b.go"}].
2. Create "out/b.go" with an artifact tag containing an outdated hash.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` for ROOT/a (rule = "parse").
- `staleness` contains a `StalenessEntry` for ROOT/b with status = "stale".
- Both are reported in the same `ValidationReport`.

---

## Cycle Detection

### Test: Simple cycle detected

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a as a leaf with depends_on = ["ROOT/b"].
   - ROOT/b as a leaf with depends_on = ["ROOT/a"].

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `cycles` is not empty.
- `cycles` contains at least one of "ROOT/a" or "ROOT/b".

---

### Test: Ranking skipped when format errors exist

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a as a leaf with depends_on = ["ROOT/missing"] (invalid target,
     causes a format error).
   - ROOT/b as a valid leaf with outputs = [{id: "code", path: "out/b.go"}].
2. Create "out/b.go" with an artifact tag containing an outdated hash.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` is not empty (contains the error for ROOT/a).
- `staleness` contains a `StalenessEntry` for ROOT/b where:
  - `rank` = 0 (default when ranking is skipped due to format errors).

---

## Edge Cases

### Test: Empty spec tree — scan fails

**Setup**

1. Do not create a `code-from-spec/` directory (the root scan target does
   not exist).

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `format_errors` contains a `FormatError` with:
  - `rule` = "scan"
- `cycles` is empty.
- `staleness` is empty.

---

### Test: Node with no outputs — not in staleness

**Setup**

1. Create a spec tree on disk with:
   - ROOT node.
   - ROOT/a as a leaf with no outputs declared.

**Actions**

Call `MCPValidateSpecs`.

**Expected outcome**

Return a `ValidationReport` where:
- `staleness` contains no `StalenessEntry` for ROOT/a.
- Staleness checking only runs for nodes that declare at least one output.
