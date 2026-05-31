<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@7DIrXPoxSqXG0VhDWtBb_p2RVCE -->

# Test Specification: MCPLoadChain

Each test case creates a spec tree on disk with `_node.md` files, then calls
`MCPLoadChain`. Setup describes the files to create (with their frontmatter
and body content). Actions describe what to call. Expected outcome describes
what the result or error must satisfy.

---

## Happy Path

---

### TC-01: Simple leaf node — context and hash

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Public
Root public content line.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
# Public
Node A public content.

# Agent
Node A agent content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- Result is of type `MCPLoadChainResult`.
- `chain_hash` is a string of exactly 27 characters.
- `context` contains ROOT's public content (`"Root public content line."`) without
  the `# Public` heading.
- `context` contains a frontmatter block between `---` delimiters listing only the
  `outputs` field for ROOT/a (no other frontmatter fields).
- `context` contains `"Node A public content."` without the `# Public` heading.
- `context` contains `"Node A agent content."` without the `# Agent` heading.
- `input` is absent.

---

### TC-02: Ancestor public content included

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Public
Root public content.
```

Create `ROOT/a/_node.md`:
```
---
name: A
---
# Public
Node A public content.
```

Create `ROOT/a/b/_node.md`:
```
---
outputs:
  - id: out
    path: out/b.txt
---
# Public
Node B public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a/b"`.

**Expected outcome**

- `context` contains `"Root public content."` (from ROOT), appearing before
  `"Node A public content."` (from ROOT/a).
- Neither `# Public` heading appears in `context`.

---

### TC-03: Ancestor without public section skipped

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Name
Root name section only — no public section.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
# Public
Node A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error.
- `context` does not contain any content from ROOT's `_node.md` body.
- `context` does contain `"Node A public content."`.

---

### TC-04: Ancestor with empty public section skipped

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Public
```
(The `# Public` section is present but has no content lines or subsections.)

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
# Public
Node A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error.
- `context` does not contain ROOT's content (empty public section is skipped).
- `context` does contain `"Node A public content."`.

---

### TC-05: Dependency without qualifier — public included

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: out
    path: out/a.txt
---
```

Create `ROOT/b/_node.md`:
```
---
name: B
---
# Public
## Interface
B interface content.

## Constraints
B constraints content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains ROOT/b's public content including both the `## Interface`
  heading with `"B interface content."` and the `## Constraints` heading with
  `"B constraints content."`.

---

### TC-06: Dependency with qualifier — subsection only

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b(interface)
outputs:
  - id: out
    path: out/a.txt
---
```

Create `ROOT/b/_node.md`:
```
---
name: B
---
# Public
## Interface
B interface content.

## Constraints
B constraints content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains `"B interface content."` (the Interface subsection).
- `context` does not contain `"B constraints content."` (Constraints not included).

---

### TC-07: ARTIFACT dependency — content minus frontmatter

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: out
    path: out/a.txt
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
// code-from-spec: ROOT/b@somehash
package main

func Hello() {}
```
(The file has a frontmatter-style artifact comment at the top and a body.)

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains the body content of `out/b.go` (the Go source lines).
- `context` does not contain the frontmatter of `out/b.go`.

---

### TC-08: External file — full content

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
external:
  - path: data/config.yaml
---
```

Create `data/config.yaml`:
```
key: value
other: 123
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains the full content of `data/config.yaml`:
  `"key: value"` and `"other: 123"`.

---

### TC-09: External file with fragments — line ranges only

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
external:
  - path: data/big.txt
    fragments:
      - lines: "2-4"
        hash: "ignored"
---
```

Create `data/big.txt` with exactly 10 lines:
```
line 1
line 2
line 3
line 4
line 5
line 6
line 7
line 8
line 9
line 10
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains `"line 2"`, `"line 3"`, and `"line 4"`.
- `context` does not contain `"line 1"`, `"line 5"`, or any line outside 2-4.

---

### TC-10: Target has reduced frontmatter with outputs only

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: out
    path: out/a.txt
---
```

Create `ROOT/b/_node.md`:
```
---
name: B
---
# Public
B content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains a frontmatter block delimited by `---` lines that includes
  the `outputs` field.
- The frontmatter block does not include the `depends_on` field.

---

### TC-11: Target agent section included

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
# Public
Node A public content.

# Agent
Node A agent guidance.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains `"Node A public content."` (without `# Public` heading).
- `context` contains `"Node A agent guidance."` (without `# Agent` heading).

---

### TC-12: Target without agent section — skipped

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
# Public
Node A public content.
```
(No `# Agent` section present.)

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error.
- `context` contains `"Node A public content."`.
- No agent-section heading or placeholder appears in `context`.

---

### TC-13: Input separated from context

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
input: ARTIFACT/b(data)
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
frontmatter: present
---
{"key": "value"}
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `result.input` contains `'{"key": "value"}'` (the body without frontmatter).
- `result.context` does not contain `'{"key": "value"}'` (input is separate).

---

### TC-14: No input — field absent

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: out/a.txt
---
```
(No `input` field in frontmatter.)

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error.
- `result.input` is absent.

---

### TC-15: Hash is deterministic

**Setup**

Create a spec tree: ROOT (with public section containing `"Deterministic content."`)
and ROOT/a (leaf with outputs and public section containing `"Leaf content."`).

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"` twice in sequence.

**Expected outcome**

- Both calls return a result without error.
- The `chain_hash` values from both calls are identical strings.

---

## Error Cases

---

### TC-E01: Invalid logical name — not ROOT/

**Setup**

No spec tree needed.

**Action**

Call `MCPLoadChain` with `logical_name = "INVALID/something"`.

**Expected outcome**

- Returns error `UnsupportedReference` (propagated from `LogicalNames` via
  `LogicalNameToPath`).

---

### TC-E02: Nonexistent node file

**Setup**

No `_node.md` file exists for `ROOT/nonexistent`.

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/nonexistent"`.

**Expected outcome**

- Returns error `FileUnreadable` (propagated from `FrontmatterParse`).

---

### TC-E03: No outputs declared

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
name: A
---
```
(No `outputs` field in frontmatter.)

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- Returns error `NoOutputs`.

---

### TC-E04: Invalid output path — traversal

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: out
    path: ../../etc/passwd
---
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- Returns error `InvalidOutputPath`.

---

### TC-E05: Unresolvable dependency

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ROOT/missing
outputs:
  - id: out
    path: out/a.txt
---
```
(Do not create `ROOT/missing/_node.md`.)

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- Returns an error. The missing node is detected during chain processing
  (hash computation or context building). The error is propagated from
  `ChainResolver` or `ChainHash`.
