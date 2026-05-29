<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@AJarAZGdIIkkxThBLGxZS2GGBeg -->

# Test Specification: NodeRankCompute

## Interface Reference

```
record NodeRankInput
  logical_name: string
  frontmatter: Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

function NodeRankCompute(entries: list of NodeRankInput)
  -> (ranked: list of NodeRankEntry, cycles: list of string)

errors:
  - "unresolvable reference"
```

---

## Happy Path

---

### TC-01: Root only

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked contains exactly one entry: logical_name = "ROOT", rank = 0
- cycles is empty

---

### TC-02: Linear chain — incrementing ranks

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/a/b", frontmatter = (empty)

Parent chain is inferred from logical names (no depends_on).

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- "ROOT" has rank 0
- "ROOT/a" has rank 1
- "ROOT/a/b" has rank 2
- cycles is empty

---

### TC-03: Independent siblings — equal rank

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/b", frontmatter = (empty)

No cross-dependencies.

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- "ROOT/a" and "ROOT/b" have the same rank (1)
- cycles is empty

---

### TC-04: depends_on increases rank

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/b" > rank of "ROOT/a"
- cycles is empty

---

### TC-05: depends_on with qualifier — qualifier stripped

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a(interface)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- No error raised
- The qualified reference "ROOT/a(interface)" resolves to the bare node "ROOT/a"
- rank of "ROOT/b" > rank of "ROOT/a"
- cycles is empty

---

### TC-06: input artifact adds dependency edge

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [{id: "code", path: "out.go"}]
- logical_name = "ROOT/b", frontmatter with input = "ARTIFACT/a(code)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked includes an artifact entry for "ARTIFACT/a(code)"
- rank of "ROOT/b" > rank of "ARTIFACT/a(code)"
- rank of "ARTIFACT/a(code)" > rank of "ROOT/a"
- cycles is empty

---

### TC-07: Artifacts get rank one above their node

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [{id: "foo", path: "foo.go"}]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked includes an artifact entry "ARTIFACT/a(foo)"
- rank of "ARTIFACT/a(foo)" = rank of "ROOT/a" + 1
- cycles is empty

---

### TC-08: Multiple outputs — each artifact ranked

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [
    {id: "x", path: "x.go"},
    {id: "y", path: "y.go"}
  ]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked includes two artifact entries: "ARTIFACT/a(x)" and "ARTIFACT/a(y)"
- both have rank = rank of "ROOT/a" + 1
- cycles is empty

---

### TC-09: depends_on ARTIFACT reference — used as-is

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [{id: "lib", path: "lib.go"}]
- logical_name = "ROOT/b", frontmatter with depends_on = ["ARTIFACT/a(lib)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/b" > rank of "ARTIFACT/a(lib)"
- rank of "ARTIFACT/a(lib)" > rank of "ROOT/a"
- cycles is empty

---

### TC-10: Output sorted by rank then logical name

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/z", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)

No cross-dependencies.

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Output order: "ROOT" (rank 0) appears first, then "ROOT/a" before "ROOT/z" (both rank 1, sorted alphabetically)
- cycles is empty

---

### TC-11: Parallel entries — equal rank means no dependency

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/b", frontmatter = (empty)
- logical_name = "ROOT/c", frontmatter = (empty)

All siblings, no cross-dependencies.

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- "ROOT/a", "ROOT/b", and "ROOT/c" all have rank 1
- cycles is empty

---

### TC-12: Diamond dependency — rank uses max not sum

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/c", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/c"]
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/c"]
- logical_name = "ROOT/d", frontmatter with depends_on = ["ROOT/a", "ROOT/b"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/c" = 1
- rank of "ROOT/a" = 2
- rank of "ROOT/b" = 2
- rank of "ROOT/d" = 3  (1 + max(2, 2) = 3, not 5)
- cycles is empty

---

### TC-13: depends_on outranks parent

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/a/b", frontmatter with depends_on = ["ROOT/c"]
- logical_name = "ROOT/c", frontmatter = (empty)
- logical_name = "ROOT/c/d", frontmatter = (empty)
- logical_name = "ROOT/c/d/e", frontmatter = (empty)

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/a/b" > rank of "ROOT/a"
- rank of "ROOT/a/b" = 1 + max(rank of "ROOT/a", rank of "ROOT/c")
- cycles is empty

---

### TC-14: Multiple depends_on — rank from highest

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter = (empty)
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a"]
- logical_name = "ROOT/c", frontmatter with depends_on = ["ROOT/b"]
- logical_name = "ROOT/d", frontmatter with depends_on = ["ROOT/a", "ROOT/b", "ROOT/c"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/a" = 1
- rank of "ROOT/b" = 2
- rank of "ROOT/c" = 3
- rank of "ROOT/d" = 4  (1 + max(1, 2, 3) = 4)
- cycles is empty

---

### TC-15: Node with both depends_on and input

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [{id: "out", path: "a.go"}]
- logical_name = "ROOT/b", frontmatter = (empty)
- logical_name = "ROOT/c", frontmatter with depends_on = ["ROOT/b"] and input = "ARTIFACT/a(out)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/c" = 1 + max(rank of "ROOT" (parent), rank of "ROOT/b", rank of "ARTIFACT/a(out)")
- cycles is empty

---

### TC-16: Empty input list

**Setup**

Input entries: (empty list)

**Action**

Call NodeRankCompute with the empty list.

**Expected outcome**

- ranked is an empty list
- cycles is empty

---

## Cycle Detection

---

### TC-17: Self-reference

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles is not empty

---

### TC-18: Simple cycle — two nodes

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/b"]
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles is not empty
- cycles contains at least one of "ROOT/a" or "ROOT/b"

---

### TC-19: Cycle through artifacts

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with outputs = [{id: "out", path: "a.go"}] and depends_on = ["ARTIFACT/b(out)"]
- logical_name = "ROOT/b", frontmatter with outputs = [{id: "out", path: "b.go"}] and depends_on = ["ARTIFACT/a(out)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles is not empty

---

### TC-20: Cycle does not prevent ranking of unrelated nodes

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/b"]
- logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a"]
- logical_name = "ROOT/c", frontmatter = (empty)

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- "ROOT" has a valid rank (0)
- "ROOT/c" has a valid rank (1)
- cycles contains entries related to "ROOT/a" and/or "ROOT/b"
- "ROOT/c" is not present in cycles

---

## Error Cases

---

### TC-21: Unresolvable ROOT reference

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/missing"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised

---

### TC-22: Unresolvable ARTIFACT reference

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with depends_on = ["ARTIFACT/missing(id)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised

---

### TC-23: Unresolvable input reference

**Setup**

Input entries:
- logical_name = "ROOT", frontmatter = (empty)
- logical_name = "ROOT/a", frontmatter with input = "ARTIFACT/missing(id)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised
