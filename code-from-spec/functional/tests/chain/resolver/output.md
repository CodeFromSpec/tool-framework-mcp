<!-- code-from-spec: ROOT/functional/tests/chain/resolver@FhtyO5A322vECd1NAmvZxpg_Jk4 -->

# Tests: ChainResolve

## Ancestors and Target

### Root as target

Setup:
- Create SPEC/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC").

Expected outcome:
- ancestors = empty list.
- target = ChainItem(unqualified_logical_name="SPEC", qualifier=absent).
- dependencies = empty list.
- input = absent.

---

### Linear chain — ancestors in root-first order

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.
- Create SPEC/a/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a/b").

Expected outcome:
- ancestors = [ChainItem("SPEC"), ChainItem("SPEC/a")] in that order.
- target = ChainItem(unqualified_logical_name="SPEC/a/b", qualifier=absent).
- dependencies = empty list.
- input = absent.

---

### Single parent

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- ancestors = [ChainItem("SPEC")].
- target = ChainItem(unqualified_logical_name="SPEC/a", qualifier=absent).
- dependencies = empty list.
- input = absent.

---

### Target with empty frontmatter

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- ancestors = [ChainItem("SPEC")].
- target = ChainItem(unqualified_logical_name="SPEC/a", qualifier=absent).
- dependencies = empty list.
- input = absent.

---

## Dependencies — SPEC/ References

### Dependency without qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="SPEC/b", qualifier=absent.

---

### Dependency with qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="SPEC/b", qualifier="interface".

---

### Dependencies sorted by logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/z", "SPEC/m", "SPEC/b"].
- Create SPEC/z/_node.md with empty frontmatter.
- Create SPEC/m/_node.md with empty frontmatter.
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies = [ChainItem("SPEC/b"), ChainItem("SPEC/m"), ChainItem("SPEC/z")] in that order (alphabetical by logical name).

---

## Dependencies — ARTIFACT/ References

### ARTIFACT dependency resolved from generating node

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="ARTIFACT/b", file_path="out/lib.go", qualifier=absent.

---

### ARTIFACT — generating node has no output

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with empty frontmatter (no output field).

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Raises error UnresolvableArtifact.

---

### ARTIFACT — artifact file does not exist on disk

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".
- Do NOT create "out/lib.go" on disk.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- No error raised.
- dependencies contains one ChainItem with file_path="out/lib.go".

---

### Mixed SPEC/, ARTIFACT/, and EXTERNAL/ dependencies

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/c", "ARTIFACT/b", "EXTERNAL/proto/api.proto"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".
- Create SPEC/c/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains three entries sorted by logical name:
  1. ChainItem(unqualified_logical_name="ARTIFACT/b", file_path="out/lib.go", qualifier=absent).
  2. ChainItem(unqualified_logical_name="EXTERNAL/proto/api.proto", file_path="proto/api.proto", qualifier=absent).
  3. ChainItem(unqualified_logical_name="SPEC/c", qualifier=absent).

---

## Dependencies — Dedup

### Exact duplicate — same file, same qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b", "SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one entry for SPEC/b (not two).

---

### No qualifier subsumes qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b", "SPEC/b(interface)"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one entry for SPEC/b with qualifier=absent.
- The "SPEC/b(interface)" entry is dropped.

---

### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)", "SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one entry for SPEC/b with qualifier=absent.

---

### Same file, different qualifiers — both kept

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)", "SPEC/b(constraints)"].
- Create SPEC/b/_node.md with empty frontmatter.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains two entries:
  1. ChainItem(unqualified_logical_name="SPEC/b", qualifier="constraints").
  2. ChainItem(unqualified_logical_name="SPEC/b", qualifier="interface").
  (sorted by qualifier alphabetically)

---

### Duplicate ARTIFACT — same logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b", "ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ARTIFACT/b entry (not two).

---

## Dependencies — EXTERNAL/ References

### EXTERNAL dependency resolved to path

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/docs/api.yaml"].

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="EXTERNAL/docs/api.yaml", file_path="docs/api.yaml", qualifier=absent.

---

### Multiple EXTERNAL dependencies sorted

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/proto/v1.proto", "EXTERNAL/docs/api.yaml"].

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies sorted by logical name:
  1. ChainItem(unqualified_logical_name="EXTERNAL/docs/api.yaml", file_path="docs/api.yaml").
  2. ChainItem(unqualified_logical_name="EXTERNAL/proto/v1.proto", file_path="proto/v1.proto").

---

### Duplicate EXTERNAL — same logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/x.proto", "EXTERNAL/x.proto"].

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one EXTERNAL/x.proto entry (not two).

---

## Input

### Input resolved from generating node

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: input = "ARTIFACT/b".
- Create SPEC/b/_node.md with frontmatter: output = "out/data.json".

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = ChainItem(unqualified_logical_name="ARTIFACT/b", file_path="out/data.json", qualifier=absent).

---

### EXTERNAL input resolved to path

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: input = "EXTERNAL/docs/vendor/spec.yaml".

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = ChainItem(unqualified_logical_name="EXTERNAL/docs/vendor/spec.yaml", file_path="docs/vendor/spec.yaml", qualifier=absent).

---

### No input — absent

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter (no input field).

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = absent.

---

## Error Cases

### Unrecognized prefix in depends_on

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["UNKNOWN/something"].

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Raises error UnresolvableArtifact.

---

### Invalid target logical name

Setup:
- No spec tree required.

Actions:
- Call ChainResolve("INVALID/something").

Expected outcome:
- Raises an error propagated from LogicalNameGetParent or LogicalNameToPath.

---

### Unreadable frontmatter

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with invalid YAML content in the frontmatter block.

Actions:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Raises error UnreadableFrontmatter.
