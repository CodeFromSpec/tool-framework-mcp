<!-- code-from-spec: ROOT/functional/utils/node_discovery@vD6RTE01A3cHzajAiepUEuGytpI -->

# node_discovery

## Records

```
record DiscoveredNode
  logical_name: string   -- the ROOT/ logical name derived from the file path
  file_path:    string   -- the relative path to the _node.md file
```

---

## Functions

---

### DiscoverNodes() -> list of DiscoveredNode

Traverses the `code-from-spec/` directory tree, collects every
`_node.md` file, converts each to a `DiscoveredNode` by calling
`ReverseResolve`, and returns the results sorted alphabetically
by logical name.

**Errors**

| Condition | Error |
|---|---|
| `code-from-spec/` does not exist | `"directory not found"` |
| filesystem error during traversal | `"walk error"` |
| no `_node.md` files found anywhere | `"no nodes found"` |

**Steps**

1. Set `root_dir` to `"code-from-spec/"` relative to the
   current working directory (project root).

2. Check that `root_dir` exists as a directory.
   If it does not exist, raise error `"directory not found"`.

3. Set `collected` to an empty list of DiscoveredNode.

4. Walk `root_dir` recursively, visiting every file and
   subdirectory at any depth.
   If the walk cannot be started or encounters a filesystem
   error at any point, raise error `"walk error"`.

5. For each entry visited during the walk:
   a. If the entry is a directory, skip it (continue the walk).
   b. If the entry's file name is NOT exactly `"_node.md"`,
      skip it (continue the walk).
   c. Otherwise, the entry is a `_node.md` file.
      Derive `file_path` as the path of this entry relative
      to the project root (e.g. `"code-from-spec/x/y/_node.md"`).
      Call `ReverseResolve(file_path)` to obtain `logical_name`.
        If `ReverseResolve` raises an error, propagate that
        error (the walk is aborted).
      Append a new DiscoveredNode record with
        `logical_name = logical_name`
        `file_path    = file_path`
      to `collected`.

6. After the walk completes, if `collected` is empty,
   raise error `"no nodes found"`.

7. Sort `collected` alphabetically by the `logical_name` field
   (ascending, case-sensitive lexicographic order).

8. Return `collected`.
