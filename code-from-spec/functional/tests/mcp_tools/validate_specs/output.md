<!-- code-from-spec: ROOT/functional/tests/mcp_tools/validate_specs@zwInsrFHRMMh5iL3_xzHos7K9gw -->

# Test Specification: MCPValidateSpecs

## Happy Path

### TC-01: Clean tree — no errors

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC` and a `# Public` section containing a `## Context` subsection.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Compute the current chain hash for node `SPEC/a` using `ChainHashCompute`.
  Create `out/a.go` with a valid artifact tag embedding that hash.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where:
  - `format_errors` is empty.
  - `cycles` is empty.
  - `staleness` is empty.

---

### TC-02: Stale artifact detected

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Create `out/a.go` with an artifact tag containing a 27-character base64url string that differs from the current chain hash for `SPEC/a`.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` contains a `StalenessEntry` for `SPEC/a` with:
  - `status` = `"stale"`.
  - `rank` is an integer (present).

---

### TC-03: Missing artifact detected

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Do not create `out/a.go`.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` contains a `StalenessEntry` for `SPEC/a` with:
  - `status` = `"missing"`.

---

### TC-04: Malformed tag detected

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Create `out/a.go` with content that contains no artifact tag (or an artifact tag that cannot be parsed).

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` contains a `StalenessEntry` for `SPEC/a` with:
  - `status` = `"malformed tag"`.

---

### TC-05: Staleness entries include rank

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Create `code-from-spec/b/_node.md` with heading `# SPEC/b`, frontmatter `output: out/b.go`, and `depends_on: ["SPEC/a"]`.
  Create `out/a.go` and `out/b.go` with artifact tags containing a 27-character base64url string that differs from their respective current chain hashes.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` contains `StalenessEntry` records for both `SPEC/a` and `SPEC/b`.
  The `rank` for `SPEC/a` is strictly lower than the `rank` for `SPEC/b`.

---

### TC-06: Staleness ordered by rank then name

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/z/_node.md` with heading `# SPEC/z` and frontmatter `output: out/z.go`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `output: out/a.go`.
  Create `out/z.go` and `out/a.go` with artifact tags containing stale hashes.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` has `SPEC/a` appearing before `SPEC/z`.
  Both entries have the same rank. Ordering between equal-rank entries is alphabetical by node name.

---

## Format Errors

### TC-07: Format error from invalid depends_on

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `depends_on: ["SPEC/missing"]`.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `format_errors` contains a `spectreevalidate.FormatError` for `SPEC/a` with:
  - `rule` = `"dependency_targets"`.

---

### TC-08: Format error from parse failure

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with content that causes a parse failure (e.g., text before any heading).

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `format_errors` contains a `spectreevalidate.FormatError` for `SPEC/a` with:
  - `rule` = `"parse"`.
  Other nodes that parse successfully are still validated.

---

### TC-09: Continues after parse failure

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with invalid content (causes parse failure).
  Create `code-from-spec/b/_node.md` with heading `# SPEC/b` and frontmatter `output: out/b.go`.
  Create `out/b.go` with a stale artifact tag.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where:
  - `format_errors` contains a `spectreevalidate.FormatError` for `SPEC/a`.
  - `staleness` contains a `StalenessEntry` for `SPEC/b`.
  Both issues are reported in the same report.

---

### TC-10: Subdirectory without _node.md detected

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a`.
  Create an empty directory `code-from-spec/b/` with no `_node.md` inside.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `format_errors` contains a `spectreevalidate.FormatError` with:
  - `rule` = `"missing_node_md"`.
  - The error references the `code-from-spec/b/` directory.

---

### TC-11: _-prefixed dir under code-from-spec not flagged

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create a directory `code-from-spec/_tools/` with no `_node.md` inside.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `format_errors` contains no `spectreevalidate.FormatError` referencing `_tools/`.
  The `_`-prefixed directory is silently ignored.

---

## Cycle Detection

### TC-12: Simple cycle detected

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `depends_on: ["SPEC/b"]`.
  Create `code-from-spec/b/_node.md` with heading `# SPEC/b` and frontmatter `depends_on: ["SPEC/a"]`.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `cycles` is not empty and contains at least one of `"SPEC/a"` or `"SPEC/b"`.

---

### TC-13: Ranking skipped when format errors exist

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and frontmatter `depends_on: ["SPEC/missing"]` (invalid dependency target).
  Create `code-from-spec/b/_node.md` with heading `# SPEC/b` and frontmatter `output: out/b.go`.
  Create `out/b.go` with a stale artifact tag.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where:
  - `format_errors` is not empty (contains the error for `SPEC/a`).
  - `staleness` contains a `StalenessEntry` for `SPEC/b` with `rank` = `0` (ranking was skipped due to format errors).

---

## Edge Cases

### TC-14: Empty spec tree — scan fails

Setup:
  Do not create a `code-from-spec/` directory.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where:
  - `format_errors` contains a `spectreevalidate.FormatError` with `rule` = `"scan"`.
  - `cycles` is empty.
  - `staleness` is empty.

---

### TC-15: Node with no output — not in staleness

Setup:
  Create `code-from-spec/_node.md` with heading `# SPEC`.
  Create `code-from-spec/a/_node.md` with heading `# SPEC/a` and no `output` frontmatter field.

Action:
  Call `MCPValidateSpecs()`.

Expected outcome:
  Returns a `ValidationReport` where `staleness` contains no `StalenessEntry` for `SPEC/a`.
  Staleness checks are only performed for nodes that declare an output.
