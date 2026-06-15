<!-- code-from-spec: SPEC/functional/tests/mcp_tools/load_chain@xgBm4ZowYRRNyBwk3svrH12Z9kg -->

## Test suite: MCPLoadChain

Each test creates a spec tree on disk using `_node.md` files, calls `MCPLoadChain`,
and parses the result by splitting on delimiter lines:
`--- context ---`, `--- input ---`, `--- existing artifact ---`.
The first line is always `chain_hash: <hash>`.

---

### Happy path

---

#### TC-01: Simple leaf node — context and hash

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  # Public
  ## Context
  Root context line.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  # Public
  ## Interface
  Interface description.
  # Agent
  Agent guidance.
  ```
- Do not create `out/a.txt`.

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- First line matches `chain_hash: ` followed by exactly 27 non-whitespace characters.
- After `--- context ---`:
  - Contains `## Context` heading and `Root context line.`
  - Does not contain `# Public` heading.
  - Contains a frontmatter block delimited by `---` / `---` with only the `output` field set to `out/a.txt`.
  - Contains `## Interface` heading and `Interface description.`
  - Contains `# Agent` heading and `Agent guidance.`
  - Does not contain `# Public` heading.
- No `--- input ---` section present.
- No `--- existing artifact ---` section present.

---

#### TC-02: Ancestor public content included

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  # Public
  ## Overview
  Root overview.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  ---
  # Public
  ## Details
  Node a details.
  ```
- Create `SPEC/a/b/_node.md`:
  ```
  ---
  output: out/b.txt
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a/b")`.

Expected outcome:
- After `--- context ---`:
  - Contains `## Overview` heading and `Root overview.`
  - Contains `## Details` heading and `Node a details.`
  - No `# Public` headings appear anywhere.

---

#### TC-03: Ancestor without public section skipped

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  # Name
  Root.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  # Public
  ## Interface
  Node a interface.
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Does not contain SPEC's `# Name` section content.
  - Contains `## Interface` heading and `Node a interface.`

---

#### TC-04: Ancestor with empty public section skipped

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  # Public
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  # Public
  ## Interface
  Node a interface.
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Does not contain SPEC's content.
  - Contains `## Interface` heading and `Node a interface.`

---

#### TC-05: Dependency without qualifier — full public included

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/b/_node.md`:
  ```
  ---
  ---
  # Public
  ## Interface
  Node b interface.
  ## Constraints
  Node b constraints.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  depends_on:
    - SPEC/b
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains `## Interface` heading and `Node b interface.`
  - Contains `## Constraints` heading and `Node b constraints.`

---

#### TC-06: Dependency with qualifier — subsection only

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/b/_node.md`:
  ```
  ---
  ---
  # Public
  ## Interface
  Node b interface.
  ## Constraints
  Node b constraints.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  depends_on:
    - SPEC/b(interface)
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains `## Interface` heading and `Node b interface.`
  - Does not contain `## Constraints` heading or `Node b constraints.`

---

#### TC-07: ARTIFACT dependency — artifact tag line removed

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/b/_node.md`:
  ```
  ---
  output: out/b.go
  ---
  ```
- Create `out/b.go`:
  ```
  // code-from-spec: SPEC/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa
  package main
  // body content
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.go
  depends_on:
    - ARTIFACT/b
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains `package main` and `// body content`.
  - Does not contain the artifact tag line
    `// code-from-spec: SPEC/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa`.

---

#### TC-08: EXTERNAL dependency — full content

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  depends_on:
    - EXTERNAL/data/config.yaml
  ---
  ```
- Create `data/config.yaml`:
  ```
  key: value
  setting: enabled
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains `key: value` and `setting: enabled`.

---

#### TC-09: Target has reduced frontmatter with output only

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  depends_on:
    - SPEC/b
  ---
  ```
- Create `SPEC/b/_node.md`:
  ```
  ---
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains a frontmatter block delimited by `---` / `---`.
  - The frontmatter block contains `output: out/a.txt`.
  - The frontmatter block does not contain `depends_on`.

---

#### TC-10: Target agent section included

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  # Public
  ## Interface
  Public interface.
  # Agent
  Agent-specific guidance.
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- After `--- context ---`:
  - Contains `## Interface` heading and `Public interface.`
  - Contains `# Agent` heading and `Agent-specific guidance.`
  - Does not contain `# Public` heading.

---

#### TC-11: Target without agent section — skipped

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  # Public
  ## Interface
  Public interface.
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- No error.
- After `--- context ---`:
  - Contains `## Interface` heading and `Public interface.`
  - No `# Agent` heading.

---

#### TC-12: Input present — in separate section

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/b/_node.md`:
  ```
  ---
  output: out/data.json
  ---
  ```
- Create `out/data.json`:
  ```
  // code-from-spec: SPEC/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa
  {"key": "value"}
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  input: ARTIFACT/b
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Result contains `--- input ---` section.
- After `--- input ---`:
  - Contains `{"key": "value"}`.
  - Does not contain the artifact tag line.
- The input content does not appear in the `--- context ---` section.

---

#### TC-13: EXTERNAL input — full content in input section

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  input: EXTERNAL/docs/vendor/spec.yaml
  ---
  ```
- Create `docs/vendor/spec.yaml`:
  ```
  version: 1
  title: Vendor spec
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Result contains `--- input ---` section.
- After `--- input ---`:
  - Contains `version: 1` and `title: Vendor spec`.

---

#### TC-14: No input — section absent

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Result does not contain `--- input ---`.

---

#### TC-15: Existing artifact present — in separate section

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.go
  ---
  ```
- Create `out/a.go`:
  ```
  package main
  func main() {}
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Result contains `--- existing artifact ---` section.
- After `--- existing artifact ---`:
  - Contains `package main` and `func main() {}`.

---

#### TC-16: Existing artifact absent — section omitted

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.go
  ---
  ```
- Do not create `out/a.go`.

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Result does not contain `--- existing artifact ---`.

---

#### TC-17: Hash is deterministic

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  # Public
  ## Overview
  Stable content.
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")` twice.

Expected outcome:
- Both calls return a result whose first line is identical.
- The 27-character hash portion is the same in both results.

---

### Error cases

---

#### TC-18: Invalid logical name — not SPEC/

Action: Call `MCPLoadChain("INVALID/something")`.

Expected outcome:
- Returns error `logicalnames.UnsupportedReference`.

---

#### TC-19: Nonexistent node file

Action: Call `MCPLoadChain("SPEC/nonexistent")` with no `_node.md` on disk for that path.

Expected outcome:
- Returns error propagated from `FrontmatterParse` (`FileUnreadable`).

---

#### TC-20: No output declared

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Returns error `NoOutput`.

---

#### TC-21: Invalid output path — path traversal

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: ../../etc/passwd
  ---
  ```

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Returns error `InvalidOutputPath`.

---

#### TC-22: Unresolvable dependency

Setup:
- Create `SPEC/_node.md`:
  ```
  ---
  ---
  ```
- Create `SPEC/a/_node.md`:
  ```
  ---
  output: out/a.txt
  depends_on:
    - SPEC/missing
  ---
  ```
- Do not create `SPEC/missing/_node.md`.

Action: Call `MCPLoadChain("SPEC/a")`.

Expected outcome:
- Returns an error — the missing node is detected during chain processing
  (hash computation or context building).
