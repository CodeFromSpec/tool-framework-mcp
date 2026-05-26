<!-- code-from-spec: ROOT/functional/utils/node_discovery@nFE7ic64qvmsWenz5dvUFCOfhaQ -->

# Node Discovery

## Records

```
record DiscoveredNode
  logical_name: string   -- the ROOT/ logical name derived from the file path
  file_path: string      -- relative path from the project root, e.g. code-from-spec/x/y/_node.md
```

## Functions

```
function DiscoverNodes() -> list of DiscoveredNode
  -- Scans the code-from-spec/ directory tree, finds every _node.md file,
  -- derives its logical name via reverse resolution, and returns the
  -- discovered nodes sorted alphabetically by logical name.
  --
  -- No parameters. The starting directory is always code-from-spec/
  -- relative to the current working directory (project root).

  1. Verify that code-from-spec/ exists in the current working directory.
     If it does not exist, raise error "directory not found: code-from-spec/ does not exist".

  2. Walk the code-from-spec/ directory tree recursively, visiting every
     file and subdirectory.
     If a filesystem error occurs during the walk, raise error "walk error: <description>".

  3. For each file encountered during the walk:
     If the file's name is not exactly "_node.md", skip it and continue to the next file.
     If the file's name is "_node.md":
       a. Record the file path relative to the project root
          (e.g. "code-from-spec/x/y/_node.md").
       b. Derive the logical name by applying ReverseResolve to the file path
          (as defined in ROOT/functional/utils/logical_names).
          -- Mapping examples:
          --   code-from-spec/_node.md          -> ROOT
          --   code-from-spec/x/_node.md        -> ROOT/x
          --   code-from-spec/x/y/_node.md      -> ROOT/x/y
       c. Create a DiscoveredNode with the derived logical_name and file_path.
       d. Add the DiscoveredNode to the result list.

  4. After the walk is complete, check whether the result list is empty.
     If it is empty, raise error "no nodes found: code-from-spec/ contains no _node.md files".

  5. Sort the result list alphabetically by logical_name (ascending, lexicographic order).

  6. Return the sorted list of DiscoveredNode records.
```

## Error Conditions

| Situation | Error message |
|---|---|
| `code-from-spec/` directory does not exist | `"directory not found: code-from-spec/ does not exist"` |
| Filesystem error while traversing the directory | `"walk error: <description>"` |
| Walk completes but no `_node.md` files were found | `"no nodes found: code-from-spec/ contains no _node.md files"` |

## Contracts and Invariants

- Only files named exactly `_node.md` are treated as nodes. All other files are ignored.
- The returned list is always sorted alphabetically by `logical_name`.
- `file_path` values are relative to the project root (starting with `code-from-spec/`).
- `logical_name` values always start with `ROOT` and use `/` as a separator.
- This function has no parameters; the root directory is fixed to `code-from-spec/`.
