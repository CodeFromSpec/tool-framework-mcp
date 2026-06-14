<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@89gKElh_1SkNvJjdiA3ZhoKiI7o -->

## Test cases for NodeRankCompute

Each test builds a list of `NodeRankInput` records, calls `NodeRankCompute`, and
checks the returned `ranked` list and `cycles` list.

---

### Happy path

---

#### TC-01 Root only

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- ranked contains exactly one entry: NodeRankEntry { logical_name: "SPEC", rank: 0 }
- cycles is empty

---

#### TC-02 Linear chain — incrementing ranks

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a/b", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- NodeRankEntry { logical_name: "SPEC", rank: 0 }
- NodeRankEntry { logical_name: "SPEC/a", rank: 1 }
- NodeRankEntry { logical_name: "SPEC/a/b", rank: 2 }
- cycles is empty

---

#### TC-03 Independent siblings — equal rank

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- SPEC/a and SPEC/b have the same rank (1)
- cycles is empty

---

#### TC-04 depends_on increases rank

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }

Action: call NodeRankCompute(entries)

Expected:
- rank of SPEC/b > rank of SPEC/a
- cycles is empty

---

#### TC-05 depends_on with qualifier — qualifier stripped

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a(interface)"] } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- rank of SPEC/b > rank of SPEC/a
- the qualified reference "SPEC/a(interface)" resolves to bare node "SPEC/a"
- cycles is empty

---

#### TC-06 EXTERNAL depends_on — skipped for ranking

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["EXTERNAL/proto/api.proto"] } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- NodeRankEntry { logical_name: "SPEC/a", rank: 1 } (rank derived from parent only)
- "EXTERNAL/proto/api.proto" does not contribute to rank
- cycles is empty

---

#### TC-07 input artifact adds dependency edge

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "out.go" } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { input: "ARTIFACT/a" } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- rank of SPEC/b > rank of ARTIFACT/a
- rank of ARTIFACT/a > rank of SPEC/a
- cycles is empty

---

#### TC-08 EXTERNAL input — skipped for ranking

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { input: "EXTERNAL/docs/spec.yaml" } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- NodeRankEntry { logical_name: "SPEC/a", rank: 1 } (rank derived from parent only)
- "EXTERNAL/docs/spec.yaml" does not contribute to rank
- cycles is empty

---

#### TC-09 Artifacts get rank one above their node

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "foo.go" } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- ranked contains an entry NodeRankEntry { logical_name: "ARTIFACT/a" }
- rank of ARTIFACT/a = rank of SPEC/a + 1
- cycles is empty

---

#### TC-10 Single output — artifact ranked

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "x.go" } }

Action: call NodeRankCompute(entries)

Expected:
- ranked contains exactly one artifact entry: NodeRankEntry { logical_name: "ARTIFACT/a" }
- rank of ARTIFACT/a = rank of SPEC/a + 1
- cycles is empty

---

#### TC-11 depends_on ARTIFACT reference — used as-is

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "lib.go" } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["ARTIFACT/a"] } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- rank of SPEC/b > rank of ARTIFACT/a
- rank of ARTIFACT/a > rank of SPEC/a
- cycles is empty

---

#### TC-12 Output sorted by rank then logical name

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/z", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- first entry in ranked is NodeRankEntry { logical_name: "SPEC", rank: 0 }
- SPEC/a appears before SPEC/z in ranked (both rank 1, alphabetical order)
- cycles is empty

---

#### TC-13 Parallel entries — equal rank means no dependency

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- SPEC/a, SPEC/b, and SPEC/c all have rank 1
- cycles is empty

---

#### TC-14 Diamond dependency — rank uses max not sum

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/c"] } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/c"] } }
  - NodeRankInput { logical_name: "SPEC/d", frontmatter: { depends_on: ["SPEC/a", "SPEC/b"] } }

Action: call NodeRankCompute(entries)

Expected:
- NodeRankEntry { logical_name: "SPEC/c", rank: 1 }
- NodeRankEntry { logical_name: "SPEC/a", rank: 2 }
- NodeRankEntry { logical_name: "SPEC/b", rank: 2 }
- NodeRankEntry { logical_name: "SPEC/d", rank: 3 } (1 + max(2, 2) = 3, not 5)
- cycles is empty

---

#### TC-15 depends_on outranks parent

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c/d", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c/d/e", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a/b", frontmatter: { depends_on: ["SPEC/c"] } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- rank of SPEC/a/b > rank of SPEC/a
- rank of SPEC/a/b = 1 + max(rank of SPEC/a, rank of SPEC/c)
- cycles is empty

---

#### TC-16 Multiple depends_on — rank from highest

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: { depends_on: ["SPEC/b"] } }
  - NodeRankInput { logical_name: "SPEC/d", frontmatter: { depends_on: ["SPEC/a", "SPEC/b", "SPEC/c"] } }

Action: call NodeRankCompute(entries)

Expected:
- NodeRankEntry { logical_name: "SPEC/a", rank: 1 }
- NodeRankEntry { logical_name: "SPEC/b", rank: 2 }
- NodeRankEntry { logical_name: "SPEC/c", rank: 3 }
- NodeRankEntry { logical_name: "SPEC/d", rank: 4 } (1 + max(1,2,3) = 4)
- cycles is empty

---

#### TC-17 Node with both depends_on and input

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "a.go" } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: { depends_on: ["SPEC/b"], input: "ARTIFACT/a" } }

Action: call NodeRankCompute(entries)

Expected:
- no error
- rank of SPEC/c = 1 + max(rank of SPEC (parent), rank of SPEC/b, rank of ARTIFACT/a)
- cycles is empty

---

#### TC-18 Empty input list

Setup:
- entries: empty list

Action: call NodeRankCompute(entries)

Expected:
- ranked is an empty list
- cycles is empty

---

### Cycle detection

---

#### TC-19 Self-reference

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/a"] } }

Action: call NodeRankCompute(entries)

Expected:
- cycles is not empty

---

#### TC-20 Simple cycle — two nodes

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"] } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }

Action: call NodeRankCompute(entries)

Expected:
- cycles is not empty
- cycles contains at least one of "SPEC/a" or "SPEC/b"

---

#### TC-21 Cycle through artifacts

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { output: "a.go", depends_on: ["ARTIFACT/b"] } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { output: "b.go", depends_on: ["ARTIFACT/a"] } }

Action: call NodeRankCompute(entries)

Expected:
- cycles is not empty

---

#### TC-22 Cycle does not prevent ranking of unrelated nodes

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"] } }
  - NodeRankInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a"] } }
  - NodeRankInput { logical_name: "SPEC/c", frontmatter: empty }

Action: call NodeRankCompute(entries)

Expected:
- SPEC has rank 0
- SPEC/c has rank 1
- cycles is not empty
- cycles contains entries related to SPEC/a and/or SPEC/b
- cycles does not contain "SPEC/c"

---

### Error cases

---

#### TC-23 Unresolvable SPEC reference

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/missing"] } }

Action: call NodeRankCompute(entries)

Expected:
- error UnresolvableReference is raised

---

#### TC-24 Unresolvable ARTIFACT reference

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["ARTIFACT/missing"] } }

Action: call NodeRankCompute(entries)

Expected:
- error UnresolvableReference is raised

---

#### TC-25 Unresolvable input reference

Setup:
- entries:
  - NodeRankInput { logical_name: "SPEC", frontmatter: empty }
  - NodeRankInput { logical_name: "SPEC/a", frontmatter: { input: "ARTIFACT/missing" } }

Action: call NodeRankCompute(entries)

Expected:
- error UnresolvableReference is raised
