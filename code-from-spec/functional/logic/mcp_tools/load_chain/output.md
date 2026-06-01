<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@oHptTnaGu9f022999JXRQgGF98c -->

namespace: mcploadchain

record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult
  errors:
    - NoOutputs: target node has no outputs field.
    - InvalidOutputPath: an output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.

  1. Call LogicalNameToPath(logical_name) to get the target file path.
     If it fails, propagate the error.

  2. Call FrontmatterParse(target_file_path) to read the target's frontmatter.
     If frontmatter.outputs is empty, raise error "NoOutputs".
     For each output in frontmatter.outputs, call PathValidateCfs(output.path).
     If any fails, raise error "InvalidOutputPath".

  3. Call ChainResolve(logical_name) to get the resolved chain.
     If it fails, propagate the error.

  4. Call ChainHashCompute(chain) with the resolved chain.
     If it fails, propagate the error.
     Store the result as chain_hash.

  5. Build the context stream by concatenating content in chain assembly order.
     For each line appended, add "\n" after it, including the last line.

     5a. Ancestors — for each entry in chain.ancestors:
           Call NodeParse(ancestor.logical_name).
           If node.public is absent, skip this ancestor.
           If node.public has empty content and no subsections, skip this ancestor.
           Otherwise:
             Append node.public.raw_heading followed by "\n".
             For each line in node.public.content, append the line followed by "\n".
             For each subsection in node.public.subsections:
               Append subsection.raw_heading followed by "\n".
               For each line in subsection.content, append the line followed by "\n".

     5b. Dependencies — for each entry in chain.dependencies:
           If LogicalNameIsArtifact(dep.logical_name) is true:
             Call FileOpen(dep.file_path) to open the file.
             Strip frontmatter from the file:
               Read lines until a line is exactly "---" — that is the frontmatter start.
               If the first non-blank line is "---", read and discard lines until
               the closing "---" line, then discard the closing "---" line as well.
               If no opening "---" is found, treat the file as having no frontmatter
               and include all content from the beginning.
             Append remaining lines, each followed by "\n".
             Call FileClose on the reader.
           Else if dep.qualifier is absent:
             Call NodeParse(dep.logical_name).
             Append node.public.raw_heading followed by "\n".
             For each line in node.public.content, append the line followed by "\n".
             For each subsection in node.public.subsections:
               Append subsection.raw_heading followed by "\n".
               For each line in subsection.content, append the line followed by "\n".
           Else:
             Call NodeParse(dep.logical_name).
             Call NormalizeText(dep.qualifier) to get the normalized qualifier.
             Find the subsection in node.public.subsections whose heading matches
             the normalized qualifier.
             Append that subsection's raw_heading followed by "\n".
             For each line in that subsection's content, append the line followed by "\n".

     5c. External — for each entry in chain.external:
           Create a PathCfs from entry.path.
           Call FileOpen(external_path) to open the file.
           Read all lines; append each line followed by "\n".
           Call FileClose on the reader.

     5d. Target Public and Agent:
           Emit a reduced frontmatter block containing only the outputs field:
             Append "---\n".
             Append "outputs:\n".
             For each output in frontmatter.outputs:
               Append "  - id: <output.id>\n".
               Append "    path: <output.path>\n".
             Append "---\n".
           Call NodeParse(chain.target.logical_name).
           If node.public is present:
             Append node.public.raw_heading followed by "\n".
             For each line in node.public.content, append the line followed by "\n".
             For each subsection in node.public.subsections:
               Append subsection.raw_heading followed by "\n".
               For each line in subsection.content, append the line followed by "\n".
           If node.agent is present:
             Append node.agent.raw_heading followed by "\n".
             For each line in node.agent.content, append the line followed by "\n".
             For each subsection in node.agent.subsections:
               Append subsection.raw_heading followed by "\n".
               For each line in subsection.content, append the line followed by "\n".

  6. Extract input:
     If chain.input is present:
       Call FileOpen(chain.input.file_path) to open the file.
       Strip frontmatter using the same procedure as in step 5b.
       Read remaining lines, appending each followed by "\n".
       Call FileClose on the reader.
       Store the accumulated content as the input field.
     If chain.input is absent, leave input absent.

  7. Return MCPLoadChainResult with:
       chain_hash: the hash computed in step 4.
       context: the concatenated stream built in step 5.
       input: the content extracted in step 6, or absent.
