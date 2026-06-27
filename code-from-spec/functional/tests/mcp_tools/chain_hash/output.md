<!-- code-from-spec: SPEC/functional/tests/mcp_tools/chain_hash@FQ8Fm2kn5jssMYJKLc1w_Rgywak -->

## Test Suite: MCPChainHash

---

### TC-01: Returns a 27-character hash

**Setup**

Create a spec tree on disk:

- `SPEC/_node.md` with content:
  ```
  # Public

  ## Context

  Root node context.
  ```

- `SPEC/a/_node.md` with content:
  ```
  ---
  output: some/output/path.md
  ---

  # Public

  ## Context

  Leaf node content.
  ```

**Actions**

1. Call `MCPChainHash` with `logical_name = "SPEC/a"`.

**Expected Outcome**

Result is a string of exactly 27 characters.

---

### TC-02: Hash is deterministic

**Setup**

Create a spec tree on disk:

- `SPEC/_node.md` with content:
  ```
  # Public

  ## Context

  Root node context.
  ```

- `SPEC/a/_node.md` with content:
  ```
  ---
  output: some/output/path.md
  ---

  # Public

  ## Context

  Leaf node content.
  ```

**Actions**

1. Call `MCPChainHash` with `logical_name = "SPEC/a"`. Record result as `hash_1`.
2. Call `MCPChainHash` again with `logical_name = "SPEC/a"`. Record result as `hash_2`.

**Expected Outcome**

`hash_1` equals `hash_2`.

---

### TC-03: Hash matches load_chain hash

**Setup**

Create a spec tree on disk:

- `SPEC/_node.md` with content:
  ```
  # Public

  ## Context

  Root node context.
  ```

- `SPEC/a/_node.md` with content:
  ```
  ---
  output: some/output/path.md
  ---

  # Public

  ## Context

  Leaf node content.
  ```

**Actions**

1. Call `MCPChainHash` with `logical_name = "SPEC/a"`. Record result as `chain_hash_result`.
2. Call `MCPLoadChain` with `logical_name = "SPEC/a"`. Parse the first line of the response to extract the value after `"chain_hash: "`. Record it as `load_chain_hash`.

**Expected Outcome**

`chain_hash_result` equals `load_chain_hash`.

---

### TC-04: Invalid logical name — not SPEC/

**Setup**

No spec tree required.

**Actions**

1. Call `MCPChainHash` with `logical_name = "INVALID/something"`.

**Expected Outcome**

Error `logicalnames.UnsupportedReference` is raised.

---

### TC-05: Nonexistent node file

**Setup**

Create a spec tree on disk:

- `SPEC/_node.md` with any valid content.

Do not create `SPEC/nonexistent/_node.md`.

**Actions**

1. Call `MCPChainHash` with `logical_name = "SPEC/nonexistent"`.

**Expected Outcome**

Error `filereader.FileUnreadable` is raised.

---

### TC-06: No output declared

**Setup**

Create a spec tree on disk:

- `SPEC/_node.md` with content:
  ```
  # Public

  ## Context

  Root node context.
  ```

- `SPEC/a/_node.md` with content (no `output` field in frontmatter):
  ```
  # Public

  ## Context

  Leaf node without output.
  ```

**Actions**

1. Call `MCPChainHash` with `logical_name = "SPEC/a"`.

**Expected Outcome**

Error `NoOutput` is raised.
