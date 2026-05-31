<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@-_XUlutMyUbIwYLtauSyP6d4Ops -->

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

## Happy Path

---

### TC-01: Root only

**Description:** Single entry with no dependencies or parent yields rank 0.

**Setup:**
- entries:
  - `{ logical_name: "ROOT", frontmatter: { depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- ranked contains exactly one entry: `{ logical_name: "ROOT", rank: 0 }`.
- cycles is empty.

---

### TC-02: Linear chain — incrementing ranks

**Description:** A parent chain with no explicit dependencies yields monotonically increasing ranks.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",      frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a",    frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a/b",  frontmatter: { depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ROOT` has rank 0.
- `ROOT/a` has rank 1.
- `ROOT/a/b` has rank 2.
- cycles is empty.

---

### TC-03: Independent siblings — equal rank

**Description:** Two sibling nodes with no cross-dependencies share the same rank.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ROOT/a` and `ROOT/b` have the same rank (1).
- cycles is empty.

---

### TC-04: depends_on increases rank

**Description:** A node that depends on a sibling is ranked higher than that sibling.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/b` > rank of `ROOT/a`.
- cycles is empty.

---

### TC-05: depends_on with qualifier — qualifier stripped

**Description:** A qualified reference such as `"ROOT/a(interface)"` resolves to the bare node `ROOT/a`.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a(interface)"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- No error is raised.
- rank of `ROOT/b` > rank of `ROOT/a`.
- The qualified reference resolved to the bare node `ROOT/a`.
- cycles is empty.

---

### TC-06: input artifact adds dependency edge

**Description:** A node whose frontmatter `input` references an artifact depends on both the artifact and its source node.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],                outputs: [{ id: "code", path: "out.go" }] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { input: "ARTIFACT/a(code)",     depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/b` > rank of artifact `ARTIFACT/a(code)`.
- rank of `ARTIFACT/a(code)` > rank of `ROOT/a`.
- cycles is empty.

---

### TC-07: Artifacts get rank one above their node

**Description:** Each output artifact is ranked exactly one above the node that declares it.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [], outputs: [{ id: "foo", path: "foo.go" }] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- ranked contains an entry for `ARTIFACT/a(foo)`.
- rank of `ARTIFACT/a(foo)` = rank of `ROOT/a` + 1.
- cycles is empty.

---

### TC-08: Multiple outputs — each artifact ranked

**Description:** Each output artifact declared by a node receives its own ranked entry.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [], outputs: [{ id: "x", path: "x.go" }, { id: "y", path: "y.go" }] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- ranked contains entries for both `ARTIFACT/a(x)` and `ARTIFACT/a(y)`.
- rank of `ARTIFACT/a(x)` = rank of `ROOT/a` + 1.
- rank of `ARTIFACT/a(y)` = rank of `ROOT/a` + 1.
- cycles is empty.

---

### TC-09: depends_on ARTIFACT reference — used as-is

**Description:** A node may declare a direct dependency on an artifact logical name.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                  outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],                  outputs: [{ id: "lib", path: "lib.go" }] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ARTIFACT/a(lib)"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/b` > rank of `ARTIFACT/a(lib)`.
- rank of `ARTIFACT/a(lib)` > rank of `ROOT/a`.
- cycles is empty.

---

### TC-10: Output sorted by rank then logical name

**Description:** The returned ranked list is ordered first by ascending rank, then alphabetically by logical name within the same rank.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/z", frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- First entry in ranked: `ROOT` (rank 0).
- Second entry: `ROOT/a` (rank 1).
- Third entry: `ROOT/z` (rank 1).
- cycles is empty.

---

### TC-11: Parallel entries — equal rank means no dependency

**Description:** Multiple siblings with no cross-dependencies all receive the same rank.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: [], outputs: [] } }`
  - `{ logical_name: "ROOT/c", frontmatter: { depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ROOT/a`, `ROOT/b`, and `ROOT/c` all have rank 1.
- cycles is empty.

---

### TC-12: Diamond dependency — rank uses max not sum

**Description:** When a node has two paths converging on it, its rank is 1 + max of dependency ranks, not the sum.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/c", frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/c"],            outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/c"],            outputs: [] } }`
  - `{ logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a", "ROOT/b"],  outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/c` = 1.
- rank of `ROOT/a` = 2.
- rank of `ROOT/b` = 2.
- rank of `ROOT/d` = 3 (not 5).
- cycles is empty.

---

### TC-13: depends_on outranks parent

**Description:** When a node's explicit dependency resolves to a higher rank than its structural parent, the dependency rank wins.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",       frontmatter: { depends_on: [],          outputs: [] } }`
  - `{ logical_name: "ROOT/a",     frontmatter: { depends_on: [],          outputs: [] } }`
  - `{ logical_name: "ROOT/a/b",   frontmatter: { depends_on: ["ROOT/c"],  outputs: [] } }`
  - `{ logical_name: "ROOT/c",     frontmatter: { depends_on: [],          outputs: [] } }`
  - `{ logical_name: "ROOT/c/d",   frontmatter: { depends_on: [],          outputs: [] } }`
  - `{ logical_name: "ROOT/c/d/e", frontmatter: { depends_on: [],          outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/a/b` > rank of `ROOT/a`.
- rank of `ROOT/a/b` = 1 + max(rank of `ROOT/a`, rank of `ROOT/c`).
- cycles is empty.

---

### TC-14: Multiple depends_on — rank from highest

**Description:** When multiple depends_on entries exist at different ranks, the node's rank is 1 + max of all dependency ranks.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                          outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],                          outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"],                  outputs: [] } }`
  - `{ logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"],                  outputs: [] } }`
  - `{ logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a","ROOT/b","ROOT/c"],outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/a` = 1.
- rank of `ROOT/b` = 2.
- rank of `ROOT/c` = 3.
- rank of `ROOT/d` = 4 (1 + max(1, 2, 3) = 4, not based on first or last).
- cycles is empty.

---

### TC-15: Node with both depends_on and input

**Description:** When a node has both `depends_on` and `input`, its rank is determined by the maximum of all edges: structural parent, depends_on targets, and input artifact.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: [],         outputs: [{ id: "out", path: "a.go" }] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"], input: "ARTIFACT/a(out)", outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- rank of `ROOT/c` = 1 + max(rank of `ROOT` (parent), rank of `ROOT/b`, rank of `ARTIFACT/a(out)`).
- cycles is empty.

---

### TC-16: Empty input list

**Description:** An empty entries list produces an empty result.

**Setup:**
- entries: (empty list)

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- ranked is an empty list.
- cycles is empty.

---

## Cycle Detection

---

### TC-17: Self-reference

**Description:** A node that depends on itself forms a cycle.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/a"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- cycles is not empty.

---

### TC-18: Simple cycle — two nodes

**Description:** Two nodes each depending on the other form a cycle.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"], outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- cycles is not empty.
- cycles contains at least one of `"ROOT/a"` or `"ROOT/b"`.

---

### TC-19: Cycle through artifacts

**Description:** A cycle can be formed through artifact references between two nodes.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ARTIFACT/b(out)"],   outputs: [{ id: "out", path: "a.go" }] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ARTIFACT/a(out)"],   outputs: [{ id: "out", path: "b.go" }] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- cycles is not empty.

---

### TC-20: Cycle does not prevent ranking of unrelated nodes

**Description:** Nodes not involved in a cycle are still assigned valid ranks even when a cycle exists elsewhere in the graph.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],         outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"], outputs: [] } }`
  - `{ logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"], outputs: [] } }`
  - `{ logical_name: "ROOT/c", frontmatter: { depends_on: [],         outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- `ROOT` has rank 0.
- `ROOT/c` has rank 1.
- cycles contains entries related to `ROOT/a` and/or `ROOT/b`, but not `ROOT/c`.

---

## Error Cases

---

### TC-21: Unresolvable ROOT reference

**Description:** A depends_on entry that references a node not present in the input raises an error.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],               outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/missing"], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Raises error `UnresolvableReference`.

---

### TC-22: Unresolvable ARTIFACT reference

**Description:** A depends_on entry that references an artifact whose source node does not exist raises an error.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                      outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { depends_on: ["ARTIFACT/missing(id)"],outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Raises error `UnresolvableReference`.

---

### TC-23: Unresolvable input reference

**Description:** A node whose `input` field references an artifact that does not exist raises an error.

**Setup:**
- entries:
  - `{ logical_name: "ROOT",   frontmatter: { depends_on: [],                    outputs: [] } }`
  - `{ logical_name: "ROOT/a", frontmatter: { input: "ARTIFACT/missing(id)",     depends_on: [], outputs: [] } }`

**Action:** Call `NodeRankCompute(entries)`.

**Expected outcome:**
- Raises error `UnresolvableReference`.
