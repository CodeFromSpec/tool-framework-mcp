<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@XKPP54rZmTM7o3Qf_mvomxLEcbk -->

## LogicalNameToPath(logical_name: string) -> PathCfs

Converts a `ROOT/` logical name to the `PathCfs` of the corresponding `_node.md` file.

Parameters:
- logical_name: string — a logical name starting with `ROOT/` or exactly `ROOT`

Returns: PathCfs

Errors:
- "unsupported reference": the logical name does not start with `ROOT/` and is not `ROOT`

Steps:

  1. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  2. If stripped is not "ROOT" and does not start with "ROOT/",
     raise error "unsupported reference".

  3. If stripped is exactly "ROOT",
     return PathCfs with value "code-from-spec/_node.md".

  4. Remove the leading "ROOT/" prefix from stripped.
     Let relative = remainder.

  5. Replace every "/" in relative with "/".
     (No-op — separators are already forward slashes.)

  6. Return PathCfs with value "code-from-spec/<relative>/_node.md".


## LogicalNameFromPath(cfs_path: PathCfs) -> string

Derives the `ROOT/` logical name from a `_node.md` file path.

Parameters:
- cfs_path: PathCfs — path to a `_node.md` file under `code-from-spec/`

Returns: string — a `ROOT/` logical name

Errors:
- "invalid path": the path does not end with `_node.md` or does not start with `code-from-spec/`

Steps:

  1. Let path = cfs_path.value.

  2. If path does not start with "code-from-spec/",
     raise error "invalid path".

  3. If path does not end with "_node.md",
     raise error "invalid path".

  4. If path is exactly "code-from-spec/_node.md",
     return "ROOT".

  5. Remove the leading "code-from-spec/" prefix from path.
     Let middle = remainder.

  6. Remove the trailing "/_node.md" suffix from middle.
     Let relative = remainder.

  7. If relative is empty,
     raise error "invalid path".

  8. Return "ROOT/<relative>".


## LogicalNameGetParent(logical_name: string) -> string

Returns the logical name of the parent node.

Parameters:
- logical_name: string — a `ROOT/` logical name

Returns: string — the parent logical name

Errors:
- "not a ROOT reference": the logical name does not start with `ROOT/` and is not `ROOT`
- "no parent": the logical name is exactly `ROOT`

Steps:

  1. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  2. If stripped is not "ROOT" and does not start with "ROOT/",
     raise error "not a ROOT reference".

  3. If stripped is exactly "ROOT",
     raise error "no parent".

  4. Remove the leading "ROOT/" prefix from stripped.
     Let relative = remainder.

  5. Find the last "/" in relative.
     If no "/" is found,
       return "ROOT".
     Else
       let parent_relative = everything before the last "/".
       return "ROOT/<parent_relative>".


## LogicalNameGetQualifier(logical_name: string) -> optional string

Extracts the parenthetical qualifier from a logical name.

Parameters:
- logical_name: string — any logical name

Returns: optional string — the qualifier text, or absent if none

Steps:

  1. Search logical_name for an opening "(" character.
     If not found, return absent.

  2. Let open_pos = position of the first "(".
  3. Let close_pos = position of the last ")" in logical_name.

  4. If close_pos is not found or close_pos is before open_pos,
     return absent.

  5. If close_pos is not the last character of logical_name,
     return absent.

  6. Let qualifier = characters between open_pos and close_pos (exclusive).

  7. If qualifier is empty,
     return absent.

  8. Return qualifier.


## LogicalNameStripQualifier(logical_name: string) -> string

Returns the logical name without any parenthetical qualifier.

Parameters:
- logical_name: string — any logical name

Returns: string — logical name without qualifier

Steps:

  1. Search logical_name for an opening "(" character.
     If not found, return logical_name unchanged.

  2. Let open_pos = position of the first "(".
  3. Let close_pos = position of the last ")" in logical_name.

  4. If close_pos is not found or close_pos is before open_pos,
     return logical_name unchanged.

  5. If close_pos is not the last character of logical_name,
     return logical_name unchanged.

  6. Return characters from the start of logical_name up to (but not including) open_pos.


## LogicalNameHasParent(logical_name: string) -> boolean

Returns true if the logical name is a `ROOT/` reference other than `ROOT` itself.

Parameters:
- logical_name: string — any logical name

Returns: boolean

Steps:

  1. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  2. If stripped is exactly "ROOT",
     return false.

  3. If stripped starts with "ROOT/",
     return true.

  4. Return false.


## LogicalNameHasQualifier(logical_name: string) -> boolean

Returns true if the logical name contains a parenthetical qualifier.

Parameters:
- logical_name: string — any logical name

Returns: boolean

Steps:

  1. Call LogicalNameGetQualifier(logical_name).

  2. If the result is absent, return false.

  3. Return true.


## LogicalNameIsArtifact(logical_name: string) -> boolean

Returns true if the logical name starts with `ARTIFACT/`.

Parameters:
- logical_name: string — any logical name

Returns: boolean

Steps:

  1. If logical_name starts with "ARTIFACT/",
     return true.

  2. Return false.


## LogicalNameGetArtifactGenerator(logical_name: string) -> string

Returns the `ROOT/` logical name of the node that generates the referenced artifact.

Parameters:
- logical_name: string — a logical name starting with `ARTIFACT/`

Returns: string — a `ROOT/` logical name

Errors:
- "not an artifact reference": the logical name does not start with `ARTIFACT/`

Steps:

  1. If logical_name does not start with "ARTIFACT/",
     raise error "not an artifact reference".

  2. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  3. Remove the leading "ARTIFACT/" prefix from stripped.
     Let relative = remainder.

  4. Return "ROOT/<relative>".
