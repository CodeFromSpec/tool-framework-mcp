<!-- code-from-spec: ROOT/functional/tests/chain/resolver@KR4iYQ8W2ozS8KIBiu-f0XxbhKs -->

# ChainResolve — Test Specification

Each test case describes: files to create on disk (setup), the call
to `ChainResolve`, and the expected outcome. All spec nodes are
represented by `_node.md` files. Frontmatter is written in YAML
between `---` delimiters at the top of each file.

---

## Ancestors and Target

### TC-AT-01: Root as target

**Setup**

Create file `<root>/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT")`.

**Expected outcome**

Return a Chain where:
- ancestors = empty list
- target = ChainItem(logical_name="ROOT", qualifier=absent)
- dependencies = empty list
- external = empty list
- input = absent

---

### TC-AT-02: Linear chain — ancestors in root-first order

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter
- `<root>/a/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a/b")`.

**Expected outcome**

Return a Chain where:
- ancestors = [ChainItem(logical_name="ROOT"), ChainItem(logical_name="ROOT/a")] in that order
- target = ChainItem(logical_name="ROOT/a/b", qualifier=absent)

---

### TC-AT-03: Single parent

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Return a Chain where:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem(logical_name="ROOT/a", qualifier=absent)

---

### TC-AT-04: Target with empty frontmatter

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — empty frontmatter (no depends_on, no external, no input)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Return a Chain where:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem(logical_name="ROOT/a", qualifier=absent)
- dependencies = empty list
- external = empty list
- input = absent

---

## Dependencies — ROOT/ References

### TC-DEP-01: Dependency without qualifier

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains one ChainItem where:
- logical_name = "ROOT/b"
- qualifier = absent

---

### TC-DEP-02: Dependency with qualifier

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains one ChainItem where:
- logical_name = "ROOT/b"
- qualifier = "interface"

---

### TC-DEP-03: Dependencies sorted by file path then qualifier

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/z", "ROOT/m", "ROOT/b"]`
- `<root>/z/_node.md` — empty frontmatter
- `<root>/m/_node.md` — empty frontmatter
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies is sorted alphabetically by file path:
- ChainItem(logical_name="ROOT/b") comes first
- ChainItem(logical_name="ROOT/m") comes second
- ChainItem(logical_name="ROOT/z") comes third

---

## Dependencies — ARTIFACT/ References

### TC-ART-01: ARTIFACT dependency resolved from generating node

**Setup**

Create files:
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

dependencies contains one ChainItem where:
- logical_name = "ARTIFACT/b(lib)"
- file_path = "out/lib.go"
- qualifier = "lib"

---

### TC-ART-02: ARTIFACT without qualifier — error

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b"]`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error `UnresolvableArtifact`.

---

### TC-ART-03: ARTIFACT — generating node has no outputs

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — empty frontmatter (no outputs field)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error `UnresolvableArtifact`.

---

### TC-ART-04: ARTIFACT — artifact file does not exist on disk

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`
- `<root>/b/_node.md` — frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
Do NOT create the file `out/lib.go` on disk.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

No error raised. dependencies contains one ChainItem where:
- file_path = "out/lib.go"

File existence is not verified by the resolver.

---

### TC-ART-05: ARTIFACT with non-existent output id — error

**Setup**

Create files:
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

Raises error `UnresolvableArtifact`.

---

### TC-ART-06: Mixed ROOT/ and ARTIFACT/ dependencies

**Setup**

Create files:
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

dependencies contains two entries — ChainItem for "ROOT/c" and
ChainItem for "ARTIFACT/b(lib)" — sorted by file path value.

---

## Dependencies — Dedup

### TC-DEDUP-01: Exact duplicate — same file, same qualifier

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b", "ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains exactly one entry for ROOT/b (not two).

---

### TC-DEDUP-02: No qualifier subsumes qualifier

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b", "ROOT/b(interface)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains exactly one entry for ROOT/b where:
- qualifier = absent

The ROOT/b(interface) entry is removed because the unqualified
entry subsumes it.

---

### TC-DEDUP-03: Qualifier before no-qualifier — no-qualifier wins

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains exactly one entry for ROOT/b where:
- qualifier = absent

Regardless of declaration order, the unqualified entry wins.

---

### TC-DEDUP-04: Same file, different qualifiers — both kept

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b(constraints)"]`
- `<root>/b/_node.md` — empty frontmatter

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

dependencies contains two entries (both kept, sorted by qualifier):
- ChainItem(logical_name="ROOT/b", qualifier="constraints")
- ChainItem(logical_name="ROOT/b", qualifier="interface")

---

### TC-DEDUP-05: Duplicate ARTIFACT — same logical name

**Setup**

Create files:
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

dependencies contains exactly one ARTIFACT entry (not two).

---

## External

### TC-EXT-01: External entries copied from frontmatter

**Setup**

Create files:
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

external list contains two entries, sorted alphabetically:
- entry with path = "docs/api.yaml"
- entry with path = "proto/v1.proto"

---

### TC-EXT-02: External with fragments preserved

**Setup**

Create files:
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

external list contains one entry where:
- path = "f.txt"
- fragments list is preserved as-is with lines = "1-10" and hash = "abc"

---

### TC-EXT-03: Empty external — no entries

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter with no `external` field

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

external list is empty.

---

## Input

### TC-INP-01: Input resolved from generating node

**Setup**

Create files:
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

input = ChainItem where:
- logical_name = "ARTIFACT/b(data)"
- file_path = "out/data.json"
- qualifier = "data"

---

### TC-INP-02: No input — absent

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter with no `input` field

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

input is absent.

---

### TC-INP-03: Input without qualifier — error

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `input: "ARTIFACT/b"`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error `UnresolvableArtifact`.

---

### TC-INP-04: Input with non-existent output id — error

**Setup**

Create files:
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

Raises error `UnresolvableArtifact`.

---

## Error Cases

### TC-ERR-01: Unrecognized prefix in depends_on

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — frontmatter: `depends_on: ["UNKNOWN/something"]`

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error `UnresolvableArtifact`.

---

### TC-ERR-02: Invalid target logical name

**Setup**

No spec tree needed.

**Action**

Call `ChainResolve("INVALID/something")`.

**Expected outcome**

Raises an error propagated from `LogicalNameGetParent` or
`LogicalNameToPath`.

---

### TC-ERR-03: Unreadable frontmatter

**Setup**

Create files:
- `<root>/_node.md` — empty frontmatter
- `<root>/a/_node.md` — file contains invalid YAML in frontmatter
  (e.g., malformed indentation or unclosed quotes between `---` delimiters)

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

Raises error `UnreadableFrontmatter`.
