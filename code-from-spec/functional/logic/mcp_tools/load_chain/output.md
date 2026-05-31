<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@XlI7kCMRQ6Y1QfeguL0cqDbNxIY -->

# MCPLoadChain

## Records

record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

## Functions

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
  errors:
    - NoOutputs: target node has no outputs field.
    - InvalidOutputPath: an output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.

  ### Step 1 — Validate and resolve

  1. Call LogicalNameToPath(logical_name) to get the target file path.
     If it fails, propagate the error.

  2. Call FrontmatterParse(target_file_path) to get the target frontmatter.
     If frontmatter.outputs is empty, raise error "NoOutputs".

  3. For each output in frontmatter.outputs:
       Call PathValidateCfs(output.path).
       If validation fails, raise error "InvalidOutputPath".

  4. Call ChainResolve(logical_name) to get the resolved Chain.
     If it fails, propagate the error.

  ### Step 2 — Compute chain hash

  5. Call ChainHashCompute(chain) with the resolved Chain.
     If it fails, propagate the error.
     Store result as chain_hash.

  ### Step 3 — Build context stream

  Initialize context as an empty string.

  Content reconstruction rule: when appending lines from any content list
  or file reader, append "\n" after each line, including the last line.

  **Ancestors** — for each item in chain.ancestors (in order):

  6. Call NodeParse(ancestor.logical_name).
     If node.public is absent, skip this ancestor.
     If node.public has empty content and no subsections, skip this ancestor.
     Otherwise:
       Append each line from node.public.content to context (with "\n").
       For each subsection in node.public.subsections:
         Append the subsection's raw_heading line to context (with "\n").
         Append each line from subsection.content to context (with "\n").

  **Dependencies** — for each item in chain.dependencies (in order):

  7. If LogicalNameIsArtifact(dep.logical_name) is true:
       Call FileOpen(dep.file_path) to get a reader.
       Strip frontmatter from the file:
         Read lines using FileReadLine.
         If the first non-blank line is exactly "---":
           Discard lines until the next line that is exactly "---" (inclusive).
         Append all remaining lines to context (with "\n").
       Call FileClose(reader).

     Else if dep.qualifier is absent:
       Call NodeParse(dep.logical_name).
       If node.public is absent, skip.
       Append each line from node.public.content to context (with "\n").
       For each subsection in node.public.subsections:
         Append the subsection's raw_heading line to context (with "\n").
         Append each line from subsection.content to context (with "\n").

     Else (dep.qualifier is present):
       Call NodeParse(dep.logical_name).
       Compute normalized_qualifier = NormalizeText(dep.qualifier).
       Find the subsection in node.public.subsections whose heading equals
       normalized_qualifier.
       If found:
         Append each line from subsection.content to context (with "\n").
       If not found, skip.

  **External** — for each item in chain.external (in order):

  8. Create a PathCfs from the external entry's path field.

     If the entry has no fragments (fragments is absent or empty):
       Call FileOpen(external_path) to get a reader.
       Read all lines using FileReadLine until EndOfFile.
       Append each line to context (with "\n").
       Call FileClose(reader).

     If the entry has fragments:
       For each fragment in the entry's fragments list (in declaration order):
         Parse fragment.lines as "start-end" (two integers separated by "-").
         Call FileOpen(external_path) to get a reader.
         Call FileSkipLines(reader, start - 1) to skip to the start line.
         Read (end - start + 1) lines using FileReadLine.
         Append each line to context (with "\n").
         Call FileClose(reader).

  **Target Public and Target Frontmatter** — using chain.target:

  9. Emit a reduced frontmatter block containing only the outputs field:
       Append "---\n" to context.
       Append "outputs:\n" to context.
       For each output in frontmatter.outputs:
         Append "  - id: <output.id>\n" to context.
         Append "    path: <output.path>\n" to context.
       Append "---\n" to context.

  10. Call NodeParse(chain.target.logical_name).
      If node.public is present and not empty:
        Append each line from node.public.content to context (with "\n").
        For each subsection in node.public.subsections:
          Append the subsection's raw_heading line to context (with "\n").
          Append each line from subsection.content to context (with "\n").

  **Target Agent**:

  11. From the same NodeParse result, if node.agent is present:
        Append each line from node.agent.content to context (with "\n").
        For each subsection in node.agent.subsections:
          Append the subsection's raw_heading line to context (with "\n").
          Append each line from subsection.content to context (with "\n").
      If node.agent is absent, skip.

  ### Step 4 — Extract input

  12. If chain.input is present:
        Call FileOpen(chain.input.file_path) to get a reader.
        Strip frontmatter from the file (same logic as in step 7).
        Read remaining lines and concatenate (with "\n") into input_content.
        Call FileClose(reader).
        Set input = input_content.

      If chain.input is absent:
        input is absent.

  ### Step 5 — Return result

  13. Return MCPLoadChainResult with:
        chain_hash = chain_hash
        context = context
        input = input (absent if chain.input was absent)
