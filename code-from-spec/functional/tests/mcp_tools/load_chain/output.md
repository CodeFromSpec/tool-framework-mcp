<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@hLEXp8jDxErw_uxUBPm6cBuiGEA -->

## Test Cases for MCPLoadChain

---

### TC-01: Simple leaf node — context and hash

**Setup**

Create `ROOT/_node.md`:
```
# Public

Root public content.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
# Public

Node a public content.

# Agent

Node a agent content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

Result is an `MCPLoadChainResult` where:
- `chain_hash` is a string of exactly 27 characters.
- `context` contains, in order:
  - The `# Public` heading and "Root public content." from ROOT.
  - A frontmatter block delimited by `---` lines containing only the `outputs` field.
  - The `# Public` heading and "Node a public content." from ROOT/a.
  - The `# Agent` heading and "Node a agent content." from ROOT/a.
- `input` is absent.

---

### TC-02: Ancestor public content included

**Setup**

Create `ROOT/_node.md`:
```
# Public

Root public content.
```

Create `ROOT/a/_node.md`:
```
# Public

Node a public content.
```

Create `ROOT/a/b/_node.md`:
```
---
outputs:
  - id: main
    path: out/b.go
---
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a/b"`.

**Expected outcome**

`context` contains ROOT's `# Public` heading and "Root public content." followed by ROOT/a's `# Public` heading and "Node a public content.", in ancestor-first order.

---

### TC-03: Ancestor without public section skipped

**Setup**

Create `ROOT/_node.md`:
```
# Name

Root name content only.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
# Public

Node a public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` does not contain ROOT's content. ROOT was skipped because it has no `# Public` section.

---

### TC-04: Ancestor with empty public section skipped

**Setup**

Create `ROOT/_node.md`:
```
# Public
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
# Public

Node a public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` does not contain ROOT's content. ROOT was skipped because its `# Public` section is empty.

---

### TC-05: Dependency without qualifier — full public included

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.go
---
```

Create `ROOT/b/_node.md`:
```
# Public

## Interface

Interface content.

## Constraints

Constraints content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains ROOT/b's `# Public` section including both the `## Interface` heading with its content and the `## Constraints` heading with its content.

---

### TC-06: Dependency with qualifier — subsection only

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b(interface)
outputs:
  - id: main
    path: out/a.go
---
```

Create `ROOT/b/_node.md`:
```
# Public

## Interface

Interface content.

## Constraints

Constraints content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains the `## Interface` heading and "Interface content." from ROOT/b. `context` does not contain the `## Constraints` heading or "Constraints content.".

---

### TC-07: ARTIFACT dependency — content minus frontmatter

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: main
    path: out/a.go
---
```

Create `ROOT/b/_node.md`:
```
---
outputs:
  - id: code
    path: out/b.go
---
```

Create `out/b.go`:
```
---
some: frontmatter
---
Body content of b.go.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains "Body content of b.go." and does not contain the frontmatter block from `out/b.go`.

---

### TC-08: External file — full content

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
external:
  - path: data/config.yaml
outputs:
  - id: main
    path: out/a.go
---
```

Create `data/config.yaml`:
```
key: value
other: data
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains the full content of `data/config.yaml`, including "key: value" and "other: data".

---

### TC-09: Target has reduced frontmatter with outputs only

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.go
---
```

Create `ROOT/b/_node.md` (minimal, no public section).

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains a frontmatter block delimited by `---` lines. That block contains only the `outputs` field. The `depends_on` field is not present in that block.

---

### TC-10: Target agent section included

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
# Public

Node a public content.

# Agent

Node a agent content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`context` contains the `# Public` heading and "Node a public content." and also the `# Agent` heading and "Node a agent content.".

---

### TC-11: Target without agent section — skipped

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
# Public

Node a public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

No error. `context` contains the public content. No `# Agent` heading appears in `context`.

---

### TC-12: Input separated from context

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
input: ARTIFACT/b(data)
outputs:
  - id: main
    path: out/a.go
---
```

Create `ROOT/b/_node.md`:
```
---
outputs:
  - id: data
    path: out/data.json
---
```

Create `out/data.json`:
```
---
some: frontmatter
---
Body content of data.json.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `result.input` contains "Body content of data.json." without the frontmatter block.
- `result.context` does not contain the input body content.

---

### TC-13: No input — field absent

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.go
---
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

`result.input` is absent.

---

### TC-14: Hash is deterministic

**Setup**

Create a spec tree: ROOT (with public section containing known content) and ROOT/a (leaf with outputs and public section with known content).

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"` twice.

**Expected outcome**

Both calls return an `MCPLoadChainResult` with identical `chain_hash` values.

---

### TC-15: Invalid logical name — not ROOT/

**Setup**

None.

**Action**

Call `MCPLoadChain` with `logical_name = "INVALID/something"`.

**Expected outcome**

Error `UnsupportedReference` is returned (propagated from `LogicalNameToPath`).

---

### TC-16: Nonexistent node file

**Setup**

No `_node.md` file exists at the path corresponding to `"ROOT/nonexistent"`.

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/nonexistent"`.

**Expected outcome**

Error `FileUnreadable` is returned (propagated from `FrontmatterParse` via `FileReader`).

---

### TC-17: No outputs declared

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
# Public

Node a public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

Error `NoOutputs` is returned.

---

### TC-18: Invalid output path — path traversal

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: ../../etc/passwd
---
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

Error `InvalidOutputPath` is returned.

---

### TC-19: Unresolvable dependency

**Setup**

Create `ROOT/_node.md` (minimal, no public section).

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/missing
outputs:
  - id: main
    path: out/a.go
---
```

No `ROOT/missing/_node.md` file is created.

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

An error is returned. The missing node is detected during chain processing (hash computation or context building).
