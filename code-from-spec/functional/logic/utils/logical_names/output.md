<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@kYPkbgROEUvCmHXG-jOkDax0RCI -->

# Logical Names — Pseudocode

## Data Types

record PathCfs
  value: string   (forward-slash relative path from project root)

---

## function LogicalNameToPath(logical_name: string) -> PathCfs

  1. If logical_name does not start with "ROOT/", and logical_name is not exactly "ROOT",
     raise error "UnsupportedReference".

  2. Strip any parenthetical qualifier from logical_name.
     (e.g. "ROOT/x/y(z)" becomes "ROOT/x/y")

  3. If the stripped name is exactly "ROOT",
     return PathCfs with value "code-from-spec/_node.md".

  4. Remove the "ROOT/" prefix from the stripped name to get the relative segment.
     (e.g. "ROOT/x/y" → "x/y")

  5. Return PathCfs with value "code-from-spec/<relative segment>/_node.md".

---

## function LogicalNameFromPath(cfs_path: PathCfs) -> string

  1. Let path be cfs_path.value.

  2. If path does not start with "code-from-spec/", raise error "InvalidPath".

  3. If path does not end with "/_node.md" and path is not exactly "code-from-spec/_node.md",
     raise error "InvalidPath".

  4. If path is exactly "code-from-spec/_node.md",
     return "ROOT".

  5. Remove the leading "code-from-spec/" prefix and the trailing "/_node.md" suffix
     from path to get the middle segment.
     (e.g. "code-from-spec/x/y/_node.md" → "x/y")

  6. Return "ROOT/<middle segment>".

---

## function LogicalNameGetParent(logical_name: string) -> string

  1. If logical_name does not start with "ROOT/",
     raise error "NotARootReference".

  2. Strip any parenthetical qualifier from logical_name.

  3. If the stripped name is exactly "ROOT",
     raise error "NoParent".

  4. Remove the "ROOT/" prefix to get the segment.
     (e.g. "ROOT/x/y" → "x/y")

  5. Find the last "/" in the segment.
     If no "/" is found, return "ROOT".
     Otherwise, take everything before the last "/" as the parent segment
     and return "ROOT/<parent segment>".

---

## function LogicalNameGetQualifier(logical_name: string) -> optional string

  1. Search logical_name for an opening "(" character.
     If not found, return absent.

  2. Let start be the index immediately after "(".
     Search for a closing ")" character after start.
     If not found, return absent.

  3. Extract the substring between "(" and ")" (exclusive).
     If the substring is empty, return absent.

  4. Return the extracted substring.

---

## function LogicalNameStripQualifier(logical_name: string) -> string

  1. Search logical_name for an opening "(" character.
     If not found, return logical_name unchanged.

  2. Find the matching closing ")" character.
     If not found, return logical_name unchanged.

  3. Return the substring of logical_name before the "(" character,
     concatenated with the substring after the ")" character.
     (e.g. "ROOT/x/y(z)" → "ROOT/x/y"; "ARTIFACT/x/y(id)" → "ARTIFACT/x/y")

---

## function LogicalNameHasParent(logical_name: string) -> boolean

  1. If logical_name does not start with "ROOT/", return false.

  2. Strip any parenthetical qualifier from logical_name.

  3. If the stripped name is exactly "ROOT", return false.

  4. Return true.

---

## function LogicalNameHasQualifier(logical_name: string) -> boolean

  1. Call LogicalNameGetQualifier(logical_name).
     If the result is absent, return false.
     Otherwise return true.

---

## function LogicalNameIsArtifact(logical_name: string) -> boolean

  1. If logical_name starts with "ARTIFACT/", return true.
     Otherwise return false.

---

## function LogicalNameGetArtifactGenerator(logical_name: string) -> string

  1. If logical_name does not start with "ARTIFACT/",
     raise error "NotAnArtifactReference".

  2. Strip any parenthetical qualifier from logical_name.

  3. Remove the "ARTIFACT/" prefix to get the path segment.
     (e.g. "ARTIFACT/x/y" → "x/y")

  4. Return "ROOT/<path segment>".
     (e.g. "ARTIFACT/x/y(id)" → "ROOT/x/y")
