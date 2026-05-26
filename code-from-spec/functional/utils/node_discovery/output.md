<!-- code-from-spec: ROOT/functional/utils/node_discovery@nFE7ic64qvmsWenz5dvUFCOfhaQ -->

# Node Discovery

## Records

```
record DiscoveredNode
  logical_name: string   -- the ROOT/ logical name derived from the file path
  file_path: string      -- the relative path from the project root, e.g. "code-from-spec/x/y/_node.md"
```

## Functions

```
function DiscoverNodes() -> list of DiscoveredNode

  -- Starts from code-from-spec/ relative to the working directory (project root).
  -- No parameters are accepted.

  1. Check that the directory "code-from-spec/" exists under the current working directory.
     If the directory does not exist, raise error "directory not found: code-from-spec/ does not exist".

  2. Walk the "code-from-spec/" directory tree recursively, visiting every file and
     subdirectory at any depth.
     If the filesystem walk fails for any reason (permissions, I/O error, etc.),
     raise error "walk error: filesystem error while traversing".

  3. For each file encountered during the walk:
     If the file's name is not exactly "_node.md", skip it and continue.

  4. For each file whose name is "_node.md":
     a. Record the file's path relative to the project root
        (e.g. "code-from-spec/x/y/_node.md").
     b. Derive the logical name from that relative path using ReverseResolve
        (defined in ROOT/functional/utils/logical_names):
          - "code-from-spec/_node.md"       -> "ROOT"
          - "code-from-spec/x/_node.md"     -> "ROOT/x"
          - "code-from-spec/x/y/_node.md"   -> "ROOT/x/y"
        If ReverseResolve raises an error (path is not a valid _node.md path),
        raise error "walk error: filesystem error while traversing".
     c. Create a DiscoveredNode record with:
          logical_name = <derived logical name>
          file_path    = <relative path from project root>
     d. Add the record to the result list.

  5. If the result list is empty (no _node.md files were found anywhere under
     "code-from-spec/"), raise error "no nodes found: code-from-spec/ contains no _node.md files".

  6. Sort the result list alphabetically by logical_name (standard lexicographic order).

  7. Return the sorted list of DiscoveredNode records.
```

## Error Conditions

| Error | Trigger |
|---|---|
| `"directory not found: code-from-spec/ does not exist"` | The `code-from-spec/` directory is absent from the working directory. |
| `"walk error: filesystem error while traversing"` | Any I/O or permission error during recursive directory traversal, or a discovered path that fails ReverseResolve. |
| `"no nodes found: code-from-spec/ contains no _node.md files"` | The walk completed successfully but zero `_node.md` files were found. |

## Contracts

- Only files named exactly `_node.md` are treated as nodes; all other files are ignored.
- The returned list is sorted alphabetically by `logical_name`.
- Sorting uses standard lexicographic (string) order on the `logical_name` field.
- Logical names are derived exclusively via the ReverseResolve function from
  `ROOT/functional/utils/logical_names`; no ad-hoc path manipulation is performed here.
