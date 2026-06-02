<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@8_kiyHf8lL8SJOPHcVE3G0VwFd0 -->

# Public

## Namespace

    namespace: mcploadchain

## Interface

```
record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
  errors:
    - NoOutput: target node has no output field.
    - InvalidOutputPath: the output path fails path
      validation.
    - (LogicalNames.*): propagated from
      LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.
```

MCPLoadChain validates the target, resolves the chain,
computes the hash, and assembles the full context stream.

### Step 1 — Validate and resolve

1. Call LogicalNameToPath(logical_name).
   If it raises an error, propagate it.

2. Call FrontmatterParse with the resolved file path.
   If it raises an error, propagate it.

3. If frontmatter.output is empty, raise error NoOutput.

4. Call PathValidateCfs(frontmatter.output).
   If it raises an error, raise InvalidOutputPath.

5. Call ChainResolve(logical_name).
   If it raises an error, propagate it.
   Store the result as chain.

### Step 2 — Compute chain hash

6. Call ChainHashCompute(chain).
   If it raises an error, propagate it.
   Store the result as chain_hash.

### Step 3 — Build context stream

The context is a single continuous text block with no
delimiters or file boundaries. Content is appended in
chain assembly order. When reconstructing content from
lines, append "\n" after each line, including the last.

**Ancestors** (from chain.ancestors)

7. For each ancestor in chain.ancestors:
   a. Call NodeParse(ancestor.logical_name).
   b. If node.public is absent, skip this ancestor.
   c. If node.public has empty content and no
      subsections, skip this ancestor.
   d. Otherwise, append node.public.raw_heading + "\n".
   e. For each line in node.public.content,
      append line + "\n".
   f. For each subsection in node.public.subsections:
      append subsection.raw_heading + "\n".
      For each line in subsection.content,
        append line + "\n".

**Dependencies** (from chain.dependencies)

8. For each dep in chain.dependencies:
   a. If LogicalNameIsArtifact(dep.logical_name):
      Open the file at dep.file_path with FileOpen.
      Strip frontmatter if present (skip lines from
      the first "---" to the closing "---" inclusive).
      Append remaining lines, each followed by "\n".
      Call FileClose.
   b. Else if dep.qualifier is absent:
      Call NodeParse(dep.logical_name).
      Append node.public.raw_heading + "\n".
      For each line in node.public.content,
        append line + "\n".
      For each subsection in node.public.subsections:
        append subsection.raw_heading + "\n".
        For each line in subsection.content,
          append line + "\n".
   c. Else (dep.qualifier is present):
      Call NodeParse(dep.logical_name).
      Find the subsection in node.public.subsections
      whose heading matches
      NormalizeText(dep.qualifier).
      Append subsection.raw_heading + "\n".
      For each line in subsection.content,
        append line + "\n".

**External** (from chain.external)

9. For each external entry in chain.external:
   Create a PathCfs from entry.path.
   Open with FileOpen.
   For each line read with FileReadLine until
   EndOfFile, append line + "\n".
   Call FileClose.

**Target Public** (from chain.target)

10. Emit a reduced frontmatter block containing only
    the output field:
    Append "---\n".
    Append "output: " + frontmatter.output + "\n".
    Append "---\n".

11. Call NodeParse(chain.target.logical_name).
    Append node.public.raw_heading + "\n".
    For each line in node.public.content,
      append line + "\n".
    For each subsection in node.public.subsections:
      append subsection.raw_heading + "\n".
      For each line in subsection.content,
        append line + "\n".

**Target Agent**

12. From the same NodeParse result:
    If node.agent is absent, skip.
    Otherwise, append node.agent.raw_heading + "\n".
    For each line in node.agent.content,
      append line + "\n".
    For each subsection in node.agent.subsections:
      append subsection.raw_heading + "\n".
      For each line in subsection.content,
        append line + "\n".

### Step 4 — Extract input

13. If chain.input is absent, input is absent in the result.

14. If chain.input is present:
    Open the file at chain.input.file_path with FileOpen.
    Strip frontmatter if present (skip lines from the
    first "---" to the closing "---" inclusive).
    Collect remaining lines, each followed by "\n",
    into the input string.
    Call FileClose.

### Step 5 — Return result

15. Return MCPLoadChainResult with:
    - chain_hash: the value from Step 2.
    - context: the concatenated stream from Step 3.
    - input: the value from Step 4, or absent.
