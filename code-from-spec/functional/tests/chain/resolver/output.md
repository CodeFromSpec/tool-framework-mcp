<!-- code-from-spec: ROOT/functional/tests/chain/resolver@q3ZKratAUJoNpR_g5Xot2IOLgY4 -->

# Test Specification: ChainResolve

Each test case describes the spec tree to create on disk, the
action to perform, and the expected outcome.

---

## Ancestors and Target

### TC-01: Root as target

**Setup**

Create one file:
- `<root>/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT")`.

**Expected outcome**

Return a Chain where:
- `ancestors` = empty list
- `dependencies` = empty list
- `external` = empty list
- `target` = ChainItem with logical_name = "ROOT", qualifier = absent
- `input` = absent

---

### TC-02: Linear chain — ancestors in root-first order

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter
- `<root>/a/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a/b")`.

**Expected outcome**

Return a Chain where:
- `ancestors` = [ChainItem(logical_name="ROOT"), ChainItem(logical_name="ROOT/a")] in that order
- `target` = ChainItem with logical_name = "ROOT/a/b"

---

### TC-03: Single parent

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Return a Chain where:
- `ancestors` = [ChainItem(logical_name="ROOT")]
- `target` = ChainItem with logical_name = "ROOT/a"

---

### TC-04: Target with empty frontmatter

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter (no depends_on, no external, no input)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Return a Chain where:
- `ancestors` = [ChainItem(logical_name="ROOT")]
- `target` = ChainItem with logical_name = "ROOT/a"
- `dependencies` = empty list
- `external` = empty list
- `input` = absent

---

## Dependencies — ROOT/ References

### TC-05: Dependency without qualifier

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains one ChainItem with:
- `logical_name` = "ROOT/b"
- `qualifier` = absent

---

### TC-06: Dependency with qualifier

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains one ChainItem with:
- `logical_name` = "ROOT/b"
- `qualifier` = "interface"

---

### TC-07: Dependencies sorted by file path then qualifier

**Setup**

Create five files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/z", "ROOT/m", "ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter
- `<root>/m/_node.md` — empty frontmatter
- `<root>/z/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` is a list of three ChainItems sorted alphabetically by
file path. The entry whose file path resolves earliest in alphabetical
order appears first, ROOT/b before ROOT/m before ROOT/z.

---

## Dependencies — ARTIFACT/ References

### TC-08: ARTIFACT dependency resolved from generating node

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains one ChainItem with:
- `logical_name` = "ARTIFACT/b(lib)"
- `file_path` = "out/lib.go"
- `qualifier` = "lib"

---

### TC-09: ARTIFACT without qualifier — error

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b"]`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

### TC-10: ARTIFACT — generating node has no outputs

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — empty frontmatter (no outputs declared)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

### TC-11: ARTIFACT — artifact file does not exist on disk

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

Do NOT create `out/lib.go` on disk.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

No error. `dependencies` contains one ChainItem with:
- `file_path` = "out/lib.go"

File existence is not verified by the resolver.

---

### TC-12: ARTIFACT with non-existent output id — error

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(missing)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

### TC-13: Mixed ROOT/ and ARTIFACT/ dependencies

**Setup**

Create four files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/c", "ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- `<root>/c/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains two ChainItems, one for ROOT/c and one for
ARTIFACT/b(lib), sorted by file path value alphabetically.

---

## Dependencies — Deduplication

### TC-14: Exact duplicate — same file, same qualifier

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b", "ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains exactly one entry for ROOT/b (not two).

---

### TC-15: No qualifier subsumes qualifier

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b", "ROOT/b(interface)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains exactly one entry for ROOT/b with
`qualifier` = absent. The ROOT/b(interface) entry is removed.

---

### TC-16: Qualifier before no-qualifier — no-qualifier wins

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains exactly one entry for ROOT/b with
`qualifier` = absent. Order of appearance in depends_on does not matter;
the no-qualifier entry always wins.

---

### TC-17: Same file, different qualifiers — both kept

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b(constraints)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains exactly two entries:
- ChainItem with qualifier = "constraints"
- ChainItem with qualifier = "interface"

Both are kept and sorted by qualifier.

---

### TC-18: Duplicate ARTIFACT — same logical name

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)", "ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`dependencies` contains exactly one ARTIFACT entry for ARTIFACT/b(lib)
(not two).

---

## External

### TC-19: External entries copied from frontmatter

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter:
  ```
  external:
    - path: docs/api.yaml
    - path: proto/v1.proto
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`external` contains two entries sorted alphabetically:
- entry with path = "docs/api.yaml"
- entry with path = "proto/v1.proto"

---

### TC-20: External with fragments preserved

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter:
  ```
  external:
    - path: f.txt
      fragments:
        - lines: "1-10"
          hash: abc
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`external` contains one entry with:
- `path` = "f.txt"
- `fragments` preserved as-is: one fragment with lines = "1-10" and hash = "abc"

---

### TC-21: Empty external — no entries

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter (no external field)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`external` = empty list.

---

## Input

### TC-22: Input resolved from generating node

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `input: "ARTIFACT/b(data)"`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`input` = ChainItem with:
- `logical_name` = "ARTIFACT/b(data)"
- `file_path` = "out/data.json"
- `qualifier` = "data"

---

### TC-23: No input — absent

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter (no input field)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

`input` = absent.

---

### TC-24: Input without qualifier — error

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `input: "ARTIFACT/b"`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

### TC-25: Input with non-existent output id — error

**Setup**

Create three files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `input: "ARTIFACT/b(missing)"`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

## Error Cases

### TC-26: Unrecognized prefix in depends_on

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["UNKNOWN/something"]`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unresolvable artifact".

---

### TC-27: Invalid target logical name

**Setup**

No spec tree required.

**Action**

Call `ChainResolve("INVALID/something")`.

**Expected outcome**

Raises an error propagated from `LogicalNameGetParent` or
`LogicalNameToPath`. The exact error message is determined by
those functions.

---

### TC-28: Unreadable frontmatter

**Setup**

Create two files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — file contains invalid YAML in the frontmatter block
  (e.g., malformed YAML that cannot be parsed)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error "unreadable frontmatter".
