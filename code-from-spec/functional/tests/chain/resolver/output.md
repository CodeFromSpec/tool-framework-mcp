<!-- code-from-spec: ROOT/functional/tests/chain/resolver@5w9PuCMhjZQnUoRqjwcjlrty8Vs -->

# Test Specification: ChainResolve

## Ancestors and Target

### Root as target

Setup:
- Create `_node.md` at ROOT with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT")`.

Expected:
- `ancestors` = empty list.
- `target` = ChainItem(logical_name="ROOT", qualifier=absent).
- `dependencies` = empty list.
- `external` = empty list.
- `input` = absent.

---

### Linear chain — ancestors in root-first order

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with empty frontmatter.
- Create `_node.md` at ROOT/a/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a/b")`.

Expected:
- `ancestors` = [ChainItem(ROOT), ChainItem(ROOT/a)] in that order.
- `target` = ChainItem(logical_name="ROOT/a/b", qualifier=absent).

---

### Single parent

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `ancestors` = [ChainItem(ROOT)].
- `target` = ChainItem(logical_name="ROOT/a", qualifier=absent).

---

### Target with empty frontmatter

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with empty frontmatter (leaf, no depends_on, no external, no input).

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `ancestors` = [ChainItem(ROOT)].
- `target` = ChainItem(logical_name="ROOT/a", qualifier=absent).
- `dependencies` = empty list.
- `external` = empty list.
- `input` = absent.

---

## Dependencies — ROOT/ References

### Dependency without qualifier

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains one ChainItem with logical_name="ROOT/b", qualifier=absent.

---

### Dependency with qualifier

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b(interface)"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains one ChainItem with logical_name="ROOT/b", qualifier="interface".

---

### Dependencies sorted by file path then qualifier

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/z", "ROOT/m", "ROOT/b"]`.
- Create `_node.md` at ROOT/z with empty frontmatter.
- Create `_node.md` at ROOT/m with empty frontmatter.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` is sorted alphabetically by file path (ROOT/b path, then ROOT/m path, then ROOT/z path).

---

## Dependencies — ARTIFACT/ References

### ARTIFACT dependency resolved from generating node

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "lib", path: "out/lib.go"}]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains one ChainItem with logical_name="ARTIFACT/b(lib)", file_path="out/lib.go", qualifier="lib".

---

### ARTIFACT without qualifier — error

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b"]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

### ARTIFACT — generating node has no outputs

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`.
- Create `_node.md` at ROOT/b with empty frontmatter (no outputs declared).

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

### ARTIFACT — artifact file does not exist on disk

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b(lib)"]`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "lib", path: "out/lib.go"}]`.
- Do NOT create the file `out/lib.go` on disk.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- No error.
- `dependencies` contains one ChainItem with file_path="out/lib.go".
- File existence is not verified by the resolver.

---

### ARTIFACT with non-existent output id — error

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b(missing)"]`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "lib", path: "out/lib.go"}]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

### Mixed ROOT/ and ARTIFACT/ dependencies

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/c", "ARTIFACT/b(lib)"]`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "lib", path: "out/lib.go"}]`.
- Create `_node.md` at ROOT/c with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains two entries (one for ROOT/c and one for ARTIFACT/b(lib)), sorted by their resolved file path values.

---

## Dependencies — Dedup

### Exact duplicate — same file, same qualifier

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b", "ROOT/b"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains exactly one entry for ROOT/b.

---

### No qualifier subsumes qualifier

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b", "ROOT/b(interface)"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains exactly one entry for ROOT/b with qualifier=absent.
- The entry with qualifier="interface" is removed.

---

### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains exactly one entry for ROOT/b with qualifier=absent.

---

### Same file, different qualifiers — both kept

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ROOT/b(interface)", "ROOT/b(constraints)"]`.
- Create `_node.md` at ROOT/b with empty frontmatter.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains two entries: one with qualifier="constraints" and one with qualifier="interface", sorted alphabetically by qualifier.

---

### Duplicate ARTIFACT — same logical name

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["ARTIFACT/b(lib)", "ARTIFACT/b(lib)"]`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "lib", path: "out/lib.go"}]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `dependencies` contains exactly one ARTIFACT entry for ARTIFACT/b(lib).

---

## External

### External entries copied from frontmatter

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter:
  ```
  external:
    - path: "docs/api.yaml"
    - path: "proto/v1.proto"
  ```

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `external` contains two entries sorted alphabetically: "docs/api.yaml" before "proto/v1.proto".

---

### Empty external — no entries

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with empty frontmatter (no external field).

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `external` = empty list.

---

## Input

### Input resolved from generating node

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `input: "ARTIFACT/b(data)"`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "data", path: "out/data.json"}]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `input` = ChainItem(logical_name="ARTIFACT/b(data)", file_path="out/data.json", qualifier="data").

---

### No input — absent

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with empty frontmatter (no input field).

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- `input` = absent.

---

### Input without qualifier — error

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `input: "ARTIFACT/b"`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

### Input with non-existent output id — error

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `input: "ARTIFACT/b(missing)"`.
- Create `_node.md` at ROOT/b with frontmatter: `outputs: [{id: "data", path: "out/data.json"}]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

## Error Cases

### Unrecognized prefix in depends_on

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with frontmatter: `depends_on: ["UNKNOWN/something"]`.

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnresolvableArtifact`.

---

### Invalid target logical name

Setup:
- No spec tree required.

Actions:
- Call `ChainResolve("INVALID/something")`.

Expected:
- Error propagated from `LogicalNameGetParent` or `LogicalNameToPath`.

---

### Unreadable frontmatter

Setup:
- Create `_node.md` at ROOT with empty frontmatter.
- Create `_node.md` at ROOT/a with invalid YAML content in frontmatter (malformed YAML between `---` delimiters).

Actions:
- Call `ChainResolve("ROOT/a")`.

Expected:
- Error `UnreadableFrontmatter`.
