<!-- code-from-spec: ROOT/functional/logic/mcp_tools/load_chain@yR6BFQKNDtyBA58mJMjZz7etK7g -->

function MCPLoadChain(logical_name: string) -> string

  1. Resolve `logical_name` to a file path by calling `LogicalNameToPath(logical_name)`.
     If it fails, propagate the error.

  2. Call `FrontmatterParse(node_path)` to read the target node's frontmatter.
     If `frontmatter.output` is empty, raise error "NoOutput".
     Call `PathValidateCfs(frontmatter.output)`.
     If it fails, raise error "InvalidOutputPath".

  3. Call `ChainResolve(logical_name)` to get the resolved `chain`.
     If it fails, propagate the error.

  4. Call `ChainHashCompute(chain)` with the resolved chain.
     If it fails, propagate the error.
     Store the result as `chain_hash`.

  5. Build the context stream as a single continuous text block,
     concatenating content in chain assembly order:

     For each ancestor in `chain.ancestors`:
       Call `NodeParse(ancestor.logical_name)`.
       If `node.public` is absent or has no subsections, skip.
       Otherwise:
         For each subsection in `node.public.subsections`:
           Append the subsection `raw_heading` + "\n".
           For each line in `subsection.content`, append line + "\n".

     For each dependency in `chain.dependencies`:
       If `LogicalNameIsArtifact(dep.logical_name)` is true:
         Call `FileOpen(dep.file_path)`.
         Read all lines with `FileReadLine` until EndOfFile.
         Remove the artifact tag line: skip the line that contains
           "code-from-spec: <name>@<hash>" (the pattern "code-from-spec:").
         For each remaining line, append line + "\n".
         Call `FileClose`.
       Else if `dep.qualifier` is absent:
         Call `NodeParse(dep.logical_name)`.
         If `node.public` is present and has subsections:
           For each subsection in `node.public.subsections`:
             Append the subsection `raw_heading` + "\n".
             For each line in `subsection.content`, append line + "\n".
       Else:
         Call `NodeParse(dep.logical_name)`.
         Compute `normalized_qualifier` = `NormalizeText(dep.qualifier)`.
         Find the subsection in `node.public.subsections` whose `heading`
           equals `normalized_qualifier`.
         Append the subsection `raw_heading` + "\n".
         For each line in `subsection.content`, append line + "\n".

     For each entry in `chain.external`:
       Create a `PathCfs` from `entry.path`.
       Call `FileOpen(external_path)`.
       Read all lines with `FileReadLine` until EndOfFile.
       For each line, append line + "\n".
       Call `FileClose`.

     For the target in `chain.target`:
       Emit a reduced frontmatter block:
         Append "---\n".
         Append "output: " + frontmatter.output + "\n".
         Append "---\n".
       Call `NodeParse(chain.target.logical_name)`.
       If `node.public` is present and has subsections:
         For each subsection in `node.public.subsections`:
           Append the subsection `raw_heading` + "\n".
           For each line in `subsection.content`, append line + "\n".
       If `node.agent` is present:
         Append the `# Agent` raw heading + "\n".
         For each line in `node.agent.content`, append line + "\n".
         For each subsection in `node.agent.subsections`:
           Append the subsection `raw_heading` + "\n".
           For each line in `subsection.content`, append line + "\n".

  6. Assemble the output string:

     Start with: "chain_hash: " + chain_hash + "\n".
     Append: "--- context ---\n".
     Append: the full context stream from step 5.

     If `chain.input` is present:
       Append: "--- input ---\n".
       Call `FileOpen(chain.input.file_path)`.
       Read all lines with `FileReadLine` until EndOfFile.
       Remove the artifact tag line: skip the line that contains
         "code-from-spec: <name>@<hash>" (the pattern "code-from-spec:").
       For each remaining line, append line + "\n".
       Call `FileClose`.

     Attempt to open the output file at `frontmatter.output` with `FileOpen`.
     If it opens successfully:
       Append: "--- existing artifact ---\n".
       Read all lines with `FileReadLine` until EndOfFile.
       For each line, append line + "\n".
       Call `FileClose`.
     If the file does not exist or cannot be read, skip silently.

  7. Return the assembled string.

  errors:
    - NoOutput: target node has no output field.
    - InvalidOutputPath: the output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.
