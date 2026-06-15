<!-- code-from-spec: SPEC/functional/tests/mcp_tools/validate_specs@wUlMnVgDmlCwDl7NHL9AzHJGcbk -->

# Test Specification: MCPValidateSpecs

## Records

```
record StalenessEntry
  node: string
  artifact_path: string
  status: string
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of spectreevalidate.FormatError
  cycles: list of string
  staleness: list of StalenessEntry
```

---

## Happy Path

### TC-HP-01: Clean tree — no errors

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Compute the current chain hash for SPEC/a using `ChainHashCompute`.
  Create `out/a.go` containing a valid artifact tag with that hash.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` with:
    - `format_errors` is empty.
    - `cycles` is empty.
    - `staleness` is empty.

---

### TC-HP-02: Stale artifact detected

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Create `out/a.go` containing an artifact tag with a 27-character
    base64url string that differs from the current chain hash of SPEC/a.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` contains exactly one `StalenessEntry` with:
        - `node` = `"SPEC/a"`
        - `status` = `"stale"`
        - `rank` is present (integer).

---

### TC-HP-03: Missing artifact detected

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Do not create `out/a.go`.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` contains exactly one `StalenessEntry` with:
        - `node` = `"SPEC/a"`
        - `status` = `"missing"`.

---

### TC-HP-04: Malformed tag detected

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Create `out/a.go` with content that has no artifact tag
    (or contains a syntactically malformed artifact tag).

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` contains exactly one `StalenessEntry` with:
        - `node` = `"SPEC/a"`
        - `status` = `"malformed tag"`.

---

### TC-HP-05: Staleness entries include rank

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Create `code-from-spec/b/_node.md` with first heading `# SPEC/b`.
    Set frontmatter `output: "out/b.go"` and `depends_on: ["SPEC/a"]`.
  Create `out/a.go` and `out/b.go` each with artifact tags containing
    a 27-character base64url string that differs from their respective
    current chain hashes.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` contains two `StalenessEntry` records, one for
      `"SPEC/a"` and one for `"SPEC/b"`.
    - Each entry has a `rank` integer value.
    - The `rank` of `"SPEC/a"` is strictly less than the `rank` of `"SPEC/b"`.

---

### TC-HP-06: Staleness ordered by rank then name

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/z/_node.md` with first heading `# SPEC/z`.
    Set frontmatter `output: "out/z.go"`.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `output: "out/a.go"`.
  Create `out/z.go` and `out/a.go` each with artifact tags containing
    a 27-character base64url string that differs from their respective
    current chain hashes.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` has two entries.
    - The entry for `"SPEC/a"` appears before the entry for `"SPEC/z"`
      (same rank, alphabetical order by node name).

---

## Format Errors

### TC-FE-01: Format error from invalid depends_on

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `depends_on: ["SPEC/missing"]`.
    "SPEC/missing" does not correspond to any real node.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains at least one `spectreevalidate.FormatError`
      for `"SPEC/a"` with `rule` = `"dependency_targets"`.

---

### TC-FE-02: Format error from parse failure

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with content that fails parsing
    (e.g., plain text before any heading, making it structurally invalid).

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains at least one `spectreevalidate.FormatError`
      for `"SPEC/a"` with `rule` = `"parse"`.
    - The function does not raise an error — it continues validating
      other nodes.

---

### TC-FE-03: Continues after parse failure

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with invalid content that fails
    parsing (plain text before any heading).
  Create `code-from-spec/b/_node.md` with first heading `# SPEC/b`.
    Set frontmatter `output: "out/b.go"`.
  Create `out/b.go` with an artifact tag containing a 27-character
    base64url string that differs from the current chain hash of SPEC/b.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains a `spectreevalidate.FormatError` for `"SPEC/a"`
      with `rule` = `"parse"`.
    - `staleness` contains a `StalenessEntry` for `"SPEC/b"` with
      `status` = `"stale"`.
    - Both issues are reported together in the same report.

---

### TC-FE-04: Subdirectory without _node.md detected

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
  Create directory `code-from-spec/b/` with no `_node.md` inside.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains at least one `spectreevalidate.FormatError`
      for the `code-from-spec/b/` directory with `rule` = `"missing_node_md"`.

---

### TC-FE-05: _-prefixed dir under code-from-spec not flagged

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create directory `code-from-spec/_tools/` with no `_node.md` inside.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains no entry for `code-from-spec/_tools/`.
    - `_`-prefixed directories directly under `code-from-spec/` are ignored.

---

## Cycle Detection

### TC-CD-01: Simple cycle detected

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `depends_on: ["SPEC/b"]`.
  Create `code-from-spec/b/_node.md` with first heading `# SPEC/b`.
    Set frontmatter `depends_on: ["SPEC/a"]`.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `cycles` is not empty.
    - `cycles` contains at least one of `"SPEC/a"` or `"SPEC/b"`.

---

### TC-CD-02: Ranking skipped when format errors exist

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Set frontmatter `depends_on: ["SPEC/missing"]`.
    "SPEC/missing" does not exist.
  Create `code-from-spec/b/_node.md` with first heading `# SPEC/b`.
    Set frontmatter `output: "out/b.go"`.
  Create `out/b.go` with an artifact tag containing a 27-character
    base64url string that differs from the current chain hash of SPEC/b.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` is not empty (contains an error for `"SPEC/a"`).
    - Ranking is skipped due to format errors.
    - Any `StalenessEntry` for `"SPEC/b"` has `rank` = 0
      (default when no ranking is available).

---

## Edge Cases

### TC-EC-01: Empty spec tree — scan fails

Setup:
  Do not create a `code-from-spec/` directory.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `format_errors` contains at least one `spectreevalidate.FormatError`
      with `rule` = `"scan"`.
    - `cycles` is empty.
    - `staleness` is empty.

---

### TC-EC-02: Node with no output — not in staleness

Setup:
  Create `code-from-spec/_node.md` with first heading `# SPEC`.
    Include a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with first heading `# SPEC/a`.
    Do not set an `output` field in its frontmatter.

Action:
  Call `MCPValidateSpecs`.

Expected:
  Result is a `ValidationReport` where:
    - `staleness` contains no `StalenessEntry` for `"SPEC/a"`.
    - Staleness checks only run for nodes that declare an output.
