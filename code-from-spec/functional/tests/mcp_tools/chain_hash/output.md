<!-- code-from-spec: ROOT/functional/tests/mcp_tools/chain_hash@biMFO7zheQRHEuN3aU-cVcE2fOg -->

## Test suite: MCPChainHash

---

### TC-01: Returns a 27-character hash

**Setup**

Create a spec tree on disk:
- `<root>/SPEC/_node.md` — contains a `# Public` section with a `## Context` subsection and some content inside it.
- `<root>/SPEC/a/_node.md` — contains frontmatter with `output: some/output/path.md` and a `# Public` section with a `## Context` subsection.

**Actions**

1. Call `MCPChainHash("SPEC/a")`.

**Expected outcome**

Result is a string of exactly 27 characters.

---

### TC-02: Hash is deterministic

**Setup**

Create a spec tree on disk:
- `<root>/SPEC/_node.md` — contains a `# Public` section with a `## Context` subsection and fixed known content.
- `<root>/SPEC/a/_node.md` — contains frontmatter with `output: some/output/path.md` and a `# Public` section with a `## Context` subsection with fixed known content.

**Actions**

1. Call `MCPChainHash("SPEC/a")` and store the result as `hash1`.
2. Call `MCPChainHash("SPEC/a")` again and store the result as `hash2`.

**Expected outcome**

`hash1` equals `hash2`.

---

### TC-03: Hash matches load_chain hash

**Setup**

Create a spec tree on disk:
- `<root>/SPEC/_node.md` — contains a `# Public` section with a `## Context` subsection and some content inside it.
- `<root>/SPEC/a/_node.md` — contains frontmatter with `output: some/output/path.md` and a `# Public` section with a `## Context` subsection.

**Actions**

1. Call `MCPChainHash("SPEC/a")` and store the result as `chain_hash_result`.
2. Call `MCPLoadChain("SPEC/a")` and extract the `chain_hash` value from the first line of its response.

**Expected outcome**

`chain_hash_result` equals the `chain_hash` extracted from the `MCPLoadChain` response.

---

### TC-04: Invalid logical name — not SPEC/

**Setup**

No spec tree required.

**Actions**

1. Call `MCPChainHash("INVALID/something")`.

**Expected outcome**

Raises error `logicalnames.UnsupportedReference`.

---

### TC-05: Nonexistent node file

**Setup**

Create a spec tree on disk where `<root>/SPEC/_node.md` exists, but `<root>/SPEC/nonexistent/_node.md` does not exist.

**Actions**

1. Call `MCPChainHash("SPEC/nonexistent")`.

**Expected outcome**

Raises error `filereader.FileUnreadable`.

---

### TC-06: No output declared

**Setup**

Create a spec tree on disk:
- `<root>/SPEC/_node.md` — contains a `# Public` section with a `## Context` subsection.
- `<root>/SPEC/a/_node.md` — contains a `# Public` section with a `## Context` subsection but no `output` field in frontmatter.

**Actions**

1. Call `MCPChainHash("SPEC/a")`.

**Expected outcome**

Raises error `NoOutput`.
