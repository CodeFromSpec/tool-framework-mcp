<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@WwUcHRllEI6EDaCO3BfsWH1yIpE -->

# Test Specification: MCPLoadChain

## Interface

```
record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
```

---

## Happy Path Tests

---

### TC-01: Simple leaf node — context and hash

**Setup**

Create `<root>/_node.md`:
```
---
---
# Public
Root public content.
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Node a public content.
# Agent
Node a agent content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.chain_hash` is a 27-character string.
- `result.context` contains:
  - ROOT's public content ("Root public content."), without a `# Public` heading line.
  - A frontmatter block delimited by `---` lines containing only the `outputs` field.
  - ROOT/a's public content ("Node a public content."), without a `# Public` heading line.
  - ROOT/a's agent content ("Node a agent content."), without a `# Agent` heading line.
- `result.input` is absent.

---

### TC-02: Ancestor public content included

**Setup**

Create `<root>/_node.md`:
```
---
---
# Public
Root public content.
```

Create `<root>/a/_node.md`:
```
---
---
# Public
Node a public content.
```

Create `<root>/a/b/_node.md`:
```
---
outputs:
  - id: main
    path: out/b.txt
---
# Public
Node b public content.
```

**Action**

Call `MCPLoadChain("ROOT/a/b")`.

**Expected outcome**

- No error.
- `result.context` contains ROOT's public content ("Root public content.") followed by ROOT/a's public content ("Node a public content."), both without `# Public` heading lines.

---

### TC-03: Ancestor without public section — skipped

**Setup**

Create `<root>/_node.md`:
```
---
---
# Name
Root
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Node a public content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` does not contain any content from ROOT's `_node.md` (no "Root" text from the name section, and no `# Public` section was present to include).

---

### TC-04: Ancestor with empty public section — skipped

**Setup**

Create `<root>/_node.md`:
```
---
---
# Public
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Node a public content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` does not contain any content originating from ROOT's public section (the section exists but is empty, so it is skipped).

---

### TC-05: Dependency without qualifier — full public section included

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.txt
---
```

Create `<root>/b/_node.md`:
```
---
---
# Public
## Interface
Interface content.
## Constraints
Constraints content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains ROOT/b's public content including both the `## Interface` and `## Constraints` subsections with their headings and content.

---

### TC-06: Dependency with qualifier — subsection only

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
depends_on:
  - ROOT/b(interface)
outputs:
  - id: main
    path: out/a.txt
---
```

Create `<root>/b/_node.md`:
```
---
---
# Public
## Interface
Interface content.
## Constraints
Constraints content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains only the Interface subsection content from ROOT/b (including the `## Interface` heading and "Interface content.").
- `result.context` does not contain the Constraints subsection content from ROOT/b.

---

### TC-07: ARTIFACT dependency — content minus frontmatter

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
depends_on:
  - ARTIFACT/b(code)
outputs:
  - id: main
    path: out/a.txt
---
```

Create `<root>/b/_node.md`:
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

func main() {}
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains the body of `out/b.go` ("package main\n\nfunc main() {}"), without the frontmatter block.

---

### TC-08: External file — full content

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
external:
  - path: data/config.yaml
---
```

Create `data/config.yaml`:
```
key: value
other: 42
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains the full content of `data/config.yaml` ("key: value\nother: 42").

---

### TC-09: External file with fragments — line ranges only

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
external:
  - path: data/big.txt
    fragments:
      - lines: "2-4"
        hash: "ignored"
---
```

Create `data/big.txt` with 10 lines:
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

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains only lines 2 through 4 from `data/big.txt` ("line 2\nline 3\nline 4").
- Lines outside the range (1, 5–10) do not appear.

---

### TC-10: Target has reduced frontmatter with outputs only

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
depends_on:
  - ROOT/b
outputs:
  - id: main
    path: out/a.txt
---
```

Create `<root>/b/_node.md`:
```
---
---
# Public
Node b content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains a frontmatter block (between `---` delimiters) that has only the `outputs` field.
- The `depends_on` field is not present in that frontmatter block.

---

### TC-11: Target agent section included

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Public content.
# Agent
Agent content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains both "Public content." and "Agent content.", without `# Public` or `# Agent` heading lines.

---

### TC-12: Target without agent section — skipped

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Public only content.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.context` contains "Public only content.".
- No agent section content appears (none was defined).

---

### TC-13: Input separated from context

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
input: "ARTIFACT/b(data)"
---
```

Create `<root>/b/_node.md`:
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
meta: info
---
{"key": "value"}
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.input` contains the body of `out/data.json` ('{"key": "value"}'), without the frontmatter block.
- The body of `out/data.json` does not appear in `result.context`.

---

### TC-14: No input — field absent

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- No error.
- `result.input` is absent.

---

### TC-15: Hash is deterministic

**Setup**

Create a spec tree with known, fixed content:

Create `<root>/_node.md`:
```
---
---
# Public
Stable content.
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: main
    path: out/a.txt
---
# Public
Node a stable content.
```

**Action**

Call `MCPLoadChain("ROOT/a")` twice.

**Expected outcome**

- Both calls return no error.
- `result.chain_hash` from the first call equals `result.chain_hash` from the second call.

---

## Error Case Tests

---

### TC-16: Invalid logical name — not ROOT/

**Setup**

No spec tree required.

**Action**

Call `MCPLoadChain("INVALID/something")`.

**Expected outcome**

- Returns error `UnsupportedReference` (propagated from `LogicalNameToPath`).

---

### TC-17: Nonexistent node file

**Setup**

No `_node.md` is created for ROOT/nonexistent.

**Action**

Call `MCPLoadChain("ROOT/nonexistent")`.

**Expected outcome**

- Returns error propagated from `FrontmatterParse` indicating the file is unreadable (`FileUnreadable`).

---

### TC-18: No outputs declared

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
---
# Public
Content without outputs.
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- Returns error `NoOutputs`.

---

### TC-19: Invalid output path — path traversal

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
outputs:
  - id: evil
    path: ../../etc/passwd
---
```

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- Returns error `InvalidOutputPath`.

---

### TC-20: Unresolvable dependency

**Setup**

Create `<root>/_node.md`:
```
---
---
```

Create `<root>/a/_node.md`:
```
---
depends_on:
  - ROOT/missing
outputs:
  - id: main
    path: out/a.txt
---
```

Do not create `<root>/missing/_node.md`.

**Action**

Call `MCPLoadChain("ROOT/a")`.

**Expected outcome**

- Returns an error indicating the missing node was not found (propagated during chain resolution or hash computation when the missing node is accessed).
