<!-- code-from-spec: ROOT/functional/tests/chain/resolver@C2OVkUKIZdO3Svhr0ph_0aal6QI -->

# ChainResolve — Test Specification

Each test creates a spec tree on disk using `_node.md` files with frontmatter,
then calls `ChainResolve` with a target logical name and checks the result.

---

## Ancestors and Target

### TC-01: Root as target

**Setup**
- Create `_node.md` at ROOT (empty frontmatter or no frontmatter).

**Action**
- Call `ChainResolve("ROOT")`.

**Expected outcome**
- `ancestors` = empty list.
- `target` = ChainItem with logical_name = "ROOT", qualifier = absent.
- `dependencies` = empty list.
- `external` = empty list.
- `input` = absent.

---

### TC-02: Linear chain — ancestors in root-first order

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a.
- Create `_node.md` at ROOT/a/b.

**Action**
- Call `ChainResolve("ROOT/a/b")`.

**Expected outcome**
- `ancestors` = [ChainItem(logical_name="ROOT"), ChainItem(logical_name="ROOT/a")] in that order.
- `target` = ChainItem with logical_name = "ROOT/a/b".

---

### TC-03: Single parent

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `ancestors` = [ChainItem(logical_name="ROOT")].
- `target` = ChainItem with logical_name = "ROOT/a".

---

### TC-04: Target with empty frontmatter

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with empty frontmatter (no depends_on, no external, no input).

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `ancestors` = [ChainItem(logical_name="ROOT")].
- `target` = ChainItem with logical_name = "ROOT/a".
- `dependencies` = empty list.
- `external` = empty list.
- `input` = absent.

---

## Dependencies — ROOT/ References

### TC-05: Dependency without qualifier

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains one ChainItem with logical_name = "ROOT/b", qualifier = absent.

---

### TC-06: Dependency with qualifier

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains one ChainItem with logical_name = "ROOT/b", qualifier = "interface".

---

### TC-07: Dependencies sorted by file path then qualifier

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/z
    - ROOT/m
    - ROOT/b
  ```
- Create `_node.md` at ROOT/z.
- Create `_node.md` at ROOT/m.
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` is sorted alphabetically by file_path.
- ROOT/b appears before ROOT/m, which appears before ROOT/z.

---

## Dependencies — ARTIFACT/ References

### TC-08: ARTIFACT dependency resolved from generating node

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains one ChainItem with:
  - logical_name = "ARTIFACT/b(lib)"
  - file_path = "out/lib.go"
  - qualifier = "lib"

---

### TC-09: ARTIFACT without qualifier — error

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised.

---

### TC-10: ARTIFACT — generating node has no outputs

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` at ROOT/b with empty frontmatter (no outputs declared).

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised.

---

### TC-11: ARTIFACT — artifact file does not exist on disk

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Do NOT create the file `out/lib.go` on disk.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- No error.
- `dependencies` contains one ChainItem with file_path = "out/lib.go".
- File existence on disk is not verified by the resolver.

---

### TC-12: ARTIFACT with non-existent output id — error

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(missing)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised (output id "missing" does not exist in ROOT/b).

---

### TC-13: Mixed ROOT/ and ARTIFACT/ dependencies

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/c
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Create `_node.md` at ROOT/c.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains two entries: the ChainItem for ROOT/c and the ChainItem for ARTIFACT/b(lib).
- Entries are sorted by file_path value.

---

## Dependencies — Dedup

### TC-14: Exact duplicate — same file, same qualifier

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains exactly one entry for ROOT/b (duplicate is removed).

---

### TC-15: No qualifier subsumes qualifier

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b(interface)
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains exactly one entry for ROOT/b with qualifier = absent.
- The ROOT/b(interface) entry is removed because the unqualified entry subsumes it.

---

### TC-16: Qualifier before no-qualifier — no-qualifier wins

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains exactly one entry for ROOT/b with qualifier = absent.
- Order in depends_on does not matter; the unqualified entry always wins.

---

### TC-17: Same file, different qualifiers — both kept

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b(constraints)
  ```
- Create `_node.md` at ROOT/b.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains two entries: one with qualifier = "constraints", one with qualifier = "interface".
- Both are retained because neither subsumes the other.
- Entries are sorted (qualifier "constraints" before "interface").

---

### TC-18: Duplicate ARTIFACT — same logical name

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `dependencies` contains exactly one ARTIFACT entry for ARTIFACT/b(lib) (duplicate is removed).

---

## External

### TC-19: External entries copied from frontmatter

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  external:
    - path: docs/api.yaml
    - path: proto/v1.proto
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `external` list contains two entries sorted alphabetically by path:
  - docs/api.yaml
  - proto/v1.proto

---

### TC-20: External with fragments preserved

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  external:
    - path: f.txt
      fragments:
        - lines: "1-10"
          hash: abc
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `external` list contains one entry with path = "f.txt".
- The fragments field is preserved as-is: one fragment with lines = "1-10" and hash = "abc".

---

### TC-21: Empty external — no entries

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter that has no `external` field.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `external` list is empty.

---

## Input

### TC-22: Input resolved from generating node

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b(data)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `input` = ChainItem with:
  - logical_name = "ARTIFACT/b(data)"
  - file_path = "out/data.json"
  - qualifier = "data"

---

### TC-23: No input — absent

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter that has no `input` field.

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- `input` is absent.

---

### TC-24: Input without qualifier — error

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised.

---

### TC-25: Input with non-existent output id — error

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b(missing)
  ```
- Create `_node.md` at ROOT/b with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised (output id "missing" does not exist in ROOT/b).

---

## Error Cases

### TC-26: Unrecognized prefix in depends_on

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  depends_on:
    - UNKNOWN/something
  ```

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnresolvableArtifact` is raised.

---

### TC-27: Invalid target logical name

**Setup**
- No spec tree required.

**Action**
- Call `ChainResolve("INVALID/something")`.

**Expected outcome**
- Error propagated from `LogicalNameGetParent` or `LogicalNameToPath`.

---

### TC-28: Unreadable frontmatter

**Setup**
- Create `_node.md` at ROOT.
- Create `_node.md` at ROOT/a with invalid YAML in the frontmatter block (malformed content).

**Action**
- Call `ChainResolve("ROOT/a")`.

**Expected outcome**
- Error `UnreadableFrontmatter` is raised.
