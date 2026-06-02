---
depends_on:
  - ARTIFACT/golang/interfaces/chain/hash
  - ARTIFACT/golang/interfaces/chain/resolver
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
  - ARTIFACT/golang/interfaces/utils/logical_names
input: ARTIFACT/functional/tests/chain/hash
output: internal/chainhash/chainhash_test.go
---

# ROOT/golang/tests/chain/hash

# Agent

## Test setup guidance

`ChainHashCompute` calls `NodeParse` internally for spec
node positions (ancestors, target, ROOT/ dependencies).
`NodeParse` requires a valid `ROOT/` logical name that
resolves to a `_node.md` file on disk.

Therefore, tests that reference spec nodes must:
1. Use `testChdir` to set the working directory.
2. Create `code-from-spec/.../_node.md` files on disk
   matching the logical names used in ChainItems.
3. Set `ChainItem.LogicalName` to a valid `ROOT/`
   logical name (e.g. `"ROOT/a"`), not a file path.
4. Set `ChainItem.FilePath` to the corresponding
   `PathCfs` (e.g. `{Value: "code-from-spec/a/_node.md"}`).

For ARTIFACT/ items, `ChainItem.LogicalName` must start
with `"ARTIFACT/"` so the implementation reads the file
directly instead of calling `NodeParse`.

Node files on disk must have valid structure for
`NodeParse`: at minimum a `# <logical_name>` heading
as the first heading.
