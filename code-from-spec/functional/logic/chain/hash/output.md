<!-- code-from-spec: ROOT/functional/logic/chain/hash@bQ0mPHWFL9qw5U3rmzbu3zHNyTw -->

function ChainHashCompute(chain: chainresolver.Chain) -> string
  errors:
    - FileUnreadable: a file in the chain cannot be read or opened.
    - ParseFailure: a node file cannot be parsed.
    - (FileReader.*): propagated from FileOpen.
    - (NodeParsing.*): propagated from NodeParse.

  1. Initialize an empty list: content_hashes.

  2. For each ancestor in chain.ancestors (root-first order):
       Call NodeParse with ancestor.logical_name.
       If NodeParse fails, raise error "parse failure".
       Hash the node's # Public section (subsections only — see HashPublicSubsections).
       If the section is absent or produces no bytes, skip.
       Else append the resulting 20-byte SHA-1 to content_hashes.

  3. For each dependency in chain.dependencies:
       If LogicalNameIsArtifact(dep.logical_name) is true:
         Hash the artifact file at dep.file_path (frontmatter stripped — see HashArtifactFile).
         Append the resulting 20-byte SHA-1 to content_hashes.
       Else if dep.qualifier is absent:
         Call NodeParse with dep.logical_name.
         If NodeParse fails, raise error "parse failure".
         Hash the node's # Public section (subsections only — see HashPublicSubsections).
         If the section is absent or produces no bytes, skip.
         Else append the resulting 20-byte SHA-1 to content_hashes.
       Else (dep.qualifier is present):
         Call NodeParse with dep.logical_name.
         If NodeParse fails, raise error "parse failure".
         Find the subsection within node.public whose heading equals
           NormalizeText(dep.qualifier).
         If not found, skip.
         Else hash the subsection (see HashSubsection).
         Append the resulting 20-byte SHA-1 to content_hashes.

  4. For each external entry in chain.external:
       Hash the external file at external.path (see HashExternalFile).
       Append the resulting 20-byte SHA-1 to content_hashes.

  5. Call NodeParse with chain.target.logical_name.
     If NodeParse fails, raise error "parse failure".

     Hash the target node's # Public section (subsections only — see HashPublicSubsections).
     If the section is present and produces bytes, append 20-byte SHA-1 to content_hashes.

     Hash the target node's # Agent section (see HashAgentSection).
     If the section is present and produces bytes, append 20-byte SHA-1 to content_hashes.

  6. If chain.input is present:
       Hash the artifact file at chain.input.file_path (frontmatter stripped — see HashArtifactFile).
       Append the resulting 20-byte SHA-1 to content_hashes.

  7. Concatenate all 20-byte entries in content_hashes as raw bytes.
     Compute SHA-1 of the concatenation.
     Encode the 20-byte result as base64url (RFC 4648 §5, no padding).
     Return the resulting 27-character string.


--- Helper: Hash public subsections ---

HashPublicSubsections(node: parsenode.Node) -> optional 20-byte hash

  Used for ancestors, the target's # Public, and ROOT/ dependencies.
  The # Public heading itself and any content directly under # Public
  (before the first ## subsection) are NOT included.

  1. If node.public is absent, return absent.

  2. If node.public.subsections is empty, return absent.

  3. Initialize byte accumulator.

  4. For each subsection in node.public.subsections (document order):
       Append subsection.raw_heading + "\n" to accumulator.
       For each line in subsection.content:
         Append line + "\n" to accumulator.

  5. Compute SHA-1 of the accumulated bytes.
     Return the 20-byte result.


--- Helper: Hash agent section ---

HashAgentSection(node: parsenode.Node) -> optional 20-byte hash

  Used for the target's # Agent section only.

  1. If node.agent is absent, return absent.

  2. Initialize byte accumulator.

  3. Append node.agent.raw_heading + "\n" to accumulator.

  4. For each line in node.agent.content:
       Append line + "\n" to accumulator.

  5. For each subsection in node.agent.subsections:
       Append subsection.raw_heading + "\n" to accumulator.
       For each line in subsection.content:
         Append line + "\n" to accumulator.

  6. If accumulator contains only the raw_heading line (no content, no subsections),
     that is still hashed — do not skip.

  7. Compute SHA-1 of the accumulated bytes.
     Return the 20-byte result.


--- Helper: Hash subsection ---

HashSubsection(subsection: parsenode.NodeSubsection) -> 20-byte hash

  1. Initialize byte accumulator.

  2. Append subsection.raw_heading + "\n" to accumulator.

  3. For each line in subsection.content:
       Append line + "\n" to accumulator.

  4. Compute SHA-1 of the accumulated bytes.
     Return the 20-byte result.


--- Helper: Hash artifact file (frontmatter stripped) ---

HashArtifactFile(file_path: pathutils.PathCfs) -> 20-byte hash

  1. Call FileOpen with file_path.
     If FileOpen fails, raise error "file unreadable".

  2. Call FileReadLine to read the first line.
     If EndOfFile, the file is empty:
       Call FileClose.
       Compute SHA-1 of empty bytes.
       Return the 20-byte result.

  3. If the first line is exactly "---":
       Read lines until a line that is exactly "---" is encountered
         (this closes the frontmatter block). Discard these lines.
       Read all remaining lines into a list.
     Else:
       The first line is content. Collect it plus all remaining lines into a list.

  4. Initialize byte accumulator.
     For each collected line:
       Apply NeutralizeArtifactTag to the line.
       Append the result + "\n" to accumulator.

  5. Call FileClose.

  6. Compute SHA-1 of the accumulated bytes.
     Return the 20-byte result.

  On any error path: call FileClose before raising.


--- Helper: Hash external file ---

HashExternalFile(path_string: string) -> 20-byte hash

  1. Create a PathCfs from path_string.
     Call FileOpen with that PathCfs.
     If FileOpen fails, raise error "file unreadable".

  2. Read all lines with FileReadLine until EndOfFile.
     For each line, append line + "\n" to byte accumulator.

  3. Call FileClose.

  4. Compute SHA-1 of the accumulated bytes.
     Return the 20-byte result.

  On any error path: call FileClose before raising.


--- Helper: Artifact tag hash neutralization ---

NeutralizeArtifactTag(line: string) -> string

  1. Search the line for the pattern:
       "code-from-spec: " followed by a logical name, then "@",
       then exactly 27 base64url characters.

  2. If the pattern is found, replace the 27-character hash portion
     with 27 hyphens ("---------------------------").
     Return the modified line.

  3. If the pattern is not found, return the line unchanged.
