<!-- code-from-spec: ROOT/functional/tests/chain/resolver@V5oBoUTGEzMTMy8ZiPBRRb4Ou_Y -->

## Ancestors and target

### Root as target

Setup: create `_node.md` for ROOT with no frontmatter.

Action: call `ChainResolve("ROOT")`.

Expected outcome:
- `Chain.ancestors` = empty
- `Chain.target` = `ChainItem(logical_name="ROOT", qualifier=absent)`
- `Chain.dependencies` = empty
- `Chain.external` = empty
- `Chain.input` = absent

---

### Linear chain — ancestors in root-first order

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a.
- Create `_node.md` for ROOT/a/b.

Action: call `ChainResolve("ROOT/a/b")`.

Expected outcome:
- `Chain.ancestors` = [`ChainItem(logical_name="ROOT")`, `ChainItem(logical_name="ROOT/a")`] in that order.
- `Chain.target` = `ChainItem(logical_name="ROOT/a/b")`

---

### Single parent

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.ancestors` = [`ChainItem(logical_name="ROOT")`]
- `Chain.target` = `ChainItem(logical_name="ROOT/a")`

---

### Target with empty frontmatter

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with empty frontmatter.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.ancestors` = [`ChainItem(logical_name="ROOT")`]
- `Chain.target` = `ChainItem(logical_name="ROOT/a")`
- `Chain.dependencies` = empty
- `Chain.external` = empty
- `Chain.input` = absent

---

## Dependencies — ROOT/ references

### Dependency without qualifier

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains one `ChainItem` with `logical_name="ROOT/b"`, `qualifier=absent`.

---

### Dependency with qualifier

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b(interface)"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains one `ChainItem` with `logical_name="ROOT/b"`, `qualifier="interface"`.

---

### Dependencies sorted by file path then qualifier

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/z", "ROOT/m", "ROOT/b"]`.
- Create `_node.md` for ROOT/z.
- Create `_node.md` for ROOT/m.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` is sorted alphabetically by file path: ROOT/b's path before ROOT/m's path before ROOT/z's path.

---

## Dependencies — ARTIFACT/ references

### ARTIFACT dependency resolved from generating node

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ARTIFACT/b"]`.
- Create `_node.md` for ROOT/b with frontmatter `output: "out/lib.go"`.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains one `ChainItem` with `logical_name="ARTIFACT/b"`, `file_path="out/lib.go"`.

---

### ARTIFACT — generating node has no output

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ARTIFACT/b"]`.
- Create `_node.md` for ROOT/b with empty frontmatter (no `output` field).

Action: call `ChainResolve("ROOT/a")`.

Expected outcome: error `UnresolvableArtifact`.

---

### ARTIFACT — artifact file does not exist on disk

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ARTIFACT/b"]`.
- Create `_node.md` for ROOT/b with frontmatter `output: "out/lib.go"`.
- Do NOT create `out/lib.go` on disk.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- No error.
- `Chain.dependencies` contains one `ChainItem` with `file_path="out/lib.go"`.

---

### Mixed ROOT/ and ARTIFACT/ dependencies

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/c", "ARTIFACT/b"]`.
- Create `_node.md` for ROOT/b with frontmatter `output: "out/lib.go"`.
- Create `_node.md` for ROOT/c.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains both entries, sorted by file path value.

---

## Dependencies — dedup

### Exact duplicate — same file, same qualifier

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b", "ROOT/b"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains exactly one entry for ROOT/b.

---

### No qualifier subsumes qualifier

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b", "ROOT/b(interface)"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains exactly one entry for ROOT/b with `qualifier=absent`. The `ROOT/b(interface)` entry is removed.

---

### Qualifier before no-qualifier — no-qualifier wins

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b(interface)", "ROOT/b"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains exactly one entry for ROOT/b with `qualifier=absent`.

---

### Same file, different qualifiers — both kept

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ROOT/b(interface)", "ROOT/b(constraints)"]`.
- Create `_node.md` for ROOT/b.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains two entries: one with `qualifier="constraints"`, one with `qualifier="interface"`, sorted by qualifier.

---

### Duplicate ARTIFACT — same logical name

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["ARTIFACT/b", "ARTIFACT/b"]`.
- Create `_node.md` for ROOT/b with frontmatter `output: "out/lib.go"`.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.dependencies` contains exactly one `ARTIFACT/b` entry.

---

## External

### External entries copied from frontmatter

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `external: [{path: "docs/api.yaml"}, {path: "proto/v1.proto"}]`.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.external` contains both entries sorted alphabetically: `docs/api.yaml` before `proto/v1.proto`.

---

### Empty external — no entries

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter and no `external` field.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.external` is empty.

---

## Input

### Input resolved from generating node

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `input: "ARTIFACT/b"`.
- Create `_node.md` for ROOT/b with frontmatter `output: "out/data.json"`.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.input` = `ChainItem(logical_name="ARTIFACT/b", file_path="out/data.json")`.

---

### No input — absent

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter and no `input` field.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome:
- `Chain.input` is absent.

---

## Error cases

### Unrecognized prefix in depends_on

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with frontmatter `depends_on: ["UNKNOWN/something"]`.

Action: call `ChainResolve("ROOT/a")`.

Expected outcome: error `UnresolvableArtifact`.

---

### Invalid target logical name

Setup: none (no spec tree required).

Action: call `ChainResolve("INVALID/something")`.

Expected outcome: error propagated from `LogicalNameGetParent` or `LogicalNameToPath`.

---

### Unreadable frontmatter

Setup:
- Create `_node.md` for ROOT.
- Create `_node.md` for ROOT/a with invalid YAML in frontmatter (e.g., malformed YAML content that cannot be parsed).

Action: call `ChainResolve("ROOT/a")`.

Expected outcome: error `UnreadableFrontmatter`.
