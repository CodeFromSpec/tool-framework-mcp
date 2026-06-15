<!-- code-from-spec: SPEC/functional/tests/utils/node_ranking@kiFxjMjT5Y5ajfjFHwp13NStxRs -->

## Test Suite: NodeRankCompute

---

### TC-01: Root only

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  ranked = [ NodeRankEntry { logical_name: "SPEC", rank: 0 } ]
  cycles = []

---

### TC-02: Linear chain — incrementing ranks

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",      frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a",    frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a/b",  frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC"     has rank 0
  NodeRankEntry "SPEC/a"   has rank 1
  NodeRankEntry "SPEC/a/b" has rank 2
  cycles = []

---

### TC-03: Independent siblings — equal rank

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC/a" and NodeRankEntry "SPEC/b" have the same rank (1)
  cycles = []

---

### TC-04: depends_on increases rank

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  rank of "SPEC/b" > rank of "SPEC/a"
  cycles = []

---

### TC-05: depends_on with qualifier — qualifier stripped

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a(interface)"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  No error raised
  rank of "SPEC/b" > rank of "SPEC/a"
  cycles = []

---

### TC-06: EXTERNAL depends_on — skipped for ranking

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["EXTERNAL/proto/api.proto"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  No error raised
  NodeRankEntry "SPEC/a" has rank 1
  cycles = []

---

### TC-07: input artifact adds dependency edge

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "out.go" } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { input: "ARTIFACT/a" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  rank of "SPEC/b" > rank of "ARTIFACT/a"
  rank of "ARTIFACT/a" > rank of "SPEC/a"
  cycles = []

---

### TC-08: EXTERNAL input — skipped for ranking

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { input: "EXTERNAL/docs/spec.yaml" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  No error raised
  NodeRankEntry "SPEC/a" has rank 1
  cycles = []

---

### TC-09: Artifacts get rank one above their node

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "foo.go" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "ARTIFACT/a" has rank = rank of "SPEC/a" + 1
  cycles = []

---

### TC-10: Single output — artifact ranked

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "x.go" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  ranked contains NodeRankEntry "ARTIFACT/a" with rank = rank of "SPEC/a" + 1
  cycles = []

---

### TC-11: depends_on ARTIFACT reference — used as-is

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "lib.go" } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["ARTIFACT/a"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  rank of "SPEC/b" > rank of "ARTIFACT/a"
  rank of "ARTIFACT/a" > rank of "SPEC/a"
  cycles = []

---

### TC-12: Output sorted by rank then logical name

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/z", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  ranked[0] = NodeRankEntry { logical_name: "SPEC",   rank: 0 }
  ranked[1] = NodeRankEntry { logical_name: "SPEC/a", rank: 1 }
  ranked[2] = NodeRankEntry { logical_name: "SPEC/z", rank: 1 }
  cycles = []

---

### TC-13: Parallel entries — equal rank means no dependency

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC/a", "SPEC/b", "SPEC/c" all have rank 1
  cycles = []

---

### TC-14: Diamond dependency — rank uses max not sum

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/c", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/c"] } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/c"] } },
    NodeRankInput { logical_name: "SPEC/d", frontmatter: { depends_on: ["SPEC/a", "SPEC/b"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC/c" has rank 1
  NodeRankEntry "SPEC/a" has rank 2
  NodeRankEntry "SPEC/b" has rank 2
  NodeRankEntry "SPEC/d" has rank 3
  cycles = []

---

### TC-15: depends_on outranks parent

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",       frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a",     frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a/b",   frontmatter: { depends_on: ["SPEC/c"] } },
    NodeRankInput { logical_name: "SPEC/c",     frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/c/d",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/c/d/e", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  rank of "SPEC/a/b" > rank of "SPEC/a"
  rank of "SPEC/a/b" = 1 + max(rank of "SPEC/a", rank of "SPEC/c")
  cycles = []

---

### TC-16: Multiple depends_on — rank from highest

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } },
    NodeRankInput { logical_name: "SPEC/c", frontmatter: { depends_on: ["SPEC/b"] } },
    NodeRankInput { logical_name: "SPEC/d", frontmatter: { depends_on: ["SPEC/a", "SPEC/b", "SPEC/c"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC/a" has rank 1
  NodeRankEntry "SPEC/b" has rank 2
  NodeRankEntry "SPEC/c" has rank 3
  NodeRankEntry "SPEC/d" has rank 4
  cycles = []

---

### TC-17: Node with both depends_on and input

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "a.go" } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/c", frontmatter: { depends_on: ["SPEC/b"], input: "ARTIFACT/a" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  rank of "SPEC/c" = 1 + max(rank of "SPEC" (parent), rank of "SPEC/b", rank of "ARTIFACT/a")
  cycles = []

---

### TC-18: Empty input list

Setup:
  entries = []

Action:
  Call NodeRankCompute(entries)

Expected:
  ranked = []
  cycles = []

---

### TC-19: Self-reference

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/a"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  cycles is not empty

---

### TC-20: Simple cycle — two nodes

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"] } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  cycles is not empty
  cycles contains at least one of "SPEC/a" or "SPEC/b"

---

### TC-21: Cycle through artifacts

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "a.go", depends_on: ["ARTIFACT/b"] } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { output: "b.go", depends_on: ["ARTIFACT/a"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  cycles is not empty

---

### TC-22: Cycle does not prevent ranking of unrelated nodes

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"] } },
    NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } },
    NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  NodeRankEntry "SPEC"   has rank 0
  NodeRankEntry "SPEC/c" has rank 1
  cycles is not empty
  cycles contains entries related to "SPEC/a" and/or "SPEC/b"
  cycles does not contain "SPEC/c"

---

### TC-23: Unresolvable SPEC reference

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/missing"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  Raises error UnresolvableReference

---

### TC-24: Unresolvable ARTIFACT reference

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["ARTIFACT/missing"] } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  Raises error UnresolvableReference

---

### TC-25: Unresolvable input reference

Setup:
  entries = [
    NodeRankInput { logical_name: "SPEC",   frontmatter: empty },
    NodeRankInput { logical_name: "SPEC/a", frontmatter: { input: "ARTIFACT/missing" } }
  ]

Action:
  Call NodeRankCompute(entries)

Expected:
  Raises error UnresolvableReference
