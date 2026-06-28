---
depends_on:
  - ARTIFACT/golang/interfaces/utils/logical_names
  - ARTIFACT/golang/interfaces/os/path_utils
output: internal/logicalnames/logicalnames.go
---

# SPEC/golang/implementation/utils/logical_names

# Agent

Implement the logical names component as a Go package.

## Logic

### LogicalNameToPath(logical_name: string) -> pathutils.PathCfs

1. Call LogicalNameStripQualifier(logical_name) to
   remove any qualifier. Store as `stripped`.
2. If `stripped` is not exactly "SPEC" and does not
   start with "SPEC/", raise error
   UnsupportedReference.
3. If `stripped` is exactly "SPEC", return PathCfs
   with value "code-from-spec/_node.md".
4. Remove the leading "SPEC/" prefix from `stripped`.
   Store as `relative_path`.
5. Return PathCfs with value
   "code-from-spec/" + relative_path + "/_node.md".

### LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string

1. Let `path_value` be the value field of cfs_path.
2. If path_value does not end with "/_node.md" and is
   not exactly "code-from-spec/_node.md", raise error
   InvalidPath.
3. If path_value does not start with "code-from-spec/",
   raise error InvalidPath.
4. If path_value is exactly "code-from-spec/_node.md",
   return "SPEC".
5. Remove the leading "code-from-spec/" prefix. Remove
   the trailing "/_node.md" suffix. Store as
   `relative_path`.
6. Return "SPEC/" + relative_path.

### LogicalNameGetParent(logical_name: string) -> string

1. Call LogicalNameStripQualifier(logical_name). Store
   as `stripped`.
2. If `stripped` is not exactly "SPEC" and does not
   start with "SPEC/", raise error NotASpecReference.
3. If `stripped` is exactly "SPEC", raise error
   NoParent.
4. Remove the leading "SPEC/" prefix. Store as
   `relative_path`.
5. Find the last "/" in `relative_path`.
6. If no "/" is found, return "SPEC".
7. Take the substring up to the last "/". Store as
   `parent_relative`.
8. Return "SPEC/" + parent_relative.

### LogicalNameGetQualifier(logical_name: string) -> optional string

1. Find the first "(" in logical_name.
2. If no "(", return absent.
3. Find the first ")" after the "(".
4. If no ")", return absent.
5. Extract substring between "(" and ")" (exclusive).
6. Return it.

### LogicalNameStripQualifier(logical_name: string) -> string

1. Find the first "(" in logical_name.
2. If no "(", return logical_name unchanged.
3. Return the substring up to (not including) the "(".

### LogicalNameHasParent(logical_name: string) -> boolean

1. Call LogicalNameStripQualifier(logical_name). Store
   as `stripped`.
2. If `stripped` starts with "SPEC/", return true.
3. Return false.

### LogicalNameHasQualifier(logical_name: string) -> boolean

1. Call LogicalNameGetQualifier(logical_name).
2. If absent, return false.
3. Return true.

### LogicalNameIsArtifact(logical_name: string) -> boolean

1. If logical_name starts with "ARTIFACT/", return true.
2. Return false.

### LogicalNameIsSpec(logical_name: string) -> boolean

1. If logical_name is exactly "SPEC", return true.
2. If logical_name starts with "SPEC/", return true.
3. Return false.

### LogicalNameIsExternal(logical_name: string) -> boolean

1. If logical_name starts with "EXTERNAL/", return true.
2. Return false.

### LogicalNameGetArtifactGenerator(logical_name: string) -> string

1. If logical_name does not start with "ARTIFACT/",
   raise error NotAnArtifactReference.
2. Remove the leading "ARTIFACT/" prefix. Store as
   `relative_path`.
3. Return "SPEC/" + relative_path.

### LogicalNameExternalToPath(logical_name: string) -> pathutils.PathCfs

1. If logical_name does not start with "EXTERNAL/",
   raise error NotAnExternalReference.
2. Remove the leading "EXTERNAL/" prefix. Store as
   `relative_path`.
3. Return PathCfs with value relative_path.

## Go-specific guidance

- Use `filepath` and `path` standard library packages
  for path manipulation.
- The package name should be `logicalnames`.
- Functions that declare errors in the spec should
  return `(result, error)` in Go.
- Functions that return `optional` in the spec should
  return `(result, bool)` in Go.
- Boolean functions return a single `bool`.
