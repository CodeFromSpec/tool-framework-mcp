<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@aHTAVtTvOkd6TgaMg940xay023M -->

# Test Specification: NodeRankCompute

## Interface

```
record NodeRankInput
  logical_name: string
  frontmatter: Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference
```

---

## Happy Path Tests

---

### TC-01: Root only

**Description:** A single entry with no dependencies or children should receive rank 0.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ranked` contains exactly one entry: `NodeRankEntry { logical_name: "ROOT", rank: 0 }`.
- `cycles` is empty.

---

### TC-02: Linear chain — incrementing ranks

**Description:** A parent-child chain with no explicit dependencies should produce incrementing ranks.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",       frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a",     frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a/b",   frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ranked` contains entries with:
  - `"ROOT"` has rank 0.
  - `"ROOT/a"` has rank 1.
  - `"ROOT/a/b"` has rank 2.
- `cycles` is empty.

---

### TC-03: Independent siblings — equal rank

**Description:** Sibling nodes with no cross-dependencies should receive the same rank.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `"ROOT"` has rank 0.
- `"ROOT/a"` and `"ROOT/b"` both have rank 1.
- `cycles` is empty.

---

### TC-04: depends_on increases rank

**Description:** A node whose `depends_on` points to a sibling must rank higher than that sibling.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `"ROOT/b"` has a higher rank than `"ROOT/a"`.
- `cycles` is empty.

---

### TC-05: depends_on with qualifier — qualifier stripped

**Description:** A qualified reference like `"ROOT/a(interface)"` must resolve to the bare node `"ROOT/a"`.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a(interface)"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- No error is raised.
- `"ROOT/b"` has a higher rank than `"ROOT/a"`.
- `cycles` is empty.

---

### TC-06: input artifact adds dependency edge

**Description:** A node whose `input` references an artifact must depend on that artifact, which in turn depends on the node that produces it.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "code", path: "out.go" }] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { input: "ARTIFACT/a(code)" } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/b")` > `rank("ARTIFACT/a(code)")` > `rank("ROOT/a")`.
- `cycles` is empty.

---

### TC-07: Artifacts get rank one above their node

**Description:** An artifact's rank is exactly one greater than the rank of the node that declares it.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "foo", path: "foo.go" }] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ARTIFACT/a(foo)")` = `rank("ROOT/a")` + 1.
- `cycles` is empty.

---

### TC-08: Multiple outputs — each artifact ranked

**Description:** When a node declares multiple outputs, each artifact appears as a separate ranked entry.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "x", path: "x.go" }, { id: "y", path: "y.go" }] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ranked` contains entries for both `"ARTIFACT/a(x)"` and `"ARTIFACT/a(y)"`.
- Both have rank = `rank("ROOT/a")` + 1.
- `cycles` is empty.

---

### TC-09: depends_on ARTIFACT reference — used as-is

**Description:** A `depends_on` entry that already uses the ARTIFACT prefix resolves directly to that artifact.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "lib", path: "lib.go" }] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ARTIFACT/a(lib)"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/b")` > `rank("ARTIFACT/a(lib)")` > `rank("ROOT/a")`.
- `cycles` is empty.

---

### TC-10: Output sorted by rank then logical name

**Description:** The returned `ranked` list is sorted first by ascending rank, then alphabetically by logical name within the same rank.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/z", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- The first entry in `ranked` is `"ROOT"` (rank 0).
- `"ROOT/a"` appears before `"ROOT/z"` (both rank 1, alphabetical order).
- `cycles` is empty.

---

### TC-11: Parallel entries — equal rank means no dependency

**Description:** Multiple siblings with no cross-dependencies all share the same rank.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/c", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `"ROOT/a"`, `"ROOT/b"`, and `"ROOT/c"` all have rank 1.
- `cycles` is empty.

---

### TC-12: Diamond dependency — rank uses max not sum

**Description:** When a node depends on two paths that converge, its rank is 1 + max(dependency ranks), not their sum.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/c", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/c"] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/c"] } }
  NodeRankInput { logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a", "ROOT/b"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/c")` = 1.
- `rank("ROOT/a")` = 2.
- `rank("ROOT/b")` = 2.
- `rank("ROOT/d")` = 3 (not 5).
- `cycles` is empty.

---

### TC-13: depends_on outranks parent

**Description:** A `depends_on` target at a deep level can push a node's rank above what its parent path alone would give.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",       frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a",     frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a/b",   frontmatter: { depends_on: ["ROOT/c"] } }
  NodeRankInput { logical_name: "ROOT/c",     frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/c/d",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/c/d/e", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/a/b")` > `rank("ROOT/a")` (parent contributes rank 2, depends_on ROOT/c contributes at least rank 1).
- `rank("ROOT/a/b")` = 1 + max(`rank("ROOT/a")`, `rank("ROOT/c")`).
- `cycles` is empty.

---

### TC-14: Multiple depends_on — rank from highest

**Description:** When a node lists multiple `depends_on` entries that have different ranks, the node's rank is based on the highest dependency rank.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
  NodeRankInput { logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"] } }
  NodeRankInput { logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a", "ROOT/b", "ROOT/c"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/a")` = 1.
- `rank("ROOT/b")` = 2.
- `rank("ROOT/c")` = 3.
- `rank("ROOT/d")` = 4 (1 + max(1, 2, 3) = 4).
- `cycles` is empty.

---

### TC-15: Node with both depends_on and input

**Description:** When a node has both `depends_on` and an `input` artifact reference, the rank is determined by the highest of: parent rank, all depends_on ranks, and the input artifact rank.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "out", path: "a.go" }] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"], input: "ARTIFACT/a(out)" } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `rank("ROOT/c")` = 1 + max(`rank("ROOT")`, `rank("ROOT/b")`, `rank("ARTIFACT/a(out)")`).
- `cycles` is empty.

---

### TC-16: Empty input list

**Description:** Calling the function with an empty list should return empty results with no error.

**Setup:**
```
entries = []
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ranked` is empty.
- `cycles` is empty.

---

## Cycle Detection Tests

---

### TC-17: Self-reference

**Description:** A node that lists itself in `depends_on` forms a cycle of length one.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/a"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `cycles` is not empty.

---

### TC-18: Simple cycle — two nodes

**Description:** Two nodes that depend on each other form a direct cycle.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `cycles` is not empty.
- `cycles` contains at least one of `"ROOT/a"` or `"ROOT/b"`.

---

### TC-19: Cycle through artifacts

**Description:** A cycle can form through artifact references when two nodes depend on each other's artifacts.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { outputs: [{ id: "out", path: "a.go" }], depends_on: ["ARTIFACT/b(out)"] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { outputs: [{ id: "out", path: "b.go" }], depends_on: ["ARTIFACT/a(out)"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `cycles` is not empty.

---

### TC-20: Cycle does not prevent ranking of unrelated nodes

**Description:** Nodes that are not part of a cycle should still receive valid ranks even when a cycle exists elsewhere in the graph.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"] } }
  NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
  NodeRankInput { logical_name: "ROOT/c", frontmatter: (empty) }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `"ROOT"` has rank 0 and `"ROOT/c"` has rank 1.
- `cycles` contains entries related to `"ROOT/a"` and/or `"ROOT/b"`, but does not contain `"ROOT/c"`.

---

## Error Case Tests

---

### TC-21: Unresolvable ROOT reference

**Description:** A `depends_on` entry that points to a node not present in the input list must raise an error.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/missing"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Error `UnresolvableReference` is raised.

---

### TC-22: Unresolvable ARTIFACT reference

**Description:** A `depends_on` entry using the ARTIFACT prefix that refers to an artifact not produced by any node in the input list must raise an error.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ARTIFACT/missing(id)"] } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Error `UnresolvableReference` is raised.

---

### TC-23: Unresolvable input reference

**Description:** An `input` field that references an artifact not produced by any node in the input list must raise an error.

**Setup:**
```
entries = [
  NodeRankInput { logical_name: "ROOT",   frontmatter: (empty) }
  NodeRankInput { logical_name: "ROOT/a", frontmatter: { input: "ARTIFACT/missing(id)" } }
]
```

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Error `UnresolvableReference` is raised.
