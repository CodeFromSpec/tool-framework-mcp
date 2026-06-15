<!-- code-from-spec: SPEC/functional/logic/mcp_tools/chain_hash@o-Z7gRAHXVs9Yw9lZc2IzZgKz_I -->

namespace: mcpchainhash

function MCPChainHash(logical_name: string) -> string
  errors:
    - NoOutput: target node has no output field.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (Frontmatter.*): propagated from FrontmatterParse.
    - (FileReader.*): propagated from FileOpen.

  1. Call LogicalNameToPath(logical_name) to get the target node's file path.
     If it fails, propagate the error.

  2. Call FrontmatterParse(file_path) to read the target node's frontmatter.
     If it fails, propagate the error.
     If frontmatter.output is empty, raise error "NoOutput".

  3. Call ChainResolve(logical_name) to get the resolved Chain.
     If it fails, propagate the error.

  4. Call ChainHashCompute(chain) with the resolved Chain.
     If it fails, propagate the error.

  5. Return the 27-character base64url hash string.
