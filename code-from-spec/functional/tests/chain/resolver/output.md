<!-- code-from-spec: SPEC/functional/tests/chain/resolver@8OqkWMHlnFoY0By388EM-Ig2Hw8 -->

## Test suite: ChainResolve

---

### Ancestors and target

---

#### TC-1: Root as target

Setup:
- Create SPEC/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC").

Expected outcome:
- ancestors = empty list.
- dependencies = empty list.
- target = ChainItem(unqualified_logical_name="SPEC", qualifier=absent).
- input = absent.

---

#### TC-2: Linear chain — ancestors in root-first order

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.
- Create SPEC/a/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a/b").

Expected outcome:
- ancestors = [ChainItem("SPEC"), ChainItem("SPEC/a")] in that order.
- target = ChainItem(unqualified_logical_name="SPEC/a/b", qualifier=absent).

---

#### TC-3: Single parent

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- ancestors = [ChainItem("SPEC")].
- target = ChainItem(unqualified_logical_name="SPEC/a", qualifier=absent).

---

#### TC-4: Target with empty frontmatter

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- ancestors = [ChainItem("SPEC")].
- target = ChainItem(unqualified_logical_name="SPEC/a", qualifier=absent).
- dependencies = empty list.
- input = absent.

---

### Dependencies — SPEC/ references

---

#### TC-5: Dependency without qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="SPEC/b", qualifier=absent.

---

#### TC-6: Dependency with qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="SPEC/b", qualifier="interface".

---

#### TC-7: Dependencies sorted by logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/z", "SPEC/m", "SPEC/b"].
- Create SPEC/z/_node.md with empty frontmatter.
- Create SPEC/m/_node.md with empty frontmatter.
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies = [ChainItem("SPEC/b"), ChainItem("SPEC/m"), ChainItem("SPEC/z")] in that order.

---

### Dependencies — ARTIFACT/ references

---

#### TC-8: ARTIFACT dependency resolved from generating node

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="ARTIFACT/b", file_path="out/lib.go", qualifier=absent.

---

#### TC-9: ARTIFACT — generating node has no output

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with empty frontmatter (no output field).

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Error UnresolvableArtifact is raised.

---

#### TC-10: ARTIFACT — artifact file does not exist on disk

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".
- Do NOT create "out/lib.go" on disk.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- No error is raised.
- dependencies contains one ChainItem with unqualified_logical_name="ARTIFACT/b", file_path="out/lib.go".

---

#### TC-11: Mixed SPEC/, ARTIFACT/, and EXTERNAL/ dependencies

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/c", "ARTIFACT/b", "EXTERNAL/proto/api.proto"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".
- Create SPEC/c/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies = [ChainItem("ARTIFACT/b"), ChainItem("EXTERNAL/proto/api.proto"), ChainItem("SPEC/c")] in that order (sorted by logical name).

---

### Dependencies — dedup

---

#### TC-12: Exact duplicate — same file, same qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b", "SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ChainItem for SPEC/b.

---

#### TC-13: No qualifier subsumes qualifier

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b", "SPEC/b(interface)"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ChainItem with unqualified_logical_name="SPEC/b", qualifier=absent.

---

#### TC-14: Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)", "SPEC/b"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ChainItem with unqualified_logical_name="SPEC/b", qualifier=absent.

---

#### TC-15: Same file, different qualifiers — both kept

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["SPEC/b(interface)", "SPEC/b(constraints)"].
- Create SPEC/b/_node.md with empty frontmatter.

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains two ChainItems: one with qualifier="constraints" and one with qualifier="interface" (sorted by qualifier).

---

#### TC-16: Duplicate ARTIFACT — same logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["ARTIFACT/b", "ARTIFACT/b"].
- Create SPEC/b/_node.md with frontmatter: output = "out/lib.go".

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ChainItem for ARTIFACT/b.

---

### Dependencies — EXTERNAL/ references

---

#### TC-17: EXTERNAL dependency resolved to path

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/docs/api.yaml"].

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains one ChainItem with unqualified_logical_name="EXTERNAL/docs/api.yaml", file_path="docs/api.yaml", qualifier=absent.

---

#### TC-18: Multiple EXTERNAL dependencies sorted

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/proto/v1.proto", "EXTERNAL/docs/api.yaml"].

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies = [ChainItem("EXTERNAL/docs/api.yaml"), ChainItem("EXTERNAL/proto/v1.proto")] in that order.

---

#### TC-19: Duplicate EXTERNAL — same logical name

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["EXTERNAL/x.proto", "EXTERNAL/x.proto"].

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- dependencies contains exactly one ChainItem for EXTERNAL/x.proto.

---

### Input

---

#### TC-20: Input resolved from generating node

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: input = "ARTIFACT/b".
- Create SPEC/b/_node.md with frontmatter: output = "out/data.json".

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = ChainItem(unqualified_logical_name="ARTIFACT/b", file_path="out/data.json", qualifier=absent).

---

#### TC-21: EXTERNAL input resolved to path

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: input = "EXTERNAL/docs/vendor/spec.yaml".

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = ChainItem(unqualified_logical_name="EXTERNAL/docs/vendor/spec.yaml", file_path="docs/vendor/spec.yaml", qualifier=absent).

---

#### TC-22: No input — absent

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with empty frontmatter (no input field).

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- input = absent.

---

### Error cases

---

#### TC-23: Unrecognized prefix in depends_on

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with frontmatter: depends_on = ["UNKNOWN/something"].

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Error UnresolvableArtifact is raised.

---

#### TC-24: Invalid target logical name

Setup:
- No spec tree required.

Action:
- Call ChainResolve("INVALID/something").

Expected outcome:
- Error propagated from LogicalNameGetParent or LogicalNameToPath is raised.

---

#### TC-25: Unreadable frontmatter

Setup:
- Create SPEC/_node.md with empty frontmatter.
- Create SPEC/a/_node.md with invalid YAML content (malformed frontmatter that cannot be parsed).

Action:
- Call ChainResolve("SPEC/a").

Expected outcome:
- Error UnreadableFrontmatter is raised.
