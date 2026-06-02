<!-- code-from-spec: ROOT/functional/tests/chain/resolver@KUQajnDlwFOgpC9JVGq9whwbaEQ -->

# Test cases for ChainResolve

All tests create a spec tree on disk with `_node.md` files containing
frontmatter as needed, then call `ChainResolve` with a target logical name.

---

## Ancestors and target

### Root as target

Setup: create spec tree with ROOT only (_node.md with empty frontmatter).

Action: call ChainResolve with "ROOT".

Expect:
- ancestors = empty list
- target = ChainItem with logical_name = "ROOT", qualifier = absent
- dependencies = empty list
- external = empty list
- input = absent

---

### Linear chain — ancestors in root-first order

Setup: create spec tree with ROOT, ROOT/a, ROOT/a/b (all with empty frontmatter).

Action: call ChainResolve with "ROOT/a/b".

Expect:
- ancestors = [ChainItem(logical_name="ROOT"), ChainItem(logical_name="ROOT/a")]
  in that order (root first)
- target = ChainItem with logical_name = "ROOT/a/b"

---

### Single parent

Setup: create spec tree with ROOT, ROOT/a (empty frontmatter).

Action: call ChainResolve with "ROOT/a".

Expect:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem with logical_name = "ROOT/a"

---

### Target with empty frontmatter

Setup: create spec tree with ROOT, ROOT/a (leaf, empty frontmatter).

Action: call ChainResolve with "ROOT/a".

Expect:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem with logical_name = "ROOT/a"
- dependencies = empty list
- external = empty list
- input = absent

---

## Dependencies — ROOT/ references

### Dependency without qualifier

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b"]), ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains one ChainItem with logical_name = "ROOT/b",
  qualifier = absent

---

### Dependency with qualifier

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b(interface)"]), ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains one ChainItem with logical_name = "ROOT/b",
  qualifier = "interface"

---

### Dependencies sorted by file path then qualifier

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/z", "ROOT/m", "ROOT/b"]),
ROOT/z, ROOT/m, ROOT/b (all empty frontmatter).

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies sorted alphabetically by file_path value
  (ROOT/b before ROOT/m before ROOT/z)

---

## Dependencies — ARTIFACT/ references

### ARTIFACT dependency resolved from generating node

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ARTIFACT/b"]),
ROOT/b (frontmatter: output = "out/lib.go").

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains one ChainItem with logical_name = "ARTIFACT/b",
  file_path = PathCfs("out/lib.go")

---

### ARTIFACT — generating node has no output

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ARTIFACT/b"]),
ROOT/b (empty frontmatter, no output).

Action: call ChainResolve with "ROOT/a".

Expect: error UnresolvableArtifact.

---

### ARTIFACT — artifact file does not exist on disk

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ARTIFACT/b"]),
ROOT/b (output = "out/lib.go"). Do NOT create "out/lib.go" on disk.

Action: call ChainResolve with "ROOT/a".

Expect:
- no error
- dependencies contains one ChainItem with file_path = PathCfs("out/lib.go")
  (existence is not verified by the resolver)

---

### Mixed ROOT/ and ARTIFACT/ dependencies

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/c", "ARTIFACT/b"]),
ROOT/b (output = "out/lib.go"), ROOT/c.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains two entries (one for ROOT/c, one for ARTIFACT/b),
  sorted by file_path value

---

## Dependencies — dedup

### Exact duplicate — same file, same qualifier

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b", "ROOT/b"]), ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains exactly one entry for ROOT/b (not two)

---

### No qualifier subsumes qualifier

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b", "ROOT/b(interface)"]),
ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains exactly one entry for ROOT/b with qualifier = absent
  (the ROOT/b(interface) entry is removed because the unqualified entry subsumes it)

---

### Qualifier before no-qualifier — no-qualifier wins

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b(interface)", "ROOT/b"]),
ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains exactly one entry for ROOT/b with qualifier = absent

---

### Same file, different qualifiers — both kept

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ROOT/b(interface)", "ROOT/b(constraints)"]),
ROOT/b.

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains two entries: one with qualifier = "constraints" and one
  with qualifier = "interface" (sorted alphabetically)

---

### Duplicate ARTIFACT — same logical name

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["ARTIFACT/b", "ARTIFACT/b"]),
ROOT/b (output = "out/lib.go").

Action: call ChainResolve with "ROOT/a".

Expect:
- dependencies contains exactly one ARTIFACT/b entry (not two)

---

## External

### External entries copied from frontmatter

Setup: create spec tree with ROOT, ROOT/a (external = [{path: "docs/api.yaml"},
{path: "proto/v1.proto"}]).

Action: call ChainResolve with "ROOT/a".

Expect:
- external list contains both entries, sorted alphabetically:
  docs/api.yaml before proto/v1.proto

---

### Empty external — no entries

Setup: create spec tree with ROOT, ROOT/a (no external field).

Action: call ChainResolve with "ROOT/a".

Expect:
- external list is empty

---

## Input

### Input resolved from generating node

Setup: create spec tree with ROOT, ROOT/a (input = "ARTIFACT/b"),
ROOT/b (output = "out/data.json").

Action: call ChainResolve with "ROOT/a".

Expect:
- input = ChainItem with logical_name = "ARTIFACT/b",
  file_path = PathCfs("out/data.json")

---

### No input — absent

Setup: create spec tree with ROOT, ROOT/a (no input field).

Action: call ChainResolve with "ROOT/a".

Expect:
- input is absent

---

## Error cases

### Unrecognized prefix in depends_on

Setup: create spec tree with ROOT, ROOT/a (depends_on = ["UNKNOWN/something"]).

Action: call ChainResolve with "ROOT/a".

Expect: error UnresolvableArtifact.

---

### Invalid target logical name

Action: call ChainResolve with "INVALID/something".

Expect: error propagated from LogicalNameGetParent or LogicalNameToPath.

---

### Unreadable frontmatter

Setup: create spec tree with ROOT, ROOT/a with invalid YAML in frontmatter.

Action: call ChainResolve with "ROOT/a".

Expect: error UnreadableFrontmatter.
