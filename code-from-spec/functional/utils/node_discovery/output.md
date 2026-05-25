<!-- code-from-spec: ROOT/functional/utils/node_discovery@PENDING -->

## Data structures

```
record DiscoveredNode
  logical_name: string
  file_path: string
```

## Functions

### DiscoverNodes() -> list of DiscoveredNode

1. Determine the project root as the current working directory.

2. Check that the directory "code-from-spec/" exists relative to
   the project root.
   If it does not exist, raise error "directory not found".

3. Walk the "code-from-spec/" directory and all its subdirectories
   recursively.
   If a filesystem error occurs during traversal,
   raise error "walk error".

4. For each file encountered during the walk:
   a. If the file name is exactly "_node.md", record it.
   b. Otherwise, ignore it.

5. For each recorded "_node.md" file:
   a. Derive the logical name from the file path using
      logical_names reverse resolution.
   b. Create a DiscoveredNode record with the logical name
      and the file path.

6. If no nodes were found, raise error "no nodes found".

7. Sort the list of DiscoveredNode records alphabetically
   by logical_name.

8. Return the sorted list.
