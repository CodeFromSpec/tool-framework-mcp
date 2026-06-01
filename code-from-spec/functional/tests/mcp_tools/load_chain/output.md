<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@NMqCZ4PTtyGuSQsUCgtZXf_OWFQ -->

# Test Specification: MCPLoadChain

Each test case creates a spec tree on disk with `_node.md` files, then calls
`MCPLoadChain`. Results are asserted against an `MCPLoadChainResult` record.

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
Root public content line one.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Leaf A public content.

# Agent
Leaf A agent guidance.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- Result is an `MCPLoadChainResult`.
- `chain_hash` is a string of exactly 27 characters.
- `context` contains the `# Public` heading followed by "Root public content line one."
- `context` contains a frontmatter block delimited by `---` lines, containing only the `outputs` field.
- `context` contains the `# Public` heading followed by "Leaf A public content."
- `context` contains the `# Agent` heading followed by "Leaf A agent guidance."
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
name: a
---
# Public
A public content.
```

Create `ROOT/a/b/_node.md`:
```
---
outputs:
  - id: main
    path: out/b.txt
---
# Public
B public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a/b"`.

**Expected outcome**

- `context` contains the `# Public` heading followed by "Root public content."
- `context` contains the `# Public` heading followed by "A public content."
- ROOT's public content appears before ROOT/a's public content in `context`.

---

### TC-03: Ancestor without public section skipped

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Name
Root name section only.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error is raised.
- `context` does not contain "Root name section only." or any ROOT content.
- `context` contains "A public content."

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

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error is raised.
- `context` does not contain ROOT's `# Public` heading or any ROOT content.
- `context` contains "A public content."

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

Create `ROOT/b/_node.md`:
```
---
name: b
---
# Public
B intro content.

## Interface
B interface details.

## Constraints
B constraint details.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains "B intro content."
- `context` contains the `## Interface` heading and "B interface details."
- `context` contains the `## Constraints` heading and "B constraint details."

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

Create `ROOT/b/_node.md`:
```
---
name: b
---
# Public
B intro content.

## Interface
B interface details.

## Constraints
B constraint details.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains the `## Interface` heading and "B interface details."
- `context` does not contain the `## Constraints` heading or "B constraint details."
- `context` does not contain "B intro content."

---

### TC-07: ARTIFACT dependency — content minus frontmatter

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
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
package main

// Body content of b.go
```

Create `ROOT/a/_node.md`:
```
---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains "package main" and "// Body content of b.go".
- `context` does not contain "some: frontmatter" or the `---` delimiters from the artifact frontmatter.

---

### TC-08: External file — full content

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
```

Create `data/config.yaml`:
```
key: value
setting: enabled
```

Create `ROOT/a/_node.md`:
```
---
external:
  - path: data/config.yaml
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains "key: value" and "setting: enabled".

---

### TC-09: External file with fragments — line ranges only

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
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

Create `ROOT/a/_node.md`:
```
---
external:
  - path: data/big.txt
    fragments:
      - lines: "2-4"
        hash: ignored
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains "line 2", "line 3", and "line 4".
- `context` does not contain "line 1", "line 5", "line 6", "line 7", "line 8", "line 9", or "line 10".

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

Create `ROOT/b/_node.md`:
```
---
name: b
---
# Public
B public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains a frontmatter block delimited by `---` lines.
- That frontmatter block contains the `outputs` field.
- That frontmatter block does not contain the `depends_on` field.

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
  - id: main
    path: out/a.txt
---
# Public
A public content.

# Agent
A agent guidance.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `context` contains the `# Public` heading and "A public content."
- `context` contains the `# Agent` heading and "A agent guidance."

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error is raised.
- `context` contains "A public content."
- `context` does not contain a `# Agent` heading.

---

### TC-13: Input separated from context

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
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
artifact: frontmatter
---
{"key": "value", "count": 42}
```

Create `ROOT/a/_node.md`:
```
---
input: ARTIFACT/b(data)
outputs:
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- `result.input` contains `{"key": "value", "count": 42}`.
- `result.input` does not contain "artifact: frontmatter" or the `---` frontmatter delimiters.
- `result.context` does not contain the body of `out/data.json`.

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- No error is raised.
- `result.input` is absent.

---

### TC-15: Hash is deterministic

**Setup**

Create `ROOT/_node.md`:
```
---
name: ROOT
---
# Public
Deterministic root content.
```

Create `ROOT/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Deterministic A content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"` twice, producing `result1` and `result2`.

**Expected outcome**

- `result1.chain_hash` equals `result2.chain_hash`.

---

## Error Cases

---

### TC-E01: Invalid logical name — not ROOT/

**Setup**

No files needed.

**Action**

Call `MCPLoadChain` with `logical_name = "INVALID/something"`.

**Expected outcome**

- An error is raised with name `UnsupportedReference` (propagated from LogicalNames via `LogicalNameToPath`).

---

### TC-E02: Nonexistent node file

**Setup**

No `_node.md` created for `ROOT/nonexistent`.

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/nonexistent"`.

**Expected outcome**

- An error is raised propagated from `FrontmatterParse` with name `FileUnreadable`.

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
name: a
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- An error is raised with name `NoOutputs`.

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
  - id: main
    path: ../../etc/passwd
---
# Public
A public content.
```

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- An error is raised with name `InvalidOutputPath`.

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
  - id: main
    path: out/a.txt
---
# Public
A public content.
```

Do not create `ROOT/missing/_node.md`.

**Action**

Call `MCPLoadChain` with `logical_name = "ROOT/a"`.

**Expected outcome**

- An error is raised. The error is detected during chain processing (hash computation or context building) because the missing node's file cannot be read.
