<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@Zg_keutwekoDE3A9bagZzVyfRts -->

# MCPLoadChain

## Records

record MCPLoadChainResult
  chain_hash: string
  context: string
  input: optional string

## Functions

---

function MCPLoadChain(logical_name: string) -> MCPLoadChainResult

  errors:
    - NoOutputs: target node declares no outputs.
    - InvalidOutputPath: an output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.

  ### Step 1 — Validate and resolve

  1. Call LogicalNameToPath(logical_name) to get the target node's file path.
     If it fails, propagate the error.

  2. Call FrontmatterParse(target_file_path).
     If frontmatter.outputs is empty, raise error "NoOutputs".
     For each output entry in frontmatter.outputs:
       Call PathValidateCfs(output.path).
       If it fails, raise error "InvalidOutputPath".

  3. Call ChainResolve(logical_name) to get the resolved Chain.
     If it fails, propagate the error.

  ### Step 2 — Compute chain hash

  4. Call ChainHashCompute(chain).
     If it fails, propagate the error.
     Store the returned 27-character string as chain_hash.

  ### Step 3 — Build context stream

  5. Initialize context as an empty string.

  6. For each ancestor in chain.ancestors (from root down, excluding target):
       Call NodeParse(ancestor.logical_name).
       If node.public is absent, skip this ancestor.
       If node.public has empty content and no subsections, skip this ancestor.
       Otherwise:
         For each line in node.public.content:
           Append the line followed by "\n" to context.
         For each subsection in node.public.subsections:
           Append subsection.raw_heading followed by "\n" to context.
           For each line in subsection.content:
             Append the line followed by "\n" to context.

  7. For each dependency in chain.dependencies:
       If LogicalNameIsArtifact(dep.logical_name) is true:
         Call FileOpen(dep.file_path) to get a reader.
         Strip frontmatter:
           Read lines until a line that is exactly "---" is found; discard it.
           Read and discard lines until a second line that is exactly "---" is found; discard it.
           If no opening "---" is found on the first non-blank line, do not skip any lines
             (treat file as having no frontmatter — start from the beginning).
         Read all remaining lines from the reader, appending each line followed by "\n" to context.
         Call FileClose(reader).
       Else if dep.qualifier is absent:
         Call NodeParse(dep.logical_name).
         If node.public is absent, skip.
         For each line in node.public.content:
           Append the line followed by "\n" to context.
         For each subsection in node.public.subsections:
           Append subsection.raw_heading followed by "\n" to context.
           For each line in subsection.content:
             Append the line followed by "\n" to context.
       Else (dep.qualifier is present):
         Call NodeParse(dep.logical_name).
         Compute normalized_qualifier = NormalizeText(dep.qualifier).
         Find the subsection in node.public.subsections whose heading equals normalized_qualifier.
         If not found, skip.
         For each line in that subsection's content:
           Append the line followed by "\n" to context.

  8. For each external entry in chain.external:
       Create a PathCfs from external_entry.path.
       If external_entry.fragments is absent or empty:
         Call FileOpen(path) to get a reader.
         Read all lines from the reader, appending each line followed by "\n" to context.
         Call FileClose(reader).
       Else (fragments are present):
         For each fragment in external_entry.fragments (in declaration order):
           Parse fragment.lines as "start-end" to get integer start and integer end.
           Call FileOpen(path) to get a reader.
           Call FileSkipLines(reader, start - 1).
           Read (end - start + 1) lines from the reader, appending each followed by "\n" to context.
           Call FileClose(reader).

  9. For the target (chain.target):
       Read the target node's frontmatter outputs using FrontmatterParse(target_file_path).
       Emit a reduced frontmatter block containing only the outputs field:
         Append "---\n" to context.
         Append "outputs:\n" to context.
         For each output entry in frontmatter.outputs:
           Append "  - id: <output.id>\n" to context.
           Append "    path: <output.path>\n" to context.
         Append "---\n" to context.
       Call NodeParse(chain.target.logical_name).
       If node.public is present:
         For each line in node.public.content:
           Append the line followed by "\n" to context.
         For each subsection in node.public.subsections:
           Append subsection.raw_heading followed by "\n" to context.
           For each line in subsection.content:
             Append the line followed by "\n" to context.
       If node.agent is present:
         For each line in node.agent.content:
           Append the line followed by "\n" to context.
         For each subsection in node.agent.subsections:
           Append subsection.raw_heading followed by "\n" to context.
           For each line in subsection.content:
             Append the line followed by "\n" to context.

  ### Step 4 — Extract input

  10. If chain.input is present:
        Call FileOpen(chain.input.file_path) to get a reader.
        Strip frontmatter using the same logic as step 7.
        Read all remaining lines, collecting them into a single string,
          appending each line followed by "\n".
        Call FileClose(reader).
        Store the collected string as input.
      Else:
        input is absent.

  ### Step 5 — Return result

  11. Return MCPLoadChainResult with:
        chain_hash = chain_hash
        context = context
        input = input (absent if chain.input was absent)
