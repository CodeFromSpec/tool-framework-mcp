<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@AJarAZGdIIkkxThBLGxZS2GGBeg -->

# Node Ranking Test Specification

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
    - unresolvable reference: a depends_on or input target cannot be resolved.
```

---

## Happy Path

### TC-01: Root only

**Setup**

Input: one NodeRankInput
- logical_name = "ROOT", frontmatter has no depends_on, no input, no outputs

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked list contains exactly one NodeRankEntry: logical_name = "ROOT", rank = 0
- cycles list is empty

---

### TC-02: Linear chain — incrementing ranks

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/a/b", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- NodeRankEntry for "ROOT" has rank = 0
- NodeRankEntry for "ROOT/a" has rank = 1
- NodeRankEntry for "ROOT/a/b" has rank = 2
- cycles list is empty

---

### TC-03: Independent siblings — equal rank

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/b", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- NodeRankEntry for "ROOT/a" and NodeRankEntry for "ROOT/b" have the same rank (1)
- cycles list is empty

---

### TC-04: depends_on increases rank

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- NodeRankEntry for "ROOT/b" has rank strictly greater than rank of "ROOT/a"
- cycles list is empty

---

### TC-05: depends_on with qualifier — qualifier stripped

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a(interface)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- No error is raised
- NodeRankEntry for "ROOT/b" has rank strictly greater than rank of "ROOT/a"
  (the qualified reference "ROOT/a(interface)" resolves to the bare node "ROOT/a")
- cycles list is empty

---

### TC-06: input artifact adds dependency edge

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "code", path: "out.go"}]
- logical_name = "ROOT/b", frontmatter has input = "ARTIFACT/a(code)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked list contains a NodeRankEntry for artifact "ARTIFACT/a(code)"
- rank of "ROOT/b" > rank of "ARTIFACT/a(code)" > rank of "ROOT/a"
- cycles list is empty

---

### TC-07: Artifacts get rank one above their node

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "foo", path: "foo.go"}]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked list contains a NodeRankEntry for artifact "ARTIFACT/a(foo)"
- rank of "ARTIFACT/a(foo)" = rank of "ROOT/a" + 1
- cycles list is empty

---

### TC-08: Multiple outputs — each artifact ranked

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "x.go"}, {id: "y", path: "y.go"}]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- ranked list contains NodeRankEntry for "ARTIFACT/a(x)" and NodeRankEntry for "ARTIFACT/a(y)"
- rank of "ARTIFACT/a(x)" = rank of "ROOT/a" + 1
- rank of "ARTIFACT/a(y)" = rank of "ROOT/a" + 1
- cycles list is empty

---

### TC-09: depends_on ARTIFACT reference — used as-is

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "lib", path: "lib.go"}]
- logical_name = "ROOT/b", frontmatter has depends_on = ["ARTIFACT/a(lib)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/b" > rank of "ARTIFACT/a(lib)" > rank of "ROOT/a"
- cycles list is empty

---

### TC-10: Output sorted by rank then logical name

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/z", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- "ROOT" appears first in ranked list (rank 0)
- "ROOT/a" appears before "ROOT/z" in ranked list (both rank 1, alphabetical order)
- cycles list is empty

---

### TC-11: Parallel entries — equal rank means no dependency

**Setup**

Input: four NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/b", empty frontmatter
- logical_name = "ROOT/c", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- NodeRankEntry for "ROOT/a", "ROOT/b", and "ROOT/c" all have rank = 1
- cycles list is empty

---

### TC-12: Diamond dependency — rank uses max not sum

**Setup**

Input: five NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/c", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/c"]
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/c"]
- logical_name = "ROOT/d", frontmatter has depends_on = ["ROOT/a", "ROOT/b"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/c" = 1
- rank of "ROOT/a" = 2
- rank of "ROOT/b" = 2
- rank of "ROOT/d" = 3  (1 + max(rank of ROOT/a, rank of ROOT/b) = 1 + 2 = 3, not the sum)
- cycles list is empty

---

### TC-13: depends_on outranks parent

**Setup**

Input: six NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/a/b", frontmatter has depends_on = ["ROOT/c"]
- logical_name = "ROOT/c", empty frontmatter
- logical_name = "ROOT/c/d", empty frontmatter
- logical_name = "ROOT/c/d/e", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/a/b" > rank of "ROOT/a"
  (rank of "ROOT/a/b" = 1 + max(rank of "ROOT/a", rank of "ROOT/c"))
- cycles list is empty

---

### TC-14: Multiple depends_on — rank from highest

**Setup**

Input: five NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", empty frontmatter
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a"]
- logical_name = "ROOT/c", frontmatter has depends_on = ["ROOT/b"]
- logical_name = "ROOT/d", frontmatter has depends_on = ["ROOT/a", "ROOT/b", "ROOT/c"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/a" = 1
- rank of "ROOT/b" = 2
- rank of "ROOT/c" = 3
- rank of "ROOT/d" = 4  (1 + max(1, 2, 3) = 4)
- cycles list is empty

---

### TC-15: Node with both depends_on and input

**Setup**

Input: four NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "out", path: "a.go"}]
- logical_name = "ROOT/b", empty frontmatter
- logical_name = "ROOT/c", frontmatter has depends_on = ["ROOT/b"] and input = "ARTIFACT/a(out)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- rank of "ROOT/c" = 1 + max(rank of "ROOT" (parent of ROOT/c), rank of "ROOT/b", rank of "ARTIFACT/a(out)")
- cycles list is empty

---

### TC-16: Empty input list

**Setup**

Input: empty list

**Action**

Call NodeRankCompute with the empty list.

**Expected outcome**

- ranked list is empty
- cycles list is empty

---

## Cycle Detection

### TC-17: Self-reference

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles list is not empty

---

### TC-18: Simple cycle — two nodes

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"]
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles list is not empty
- cycles list contains at least one of "ROOT/a" or "ROOT/b"

---

### TC-19: Cycle through artifacts

**Setup**

Input: three NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has outputs = [{id: "out", path: "a.go"}] and depends_on = ["ARTIFACT/b(out)"]
- logical_name = "ROOT/b", frontmatter has outputs = [{id: "out", path: "b.go"}] and depends_on = ["ARTIFACT/a(out)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- cycles list is not empty

---

### TC-20: Cycle does not prevent ranking of unrelated nodes

**Setup**

Input: four NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"]
- logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a"]
- logical_name = "ROOT/c", empty frontmatter

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- NodeRankEntry for "ROOT" has a valid rank (0)
- NodeRankEntry for "ROOT/c" has a valid rank (1)
- cycles list contains entries related to "ROOT/a" and/or "ROOT/b"
- cycles list does not contain "ROOT/c"

---

## Error Cases

### TC-21: Unresolvable ROOT reference

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/missing"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised

---

### TC-22: Unresolvable ARTIFACT reference

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has depends_on = ["ARTIFACT/missing(id)"]

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised

---

### TC-23: Unresolvable input reference

**Setup**

Input: two NodeRankInput entries
- logical_name = "ROOT", empty frontmatter
- logical_name = "ROOT/a", frontmatter has input = "ARTIFACT/missing(id)"

**Action**

Call NodeRankCompute with the input list.

**Expected outcome**

- Error "unresolvable reference" is raised
