---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/load_chain
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/utils/text_normalization
input: ARTIFACT/functional/logic/mcp_tools/load_chain
output: code-from-spec/golang/implementation/mcp_tools/load_chain_debug/chain.md
---

# SPEC/golang/implementation/mcp_tools/load_chain_debug

Diagnostic node. Saves the complete chain content
received by the generation subagent to a file for
inspection.

# Agent

Save the entire content you received from `load_chain`
to the output file, verbatim. Place the artifact tag
as the first line, inside a markdown comment.
