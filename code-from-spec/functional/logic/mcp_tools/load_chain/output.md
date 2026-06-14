<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@3R8cIcCjMy-UfBV_1ASXkrdVnB0 -->

namespace: mcploadchain

function MCPLoadChain(logical_name: string) -> string
  errors:
    - NoOutput: target node has no output field.
    - InvalidOutputPath: the output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.

  1. Call LogicalNameToPath(logical_name).
     If it fails, propagate the error.

  2. Call FrontmatterParse on the resolved path.
     If frontmatter.output is empty, raise error "NoOutput".
     Call PathValidateCfs(frontmatter.output).
     If it fails, raise error "InvalidOutputPath".

  3. Call ChainResolve(logical_name) to get chain.
     If it fails, propagate the error.

  4. Call ChainHashCompute(chain) to get chain_hash.
     If it fails, propagate the error.

  5. Build the context stream as a single text block by
     processing chain entries in assembly order.
     Maintain a list of rendered blocks; each block is a
     string ending with exactly one LF.
     When appending blocks, separate them with exactly one
     blank line.

     Block extraction rules (applied to each spec node section
     or subsection):
       - Take the raw_heading line with trailing whitespace removed.
       - Take the content lines.
       - Remove all leading blank lines from the content.
       - Remove all trailing blank lines from the content.
       - Combine: heading line + LF + content lines (each
         followed by LF) + exactly one trailing LF.
       When content is empty, the block is just the heading
       line + LF.
       When concatenating multiple subsection blocks within
       one entry, separate them with exactly one blank line.

     5a. Ancestors (chain.ancestors, in order)

       For each ancestor in chain.ancestors:
         Call NodeParse(ancestor.unqualified_logical_name).
         If node.public is absent, skip this ancestor.
         If node.public has no subsections, skip this ancestor.
         Otherwise:
           Collect each subsection in node.public.subsections
           (document order), applying block extraction to each.
           Concatenate the subsection blocks, separated by
           exactly one blank line.
           Append the resulting block to the context stream.

     5b. Dependencies (chain.dependencies, in order)

       For each dep in chain.dependencies:

         If LogicalNameIsArtifact(dep.unqualified_logical_name):
           Call FileOpen(dep.file_path).
           Read all lines with FileReadLine until EndOfFile.
           Collect lines, removing the line that contains
           "code-from-spec: <name>@<hash>" (the artifact tag line).
           Call FileClose.
           Join collected lines (each followed by LF), ensure
           content ends with exactly one LF.
           Append to context stream.

         Else if LogicalNameIsExternal(dep.unqualified_logical_name):
           Call FileOpen(dep.file_path).
           Read all lines with FileReadLine until EndOfFile.
           Call FileClose.
           Join collected lines (each followed by LF), ensure
           content ends with exactly one LF.
           Append to context stream.

         Else if LogicalNameIsSpec(dep.unqualified_logical_name)
         and dep.qualifier is absent:
           Call NodeParse(dep.unqualified_logical_name).
           If node.public is absent or has no subsections, skip.
           Otherwise:
             Collect each subsection in node.public.subsections
             (document order), applying block extraction to each.
             Concatenate, separated by exactly one blank line.
             Append to context stream.

         Else if LogicalNameIsSpec(dep.unqualified_logical_name)
         and dep.qualifier is present:
           Call NodeParse(dep.unqualified_logical_name).
           Compute target_heading = NormalizeText(dep.qualifier).
           Find the subsection in node.public.subsections whose
           heading equals target_heading.
           If not found, skip.
           Otherwise:
             Apply block extraction to the subsection.
             Append to context stream.

     5c. Target Public and Target Agent (chain.target)

       Build a reduced frontmatter block:
         Lines: "---" + LF + "output: <frontmatter.output>" + LF + "---"
         followed by LF.
         This is one block; append to context stream.

       Call NodeParse(chain.target.unqualified_logical_name).

       If node.public is present and has subsections:
         Collect each subsection in node.public.subsections
         (document order), applying block extraction to each.
         Concatenate, separated by exactly one blank line.
         Append to context stream.

       If node.agent is present:
         Build the agent block:
           Start with node.agent.raw_heading with trailing
           whitespace removed, followed by LF.
           Take node.agent.content lines (leading blank lines
           removed, trailing blank lines removed), each
           followed by LF.
           For each subsection in node.agent.subsections
           (document order):
             Separate from previous block with exactly one
             blank line.
             Apply block extraction to the subsection and append.
           Ensure the block ends with exactly one LF.
         Append to context stream.

  6. Assemble the final output string:

     Start with:
       "chain_hash: <chain_hash>" + LF

     Append:
       "--- context ---" + LF

     Append the context stream (already ends with one LF).

     If chain.input is present:
       Append "--- input ---" + LF.
       If LogicalNameIsArtifact(chain.input.unqualified_logical_name):
         Call FileOpen(chain.input.file_path).
         Read all lines until EndOfFile.
         Remove the artifact tag line.
         Call FileClose.
         Append lines (each followed by LF).
       Else (EXTERNAL/ input):
         Call FileOpen(chain.input.file_path).
         Read all lines until EndOfFile.
         Call FileClose.
         Append lines (each followed by LF).

     If the output file at frontmatter.output exists and
     is readable:
       Call FileOpen(PathCfs of frontmatter.output).
       If FileOpen succeeds:
         Read all lines until EndOfFile.
         Call FileClose.
         Append "--- existing artifact ---" + LF.
         Append lines (each followed by LF).
       If FileOpen fails (file absent or unreadable):
         Silently omit the section; do not raise an error.

  7. Return the assembled string.
