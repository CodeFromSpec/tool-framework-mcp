<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@XKPP54rZmTM7o3Qf_mvomxLEcbk -->

# logical_names

Pure string manipulation functions for converting between logical names and
file paths. No I/O is performed. All returned paths use forward slashes.

---

## Data types

PathCfs — a record with field:
  - value: string (forward-slash-separated relative path from project root)

---

## Functions

---

### LogicalNameToPath(logical_name: string) -> PathCfs

Converts a `ROOT/` logical name to the PathCfs of its `_node.md` file.
Qualifiers are stripped before resolving.

  1. If logical_name does not start with "ROOT/" and is not exactly "ROOT",
     raise error "unsupported reference".

  2. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  3. If stripped is exactly "ROOT",
     return PathCfs with value "code-from-spec/_node.md".

  4. Remove the leading "ROOT/" prefix from stripped.
     Let relative = result.

  5. Replace all "/" separators in relative with "/".
     (They are already forward slashes — no change needed.)

  6. Return PathCfs with value "code-from-spec/<relative>/_node.md".

Errors:
  - "unsupported reference": logical_name does not start with "ROOT/" and
    is not exactly "ROOT".

---

### LogicalNameFromPath(cfs_path: PathCfs) -> string

Derives the `ROOT/` logical name from a `_node.md` file path.
The inverse of LogicalNameToPath.

  1. Let path = cfs_path.value.

  2. If path does not start with "code-from-spec/" or does not end with
     "/_node.md",
     raise error "invalid path".

  3. If path is exactly "code-from-spec/_node.md",
     return "ROOT".

  4. Remove the leading "code-from-spec/" prefix from path.
     Remove the trailing "/_node.md" suffix from the result.
     Let middle = result.

  5. If middle is empty,
     raise error "invalid path".

  6. Return "ROOT/<middle>".

Errors:
  - "invalid path": the path is not a _node.md file under code-from-spec/.

---

### LogicalNameGetParent(logical_name: string) -> string

Returns the logical name of the parent node.
Strips any qualifier before computing the parent.

  1. If logical_name does not start with "ROOT/" and is not exactly "ROOT",
     raise error "not a ROOT reference".

  2. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let stripped = result.

  3. If stripped is exactly "ROOT",
     raise error "no parent".

  4. Find the last "/" in stripped.
     Let parent = everything before that "/".

  5. If parent is empty,
     raise error "no parent".

  6. Return parent.

Errors:
  - "no parent": the logical name is ROOT itself.
  - "not a ROOT reference": the logical name does not start with "ROOT/".

---

### LogicalNameGetQualifier(logical_name: string) -> optional string

Extracts the parenthetical qualifier from a logical name.
Returns absent if no qualifier is present.
Works with both `ROOT/` and `ARTIFACT/` references.

  1. Look for an opening "(" character in logical_name.
     If not found, return absent.

  2. Let open_pos = position of "(".
     Let close_pos = position of ")" after open_pos.
     If ")" is not found, return absent.

  3. Let qualifier = substring of logical_name between open_pos and
     close_pos (exclusive).

  4. If qualifier is empty, return absent.

  5. Return qualifier.

---

### LogicalNameStripQualifier(logical_name: string) -> string

Returns the logical name with the parenthetical qualifier removed.
If no qualifier is present, returns the input unchanged.
Works with both `ROOT/` and `ARTIFACT/` references.

  1. Look for an opening "(" character in logical_name.
     If not found, return logical_name unchanged.

  2. Let open_pos = position of "(".

  3. Return substring of logical_name from the start up to (but not
     including) open_pos.

---

### LogicalNameHasParent(logical_name: string) -> boolean

Returns true if the logical name is a `ROOT/` reference other than
`ROOT` itself.

  1. If logical_name does not start with "ROOT/",
     return false.

  2. Strip any qualifier using LogicalNameStripQualifier.
     Let stripped = result.

  3. If stripped is exactly "ROOT",
     return false.

  4. Return true.

---

### LogicalNameHasQualifier(logical_name: string) -> boolean

Returns true if the logical name contains a parenthetical qualifier.
Works with both `ROOT/` and `ARTIFACT/` references.

  1. Call LogicalNameGetQualifier(logical_name).

  2. If the result is absent, return false.

  3. Return true.

---

### LogicalNameIsArtifact(logical_name: string) -> boolean

Returns true if the logical name starts with `ARTIFACT/`.

  1. If logical_name starts with "ARTIFACT/",
     return true.

  2. Return false.

---

### LogicalNameGetArtifactGenerator(logical_name: string) -> string

Returns the `ROOT/` logical name of the node that generates the
referenced artifact.
Strips the `ARTIFACT/` prefix and any qualifier.

  1. If logical_name does not start with "ARTIFACT/",
     raise error "not an artifact reference".

  2. Strip any qualifier using LogicalNameStripQualifier.
     Let stripped = result.

  3. Remove the leading "ARTIFACT/" prefix from stripped.
     Let relative = result.

  4. Return "ROOT/<relative>".

Errors:
  - "not an artifact reference": logical_name does not start with "ARTIFACT/".

---

## Contracts

- All returned PathCfs values use forward slashes as separators.
- All functions are pure — no I/O is performed.
- Unrecognized prefixes cause functions that declare errors to raise an error,
  and boolean-returning functions to return false.
