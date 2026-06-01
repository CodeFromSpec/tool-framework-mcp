<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@wFQ-I-5jYXxBtNjsCjCydFWBSNM -->

## Namespace

    namespace: mcploadchain

## Records

```
record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string
```

## Functions

---

### MCPLoadChain

```
function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
  errors:
    - NoOutputs: target node has no outputs field.
    - InvalidOutputPath: an output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.
```

**Step 1 — Validate and resolve**

1. Call `LogicalNameToPath(logical_name)` to get the target node's file path.
   If it fails, propagate the error.

2. Call `FrontmatterParse(file_path)` on the resolved path.
   If `frontmatter.outputs` is empty, raise error "NoOutputs".

3. For each output in `frontmatter.outputs`:
   Call `PathValidateCfs(output.path)`.
   If it fails, raise error "InvalidOutputPath".

4. Call `ChainResolve(logical_name)` to get the resolved `Chain`.
   If it fails, propagate the error.

**Step 2 — Compute chain hash**

5. Call `ChainHashCompute(chain)` with the resolved `Chain`.
   If it fails, propagate the error.
   Store the result as `chain_hash`.

**Step 3 — Build context stream**

6. Initialize `context` as an empty string.

7. **Ancestors** — for each item in `chain.ancestors` (in order):

   a. Call `NodeParse(ancestor.logical_name)`.

   b. If `node.public` is absent, skip this ancestor.
      If `node.public` has empty `content` and an empty `subsections` list, skip.

   c. Otherwise, append the `node.public.raw_heading` followed by `\n`.
      For each line in `node.public.content`, append the line followed by `\n`.
      For each subsection in `node.public.subsections` (in order):
        Append the subsection's `raw_heading` followed by `\n`.
        For each line in the subsection's `content`, append the line followed by `\n`.

8. **Dependencies** — for each item in `chain.dependencies` (in order):

   a. If `LogicalNameIsArtifact(dep.logical_name)` is true:
      Open the file at `dep.file_path` with `FileOpen`.
      Strip frontmatter from the file (see "Frontmatter stripping" below).
      Read all remaining lines; for each line append the line followed by `\n`.
      Call `FileClose`.

   b. Else if `dep.qualifier` is absent:
      Call `NodeParse(dep.logical_name)`.
      Append `node.public.raw_heading` followed by `\n`.
      For each line in `node.public.content`, append the line followed by `\n`.
      For each subsection in `node.public.subsections` (in order):
        Append the subsection's `raw_heading` followed by `\n`.
        For each line in the subsection's `content`, append the line followed by `\n`.

   c. Else (`dep.qualifier` is present):
      Call `NodeParse(dep.logical_name)`.
      Compute `normalized_qualifier` = `NormalizeText(dep.qualifier)`.
      Find the subsection in `node.public.subsections` whose `heading` equals `normalized_qualifier`.
      Append that subsection's `raw_heading` followed by `\n`.
      For each line in the subsection's `content`, append the line followed by `\n`.

9. **External** — for each item in `chain.external` (in order):

   a. Create a `PathCfs` from `external.path`.

   b. If `external.fragments` is absent or empty:
      Open the file with `FileOpen`.
      Read all lines; for each line append the line followed by `\n`.
      Call `FileClose`.

   c. Else (fragments present):
      For each fragment in `external.fragments` (in declaration order):
        Parse `fragment.lines` as "start-end" to get integer `start` and `end`.
        Open the file at the external path with `FileOpen`.
        Call `FileSkipLines(reader, start - 1)` to skip preceding lines.
        Read `end - start + 1` lines; for each line append the line followed by `\n`.
        Call `FileClose`.

10. **Target Public** — using `chain.target`:

    a. Emit a reduced frontmatter block for the target.
       Append `"---\n"`.
       Append `"outputs:\n"`.
       For each output in `frontmatter.outputs`:
         Append `"  - id: "` + `output.id` + `"\n"`.
         Append `"    path: "` + `output.path` + `"\n"`.
       Append `"---\n"`.

    b. Call `NodeParse(chain.target.logical_name)`.

    c. If `node.public` is not absent:
       Append `node.public.raw_heading` followed by `\n`.
       For each line in `node.public.content`, append the line followed by `\n`.
       For each subsection in `node.public.subsections` (in order):
         Append the subsection's `raw_heading` followed by `\n`.
         For each line in the subsection's `content`, append the line followed by `\n`.

11. **Target Agent** — from the same `NodeParse` result:

    a. If `node.agent` is absent, skip.

    b. Otherwise, append `node.agent.raw_heading` followed by `\n`.
       For each line in `node.agent.content`, append the line followed by `\n`.
       For each subsection in `node.agent.subsections` (in order):
         Append the subsection's `raw_heading` followed by `\n`.
         For each line in the subsection's `content`, append the line followed by `\n`.

**Step 4 — Extract input**

12. If `chain.input` is present:
    Open the file at `chain.input.file_path` with `FileOpen`.
    Strip frontmatter from the file (see "Frontmatter stripping" below).
    Read all remaining lines; for each line append the line followed by `\n`.
    Store the result as `input`.
    Call `FileClose`.

    If `chain.input` is absent, `input` is absent in the result.

**Step 5 — Return result**

13. Return `MCPLoadChainResult` with:
    - `chain_hash`: the value computed in step 5.
    - `context`: the concatenated stream built in steps 7–11.
    - `input`: the value from step 12, or absent if `chain.input` was absent.

---

### Frontmatter stripping (helper procedure)

Used when reading artifact files and the input file.

```
procedure StripFrontmatter(reader: FileReader)
```

1. Read the first line from `reader`.
   If it equals `"---"`, frontmatter is present — continue to step 2.
   Otherwise, the file has no frontmatter.
   The line already read is content — include it in the output.
   Continue reading the remaining lines normally.

2. Read lines until a line equal to `"---"` is encountered (the closing delimiter).
   Discard all lines including the closing `"---"`.
   All subsequent lines are content.
