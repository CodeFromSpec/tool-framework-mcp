<!-- code-from-spec: ROOT/functional/tests/utils/node_ranking@Dp9AEnA0DE3fcKtTMoNI04LXCTo -->

# Test cases for NodeRankCompute

Input is always a list of `NodeRankInput`. Each entry has: logical_name and
frontmatter (Frontmatter record). No file I/O is performed.

---

## Happy path

### Root only

Setup: input = [NodeRankInput(logical_name="ROOT", frontmatter=empty)].

Action: call NodeRankCompute.

Expect:
- ranked list contains one entry: NodeRankEntry(logical_name="ROOT", rank=0)
- cycles = empty list

---

### Linear chain — incrementing ranks

Setup: input = [
  NodeRankInput(logical_name="ROOT", frontmatter=empty),
  NodeRankInput(logical_name="ROOT/a", frontmatter=empty),
  NodeRankInput(logical_name="ROOT/a/b", frontmatter=empty)
]

Action: call NodeRankCompute.

Expect:
- NodeRankEntry("ROOT", rank=0)
- NodeRankEntry("ROOT/a", rank=1)
- NodeRankEntry("ROOT/a/b", rank=2)
- cycles = empty list

---

### Independent siblings — equal rank

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/b", empty)
]

Action: call NodeRankCompute.

Expect:
- ROOT/a and ROOT/b have the same rank (1)
- cycles = empty list

---

### depends_on increases rank

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/a"])
]

Action: call NodeRankCompute.

Expect:
- rank of ROOT/b > rank of ROOT/a
- cycles = empty list

---

### depends_on with qualifier — qualifier stripped

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/a(interface)"])
]

Action: call NodeRankCompute.

Expect:
- no error
- rank of ROOT/b > rank of ROOT/a
  (qualified reference resolves to bare node ROOT/a)
- cycles = empty list

---

### input artifact adds dependency edge

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "out.go"),
  NodeRankInput("ROOT/b", frontmatter with input = "ARTIFACT/a")
]

Action: call NodeRankCompute.

Expect:
- rank of ROOT/b > rank of ARTIFACT/a > rank of ROOT/a
- cycles = empty list

---

### Artifacts get rank one above their node

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "foo.go")
]

Action: call NodeRankCompute.

Expect:
- ranked list contains an entry for ARTIFACT/a
- rank of ARTIFACT/a = rank of ROOT/a + 1
- cycles = empty list

---

### Single output — artifact ranked

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "x.go")
]

Action: call NodeRankCompute.

Expect:
- ranked list contains NodeRankEntry("ARTIFACT/a", rank = rank of ROOT/a + 1)
- cycles = empty list

---

### depends_on ARTIFACT reference — used as-is

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "lib.go"),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ARTIFACT/a"])
]

Action: call NodeRankCompute.

Expect:
- rank of ROOT/b > rank of ARTIFACT/a > rank of ROOT/a
- cycles = empty list

---

### Output sorted by rank then logical name

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/z", empty),
  NodeRankInput("ROOT/a", empty)
]

Action: call NodeRankCompute.

Expect:
- output order: ROOT (rank 0), then ROOT/a before ROOT/z
  (both rank 1, sorted alphabetically)
- cycles = empty list

---

### Parallel entries — equal rank means no dependency

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/b", empty),
  NodeRankInput("ROOT/c", empty)
]

Action: call NodeRankCompute.

Expect:
- ROOT/a, ROOT/b, ROOT/c all have rank 1
- cycles = empty list

---

### Diamond dependency — rank uses max not sum

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/c", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ROOT/c"]),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/c"]),
  NodeRankInput("ROOT/d", frontmatter with depends_on = ["ROOT/a", "ROOT/b"])
]

Action: call NodeRankCompute.

Expect:
- ROOT/c rank = 1
- ROOT/a rank = 2
- ROOT/b rank = 2
- ROOT/d rank = 3 (not 5 — rank is 1 + max of dependencies, not sum)
- cycles = empty list

---

### depends_on outranks parent

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/a/b", frontmatter with depends_on = ["ROOT/c"]),
  NodeRankInput("ROOT/c", empty),
  NodeRankInput("ROOT/c/d", empty),
  NodeRankInput("ROOT/c/d/e", empty)
]

Action: call NodeRankCompute.

Expect:
- rank of ROOT/a/b = 1 + max(rank of ROOT/a, rank of ROOT/c)
- rank of ROOT/a/b > rank of ROOT/a
- cycles = empty list

---

### Multiple depends_on — rank from highest

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", empty),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/a"]),
  NodeRankInput("ROOT/c", frontmatter with depends_on = ["ROOT/b"]),
  NodeRankInput("ROOT/d", frontmatter with depends_on = ["ROOT/a", "ROOT/b", "ROOT/c"])
]

Action: call NodeRankCompute.

Expect:
- ROOT/a rank = 1
- ROOT/b rank = 2
- ROOT/c rank = 3
- ROOT/d rank = 4 (1 + max(1, 2, 3) = 4)
- cycles = empty list

---

### Node with both depends_on and input

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "a.go"),
  NodeRankInput("ROOT/b", empty),
  NodeRankInput("ROOT/c", frontmatter with depends_on = ["ROOT/b"] and
    input = "ARTIFACT/a")
]

Action: call NodeRankCompute.

Expect:
- rank of ROOT/c = 1 + max(rank of ROOT (parent), rank of ROOT/b, rank of ARTIFACT/a)
- cycles = empty list

---

### Empty input list

Setup: input = empty list.

Action: call NodeRankCompute.

Expect:
- ranked list = empty
- cycles = empty list

---

## Cycle detection

### Self-reference

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ROOT/a"])
]

Action: call NodeRankCompute.

Expect:
- cycles list is not empty

---

### Simple cycle — two nodes

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ROOT/b"]),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/a"])
]

Action: call NodeRankCompute.

Expect:
- cycles list is not empty and contains at least one of "ROOT/a" or "ROOT/b"

---

### Cycle through artifacts

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with output = "a.go" and
    depends_on = ["ARTIFACT/b"]),
  NodeRankInput("ROOT/b", frontmatter with output = "b.go" and
    depends_on = ["ARTIFACT/a"])
]

Action: call NodeRankCompute.

Expect:
- cycles list is not empty

---

### Cycle does not prevent ranking of unrelated nodes

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ROOT/b"]),
  NodeRankInput("ROOT/b", frontmatter with depends_on = ["ROOT/a"]),
  NodeRankInput("ROOT/c", empty)
]

Action: call NodeRankCompute.

Expect:
- ROOT has a valid rank (0)
- ROOT/c has a valid rank (1)
- cycles list contains entries related to ROOT/a and/or ROOT/b but not ROOT/c

---

## Error cases

### Unresolvable ROOT reference

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ROOT/missing"])
]

Action: call NodeRankCompute.

Expect: error UnresolvableReference.

---

### Unresolvable ARTIFACT reference

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with depends_on = ["ARTIFACT/missing"])
]

Action: call NodeRankCompute.

Expect: error UnresolvableReference.

---

### Unresolvable input reference

Setup: input = [
  NodeRankInput("ROOT", empty),
  NodeRankInput("ROOT/a", frontmatter with input = "ARTIFACT/missing")
]

Action: call NodeRankCompute.

Expect: error UnresolvableReference.
