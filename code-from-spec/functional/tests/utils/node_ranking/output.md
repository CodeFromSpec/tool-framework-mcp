<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@SN8E_GyL7bWuk_Zh0ny5IgpoG6E -->

## Test suite: NodeRankCompute

---

### Happy path

#### Test: Root only

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked contains one NodeRankEntry: logical_name="ROOT", rank=0.
  - cycles is empty.

---

#### Test: Linear chain — incrementing ranks

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",     frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a/b", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - NodeRankEntry for "ROOT"     has rank 0.
  - NodeRankEntry for "ROOT/a"   has rank 1.
  - NodeRankEntry for "ROOT/a/b" has rank 2.
  - cycles is empty.

---

#### Test: Independent siblings — equal rank

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - NodeRankEntry for "ROOT/a" has rank 1.
  - NodeRankEntry for "ROOT/b" has rank 1.
  - cycles is empty.

---

#### Test: depends_on increases rank

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/b" > rank of "ROOT/a".
  - cycles is empty.

---

#### Test: depends_on with qualifier — qualifier stripped

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a(interface)"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - The qualified reference "ROOT/a(interface)" resolves to bare node "ROOT/a".
  - rank of "ROOT/b" > rank of "ROOT/a".
  - cycles is empty.

---

#### Test: input artifact adds dependency edge

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "out.go" } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { input: "ARTIFACT/a" } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked includes an artifact entry "ARTIFACT/a".
  - rank of "ROOT/b" > rank of "ARTIFACT/a".
  - rank of "ARTIFACT/a" > rank of "ROOT/a".
  - cycles is empty.

---

#### Test: Artifacts get rank one above their node

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "foo.go" } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked includes artifact entry "ARTIFACT/a".
  - rank of "ARTIFACT/a" = rank of "ROOT/a" + 1.
  - cycles is empty.

---

#### Test: Single output — artifact ranked

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "x.go" } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked includes exactly one artifact entry "ARTIFACT/a".
  - rank of "ARTIFACT/a" = rank of "ROOT/a" + 1.
  - cycles is empty.

---

#### Test: depends_on ARTIFACT reference — used as-is

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "lib.go" } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ARTIFACT/a"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/b" > rank of "ARTIFACT/a".
  - rank of "ARTIFACT/a" > rank of "ROOT/a".
  - cycles is empty.

---

#### Test: Output sorted by rank then logical name

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/z", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked is ordered: "ROOT" first (rank 0), then "ROOT/a" before "ROOT/z" (both rank 1, alphabetical).
  - cycles is empty.

---

#### Test: Parallel entries — equal rank means no dependency

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/c", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - "ROOT/a", "ROOT/b", and "ROOT/c" all have rank 1.
  - cycles is empty.

---

#### Test: Diamond dependency — rank uses max not sum

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/c", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/c"] } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/c"] } },
    NodeRankInput { logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a", "ROOT/b"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/c" = 1.
  - rank of "ROOT/a" = 2.
  - rank of "ROOT/b" = 2.
  - rank of "ROOT/d" = 3 (1 + max(2, 2) = 3, not the sum).
  - cycles is empty.

---

#### Test: depends_on outranks parent

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",       frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a",     frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a/b",   frontmatter: { depends_on: ["ROOT/c"] } },
    NodeRankInput { logical_name: "ROOT/c",     frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/c/d",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/c/d/e", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/a/b" > rank of "ROOT/a".
  - rank of "ROOT/a/b" = 1 + max(rank of "ROOT/a", rank of "ROOT/c").
  - cycles is empty.

---

#### Test: Multiple depends_on — rank from highest

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } },
    NodeRankInput { logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"] } },
    NodeRankInput { logical_name: "ROOT/d", frontmatter: { depends_on: ["ROOT/a", "ROOT/b", "ROOT/c"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/a" = 1.
  - rank of "ROOT/b" = 2.
  - rank of "ROOT/c" = 3.
  - rank of "ROOT/d" = 4 (1 + max(1, 2, 3) = 4).
  - cycles is empty.

---

#### Test: Node with both depends_on and input

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "a.go" } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/c", frontmatter: { depends_on: ["ROOT/b"], input: "ARTIFACT/a" } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - rank of "ROOT/c" = 1 + max(rank of "ROOT" (parent of ROOT/c), rank of "ROOT/b", rank of "ARTIFACT/a").
  - cycles is empty.

---

#### Test: Empty input list

Setup:
  entries = []

Action: call NodeRankCompute(entries)

Expected outcome:
  - No error.
  - ranked is empty.
  - cycles is empty.

---

### Cycle detection

#### Test: Self-reference

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/a"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - cycles is not empty.

---

#### Test: Simple cycle — two nodes

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"] } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - cycles is not empty.
  - cycles contains at least one of "ROOT/a" or "ROOT/b".

---

#### Test: Cycle through artifacts

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { output: "a.go", depends_on: ["ARTIFACT/b"] } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { output: "b.go", depends_on: ["ARTIFACT/a"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - cycles is not empty.

---

#### Test: Cycle does not prevent ranking of unrelated nodes

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/b"] } },
    NodeRankInput { logical_name: "ROOT/b", frontmatter: { depends_on: ["ROOT/a"] } },
    NodeRankInput { logical_name: "ROOT/c", frontmatter: empty }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - NodeRankEntry for "ROOT"   has rank 0.
  - NodeRankEntry for "ROOT/c" has rank 1.
  - cycles is not empty and contains entries related to "ROOT/a" and/or "ROOT/b".
  - "ROOT/c" is not in cycles.

---

### Error cases

#### Test: Unresolvable ROOT reference

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ROOT/missing"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - Error UnresolvableReference is raised.

---

#### Test: Unresolvable ARTIFACT reference

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { depends_on: ["ARTIFACT/missing"] } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - Error UnresolvableReference is raised.

---

#### Test: Unresolvable input reference

Setup:
  entries = [
    NodeRankInput { logical_name: "ROOT",   frontmatter: empty },
    NodeRankInput { logical_name: "ROOT/a", frontmatter: { input: "ARTIFACT/missing" } }
  ]

Action: call NodeRankCompute(entries)

Expected outcome:
  - Error UnresolvableReference is raised.
