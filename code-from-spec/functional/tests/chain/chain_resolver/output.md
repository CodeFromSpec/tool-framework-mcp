<!-- code-from-spec: ROOT/functional/tests/chain/chain_resolver@bgp5bfkhI6Ubj5NL0A3m6TCluxI -->

# ChainResolve — Test Specification

All tests create a spec tree on disk using `_node.md` files containing
frontmatter as needed, then call `ChainResolve` with a target logical name.

Each test case describes:
- **Setup**: files to create and their frontmatter content.
- **Action**: the call to make.
- **Expected outcome**: what the returned `Chain` or error must look like.

---

## Ancestors and target

### Test: Root as target

Setup:
- Create `_node.md` for ROOT (empty frontmatter).

Action:
- Call ChainResolve("ROOT").

Expected outcome:
- ancestors = empty list
- target = ChainItem(logical_name="ROOT", qualifier=absent)
- dependencies = empty list
- external = empty list
- input = absent

---

### Test: Linear chain — ancestors in root-first order

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a (empty frontmatter).
- Create `_node.md` for ROOT/a/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a/b").

Expected outcome:
- ancestors = [ChainItem(logical_name="ROOT"), ChainItem(logical_name="ROOT/a")]
  in that order (root first).
- target = ChainItem(logical_name="ROOT/a/b", qualifier=absent)

---

### Test: Single parent

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem(logical_name="ROOT/a", qualifier=absent)

---

### Test: Target with empty frontmatter

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- ancestors = [ChainItem(logical_name="ROOT")]
- target = ChainItem(logical_name="ROOT/a", qualifier=absent)
- dependencies = empty list
- external = empty list
- input = absent

---

## Dependencies — ROOT/ references

### Test: Dependency without qualifier

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains one ChainItem:
  - logical_name = "ROOT/b"
  - qualifier = absent

---

### Test: Dependency with qualifier

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains one ChainItem:
  - logical_name = "ROOT/b"
  - qualifier = "interface"

---

### Test: Dependencies sorted by file path then qualifier

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/z
    - ROOT/m
    - ROOT/b
  ```
- Create `_node.md` for ROOT/z (empty frontmatter).
- Create `_node.md` for ROOT/m (empty frontmatter).
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies is sorted alphabetically by the resolved file_path of each entry.
  The entry whose file_path sorts first appears first (ROOT/b before ROOT/m
  before ROOT/z, assuming file paths follow the same alphabetical order as the
  logical names).

---

## Dependencies — ARTIFACT/ references

### Test: ARTIFACT dependency resolved from generating node

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains one ChainItem:
  - logical_name = "ARTIFACT/b(lib)"
  - file_path = "out/lib.go"
  - qualifier = "lib"

---

### Test: ARTIFACT without qualifier — error

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

### Test: ARTIFACT — generating node has no outputs

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` for ROOT/b (empty frontmatter, no outputs declared).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

### Test: ARTIFACT — artifact file does not exist on disk

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Do NOT create "out/lib.go" on disk.

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- No error.
- dependencies contains one ChainItem with file_path = "out/lib.go".
  (The resolver does not verify that the artifact file exists on disk.)

---

### Test: ARTIFACT with non-existent output id — error

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(missing)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

### Test: Mixed ROOT/ and ARTIFACT/ dependencies

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/c
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```
- Create `_node.md` for ROOT/c (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains two ChainItems (one for ROOT/c, one for ARTIFACT/b(lib)),
  sorted by their resolved file_path values alphabetically.

---

## Dependencies — dedup

### Test: Exact duplicate — same file, same qualifier

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains exactly one ChainItem for ROOT/b (not two).

---

### Test: No qualifier subsumes qualifier

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b
    - ROOT/b(interface)
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains exactly one ChainItem for ROOT/b with qualifier = absent.
  The ROOT/b(interface) entry is removed because the no-qualifier entry subsumes it.

---

### Test: Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains exactly one ChainItem for ROOT/b with qualifier = absent.
  Order of appearance in depends_on does not affect the outcome: no-qualifier
  always wins over any qualified entry for the same file.

---

### Test: Same file, different qualifiers — both kept

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ROOT/b(interface)
    - ROOT/b(constraints)
  ```
- Create `_node.md` for ROOT/b (empty frontmatter).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains two ChainItems for ROOT/b:
  - one with qualifier = "constraints"
  - one with qualifier = "interface"
  Both are kept (no subsumption applies). Sorted alphabetically by qualifier.

---

### Test: Duplicate ARTIFACT — same logical name

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - ARTIFACT/b(lib)
    - ARTIFACT/b(lib)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: lib
      path: out/lib.go
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- dependencies contains exactly one ARTIFACT ChainItem for ARTIFACT/b(lib)
  (not two).

---

## External

### Test: External entries copied from frontmatter

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  external:
    - path: docs/api.yaml
    - path: proto/v1.proto
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- external list contains two entries sorted alphabetically by path:
  - first: path = "docs/api.yaml"
  - second: path = "proto/v1.proto"

---

### Test: External with fragments preserved

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  external:
    - path: f.txt
      fragments:
        - lines: "1-10"
          hash: abc
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- external list contains one entry:
  - path = "f.txt"
  - fragments preserved as-is: one fragment with lines = "1-10" and hash = "abc"

---

### Test: Empty external — no entries

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a (empty frontmatter, no external field).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- external list is empty.

---

## Input

### Test: Input resolved from generating node

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b(data)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- input = ChainItem:
  - logical_name = "ARTIFACT/b(data)"
  - file_path = "out/data.json"
  - qualifier = "data"

---

### Test: No input — absent

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a (empty frontmatter, no input field).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- input is absent.

---

### Test: Input without qualifier — error

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

### Test: Input with non-existent output id — error

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  input: ARTIFACT/b(missing)
  ```
- Create `_node.md` for ROOT/b with frontmatter:
  ```
  outputs:
    - id: data
      path: out/data.json
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

## Error cases

### Test: Unrecognized prefix in depends_on

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with frontmatter:
  ```
  depends_on:
    - UNKNOWN/something
  ```

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unresolvable artifact".

---

### Test: Invalid target logical name

Setup:
- No spec tree required.

Action:
- Call ChainResolve("INVALID/something").

Expected outcome:
- Error propagated from LogicalNameGetParent or LogicalNameToPath.
  The exact error message is determined by those functions.

---

### Test: Unreadable frontmatter

Setup:
- Create `_node.md` for ROOT (empty frontmatter).
- Create `_node.md` for ROOT/a with invalid YAML in the frontmatter block
  (e.g., malformed indentation or unclosed delimiter that causes a parse failure).

Action:
- Call ChainResolve("ROOT/a").

Expected outcome:
- Error "unreadable frontmatter".
