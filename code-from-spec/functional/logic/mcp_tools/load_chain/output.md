<!-- code-from-spec: SPEC/functional/logic/mcp_tools/load_chain@vvsydp-6zcj8tMjlG60PVEcHGWg -->

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

  1. Call `LogicalNameToPath(logical_name)` to get the target node's file path.
     If it fails, propagate the error.

  2. Call `FrontmatterParse(target_file_path)` to read the target node's frontmatter.
     If `frontmatter.output` is empty, raise error "NoOutput".
     Call `PathValidateCfs(frontmatter.output)`.
     If it fails, raise error "InvalidOutputPath".

  3. Call `ChainResolve(logical_name)` to get the resolved `Chain`.
     If it fails, propagate the error.

  4. Call `ChainHashCompute(chain)` with the resolved chain.
     If it fails, propagate the error.
     Store the result as `chain_hash`.

  5. Build the context stream:

     Set `context_parts` to an empty list of strings.

     For each `ancestor` in `chain.ancestors` (in order):
       Call `NodeParse(ancestor.unqualified_logical_name)`.
       If `node.public` is absent or `node.public.subsections` is empty, skip.
       Otherwise:
         Build `block` by concatenating all subsections in document order:
           For each subsection in `node.public.subsections`:
             Add the subsection `raw_heading` (trailing whitespace removed).
             Add each line in `subsection.content` with leading blank lines
               after the heading removed and trailing blank lines removed.
             Ensure the block ends with exactly one LF.
           Separate consecutive subsection blocks with exactly one blank line.
         Append `block` to `context_parts`.

     For each `dep` in `chain.dependencies` (in order):
       If `LogicalNameIsArtifact(dep.unqualified_logical_name)` is true:
         Call `FileOpen(dep.file_path, "read", 30000)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Skip the first line that contains "code-from-spec:" (the artifact tag line).
         Include all other lines.
         Call `FileClose`.
         Append the resulting text to `context_parts`.
       Else if `LogicalNameIsExternal(dep.unqualified_logical_name)` is true:
         Call `FileOpen(dep.file_path, "read", 30000)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Call `FileClose`.
         Append the full file content to `context_parts`.
       Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` is true and `dep.qualifier` is absent:
         Call `NodeParse(dep.unqualified_logical_name)`.
         If `node.public` is absent or `node.public.subsections` is empty, skip.
         Otherwise:
           Build `block` by concatenating all subsections in document order
             (same boundary normalization rules as for ancestors).
           Append `block` to `context_parts`.
       Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` is true and `dep.qualifier` is present:
         Call `NodeParse(dep.unqualified_logical_name)`.
         Compute `normalized_qualifier` = `NormalizeText(dep.qualifier)`.
         Find the subsection in `node.public.subsections` whose `heading` equals `normalized_qualifier`.
         If found:
           Build `block` from the subsection `raw_heading` (trailing whitespace removed)
             and its content (leading blank lines removed, trailing blank lines removed,
             ends with exactly one LF).
           Append `block` to `context_parts`.

     For the target node `chain.target`:
       Build a reduced frontmatter block:
         Line 1: "---"
         Line 2: "output: <frontmatter.output>"
         Line 3: "---"
       Append this block to `context_parts`.

       Call `NodeParse(chain.target.unqualified_logical_name)`.
       If `node.public` is present and `node.public.subsections` is non-empty:
         Build `block` by concatenating all subsections in document order
           (same boundary normalization rules as above).
         Append `block` to `context_parts`.

       If `node.agent` is present:
         Build `agent_block`:
           Add `node.agent.raw_heading` (trailing whitespace removed).
           Add each line in `node.agent.content`
             (leading blank lines removed, trailing blank lines removed).
           For each subsection in `node.agent.subsections`:
             Separate from previous block with exactly one blank line.
             Add the subsection `raw_heading` (trailing whitespace removed).
             Add each line in `subsection.content`
               (leading blank lines removed, trailing blank lines removed).
           Ensure the block ends with exactly one LF.
         Append `agent_block` to `context_parts`.

  6. Assemble the output string:

     Start with line: "chain_hash: <chain_hash>"
     Append line: "--- context ---"
     Append the context stream: join all entries in `context_parts`
       separated by exactly one blank line.

     If `chain.input` is present:
       Append line: "--- input ---"
       If `LogicalNameIsArtifact(chain.input.unqualified_logical_name)` is true:
         Call `FileOpen(chain.input.file_path, "read", 30000)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Skip the first line that contains "code-from-spec:".
         Include all other lines.
         Call `FileClose`.
         Append the resulting text.
       Else (EXTERNAL/ or other):
         Call `FileOpen(chain.input.file_path, "read", 30000)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Call `FileClose`.
         Append the full file content.

     If the file at `frontmatter.output` exists and is readable:
       Append line: "--- existing artifact ---"
       Call `FileOpen` with the `PathCfs` of `frontmatter.output` in "read" mode with timeout 30000.
       Read all lines with `FileReadLine` until `EndOfFile`.
       Call `FileClose`.
       Append the full file content.
       If the file does not exist or cannot be read, omit this section silently.

  7. Return the assembled output string.
