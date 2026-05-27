<!-- code-from-spec: ROOT/functional/mcp_tools/load_chain@eFK4v0N5MoJKIzShqVG7Jetec3A -->

# LoadChain

Loads the complete spec chain for a given node and returns
the chain hash, context, and input.

## Interface

```
function LoadChain(logical_name) -> list of text items
  errors:
    - invalid logical name: not a recognized ROOT/ reference.
    - no outputs: target node has no outputs field.
    - invalid output path: an output path fails path validation.
    - chain resolution failure: a dependency cannot be resolved.
    - unreadable file: a file in the chain cannot be read or parsed.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the target node. |

### Output

The result contains separate text items:

| Item | Always present | Content |
|---|---|---|
| Chain hash | yes | The 27-character base64url chain hash. |
| Context | yes | All chain content concatenated as a single stream. |
| Input | only if `input` field exists | Content of the input artifact, excluding frontmatter. |

---

## Algorithm

### Step 1 — Validation

1. The logical name must be a valid `ROOT/` reference.
   Use `ResolvePath(logical_name)` — if it fails, raise
   "invalid logical name".
2. Read the frontmatter of the target node using
   `ParseFrontmatter`. It must have `outputs` declared.
   If `outputs` is empty, raise "no outputs".
3. For each output, call `ValidatePath` on the output path.
   If any fails, raise "invalid output path".

### Step 2 — Chain hash

Call `ComputeChainHash(logical_name)` to compute the
27-character base64url chain hash. This is returned as
the first text item.

### Step 3 — Context stream

The context is a single continuous text block — no
delimiters, no headers, no file boundaries. All files are
read using `OpenFileReader` (close each reader after
reading). Content is concatenated in this exact order:

**3a — Ancestors** (root to target's parent)

For each ancestor, from the root node down to the target's
direct parent, in tree depth order:
- Use `ParseNode` to parse the ancestor.
- Include the `# Public` section — both the direct content
  and all `##` subsections (with their headings). Omit only
  the `# Public` heading itself.
- If `# Public` is absent or has no content and no
  subsections, skip this ancestor entirely.

**3b — Dependencies** (`depends_on`)

For each entry in the target's `depends_on`, in alphabetical
order by logical name:
- `ROOT/x/y` — use `ParseNode` to parse the referenced node.
  Include the full `# Public` section — direct content and
  all `##` subsections (with their headings). Omit the
  `# Public` heading.
- `ROOT/x/y(z)` — use `ParseNode` to parse the referenced
  node. Find the `## z` subsection within `# Public` by
  normalizing headings with `NormalizeName`. Include only
  that subsection's content.
- `ARTIFACT/x/y(id)` — resolve the artifact path using
  `ResolveArtifactReference`, then look up the output path
  in the referenced node's frontmatter. Read the file and
  exclude any frontmatter.

**3c — External files** (`external`)

For each entry in the target's `external`, in alphabetical
order by path:
- If no `fragments` declared — read and include the full
  file content.
- If `fragments` declared — for each fragment, extract the
  declared line range. Concatenate all fragments in
  declaration order.

**3d — Target `# Public`**

Use `ParseNode` to parse the target. Emit a reduced
frontmatter block containing only `outputs`. Then include
the `# Public` section — both the direct content and all
`##` subsections (with their headings). Omit only the
`# Public` heading itself.

**3e — Target `# Agent`**

Include the target node's `# Agent` section — both the
direct content and all `##` subsections (with their
headings). Omit only the `# Agent` heading itself. If the
section is absent, skip.

### Step 4 — Input separation

If the target node has an `input` field, read the referenced
artifact file, exclude any frontmatter, and return its content
as a separate text item. This is not concatenated into the
context stream — it allows the consumer to distinguish context
(what informs) from input (what to transform).

---

## Contracts

- Returns everything in one call — no pagination.
- If any file in the chain is unreadable, returns an error
  (no partial results).
- The context stream contains no metadata or structural
  markers — only spec content.
- The chain hash is computed by `ComputeChainHash`, not
  reimplemented. This guarantees consistency with other tools
  that use the same function (e.g. `validate_specs`).

---

## Dependencies

| Function | Package | Used for |
|---|---|---|
| `ComputeChainHash` | chain_hash | Compute the 27-char chain hash |
| `OpenFileReader`, `ReadLine`, `Close` | file_reader | Read chain files |
| `ParseFrontmatter` | frontmatter | Parse target's frontmatter |
| `ResolvePath`, `GetParent`, `ResolveArtifactReference`, `ExtractQualifier` | logical_names | Resolve logical names to paths |
| `NormalizeName` | name_normalization | Normalize headings for matching |
| `ParseNode` | node_parsing | Parse node body into sections |
| `ValidatePath` | path_validation | Validate output paths |
