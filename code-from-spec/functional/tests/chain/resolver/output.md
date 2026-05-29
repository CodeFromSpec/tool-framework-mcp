<!-- code-from-spec: ROOT/functional/tests/chain/resolver@q3ZKratAUJoNpR_g5Xot2IOLgY4 -->

# Chain Resolver Tests

Each test creates a spec tree on disk with `_node.md` files, then calls
`ChainResolve` with a target logical name and checks the returned `Chain`
record.

---

## Ancestors and Target

### Test: Root as target

**Setup**

- Create `<root>/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT")`.

**Expected outcome**

- `ancestors` = empty list
- `target` = ChainItem(logical_name = "ROOT", qualifier = absent)
- `dependencies` = empty list
- `external` = empty list
- `input` = absent

---

### Test: Linear chain — ancestors in root-first order

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with empty frontmatter.
- Create `<root>/a/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a/b")`.

**Expected outcome**

- `ancestors` = [ChainItem("ROOT"), ChainItem("ROOT/a")] in that order
- `target` = ChainItem(logical_name = "ROOT/a/b", qualifier = absent)

---

### Test: Single parent

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `ancestors` = [ChainItem("ROOT")]
- `target` = ChainItem(logical_name = "ROOT/a", qualifier = absent)

---

### Test: Target with empty frontmatter

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with empty frontmatter (no depends_on, no
  external, no input).

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `ancestors` = [ChainItem("ROOT")]
- `target` = ChainItem(logical_name = "ROOT/a", qualifier = absent)
- `dependencies` = empty list
- `external` = empty list
- `input` = absent

---

## Dependencies — ROOT/ References

### Test: Dependency without qualifier

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains one ChainItem with
  logical_name = "ROOT/b", qualifier = absent.

---

### Test: Dependency with qualifier

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains one ChainItem with
  logical_name = "ROOT/b", qualifier = "interface".

---

### Test: Dependencies sorted by file path then qualifier

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/z
    - ROOT/m
    - ROOT/b
  ```
- Create `<root>/z/_node.md` with empty frontmatter.
- Create `<root>/m/_node.md` with empty frontmatter.
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` is sorted alphabetically by each ChainItem's file_path.
  The entry whose file_path resolves earliest alphabetically comes first,
  the one resolving latest comes last.

---

## Dependencies — ARTIFACT/ References

### Test: ARTIFACT dependency resolved from generating node

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains one ChainItem with
  logical_name = "ARTIFACT/b(lib)", file_path = "out/lib.go",
  qualifier = "lib".

---

### Test: ARTIFACT without qualifier — error

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

### Test: ARTIFACT — generating node has no outputs

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `<root>/b/_node.md` with empty frontmatter (no outputs declared).

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

### Test: ARTIFACT — artifact file does not exist on disk

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Do NOT create `out/lib.go` on disk.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- No error.
- `dependencies` contains one ChainItem with file_path = "out/lib.go".
  File existence on disk is not verified by the resolver.

---

### Test: ARTIFACT with non-existent output id — error

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(missing)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

### Test: Mixed ROOT/ and ARTIFACT/ dependencies

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/c
    - ARTIFACT/b(lib)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Create `<root>/c/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains both entries (one for ROOT/c, one for
  ARTIFACT/b(lib)), sorted alphabetically by each entry's resolved
  file_path value.

---

## Dependencies — Dedup

### Test: Exact duplicate — same file, same qualifier

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains exactly one entry for ROOT/b (not two).

---

### Test: No qualifier subsumes qualifier

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b(interface)
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains exactly one entry for ROOT/b with
  qualifier = absent. The qualified entry ROOT/b(interface) is removed.

---

### Test: Qualifier before no-qualifier — no-qualifier wins

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains exactly one entry for ROOT/b with
  qualifier = absent. Order of declaration does not affect the outcome.

---

### Test: Same file, different qualifiers — both kept

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b(constraints)
  ```
- Create `<root>/b/_node.md` with empty frontmatter.

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains two entries for ROOT/b:
  one with qualifier = "constraints", one with qualifier = "interface",
  sorted by qualifier alphabetically.

---

### Test: Duplicate ARTIFACT — same logical name

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
    - ARTIFACT/b(lib)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `dependencies` contains exactly one ARTIFACT entry for
  ARTIFACT/b(lib) (not two).

---

## External

### Test: External entries copied from frontmatter

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  external:
    - path: docs/api.yaml
    - path: proto/v1.proto
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `external` list contains both entries sorted alphabetically by path:
  docs/api.yaml before proto/v1.proto.

---

### Test: External with fragments preserved

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
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

- `external` list contains one entry with path = "f.txt" and
  fragments preserved as-is: [{lines: "1-10", hash: "abc"}].

---

### Test: Empty external — no entries

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with empty frontmatter (no external field).

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `external` list is empty.

---

## Input

### Test: Input resolved from generating node

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  input: ARTIFACT/b(data)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `input` = ChainItem with logical_name = "ARTIFACT/b(data)",
  file_path = "out/data.json", qualifier = "data".

---

### Test: No input — absent

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with empty frontmatter (no input field).

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- `input` is absent.

---

### Test: Input without qualifier — error

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  input: ARTIFACT/b
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

### Test: Input with non-existent output id — error

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  input: ARTIFACT/b(missing)
  ```
- Create `<root>/b/_node.md` with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

## Error Cases

### Test: Unrecognized prefix in depends_on

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with frontmatter:
  ```
  depends_on:
    - UNKNOWN/something
  ```

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unresolvable artifact" is raised.

---

### Test: Invalid target logical name

**Setup**

No spec tree needed.

**Action**

Call `ChainResolve("INVALID/something")`.

**Expected outcome**

- Error propagated from `LogicalNameGetParent` or `LogicalNameToPath`.

---

### Test: Unreadable frontmatter

**Setup**

- Create `<root>/_node.md` with empty frontmatter.
- Create `<root>/a/_node.md` with a `_node.md` file whose frontmatter
  block contains invalid YAML (e.g., malformed indentation or unclosed
  quotes).

**Action**

Call `ChainResolve("ROOT/a")`.

**Expected outcome**

- Error "unreadable frontmatter" is raised.
