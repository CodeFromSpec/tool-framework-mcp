<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@K89veF6GOJK9Y04NVCRxo-vmc7c -->

# logical_names

Pure string-manipulation functions for converting between logical names and
file paths, navigating the node hierarchy, and extracting qualifier
components. No I/O is performed by any function in this module.

---

## Data types

All inputs and outputs are plain strings or optional strings. `PathCfs`
represents a forward-slash, root-relative path as defined in the path
utilities spec.

---

## function LogicalNameToPath(logical_name: string) -> PathCfs

Converts a `ROOT/` logical name to the `PathCfs` of its `_node.md` file.
Strips any qualifier before resolving.

**Errors**
- `"unsupported reference"` — logical_name does not start with `ROOT/` and
  is not exactly `ROOT`.

**Steps**

1. If logical_name does not start with `ROOT/` and is not exactly `"ROOT"`,
   raise error `"unsupported reference"`.

2. Strip any qualifier suffix:
   - Find the last `(` character in logical_name.
   - If found, set logical_name to the substring before that `(`.

3. If logical_name is exactly `"ROOT"`,
   return `"code-from-spec/_node.md"`.

4. Remove the leading `"ROOT/"` prefix to get the relative segment.
   Replace every `/` separator with `/` (no change needed — separators are
   already forward slashes).

5. Return `"code-from-spec/" + <relative segment> + "/_node.md"`.

---

## function LogicalNameFromPath(cfs_path: PathCfs) -> string

Derives the `ROOT/` logical name from a `_node.md` file path.
The inverse of `LogicalNameToPath`.

**Errors**
- `"invalid path"` — cfs_path is not a `_node.md` file whose path starts
  with `code-from-spec/`.

**Steps**

1. If cfs_path does not start with `"code-from-spec/"`,
   raise error `"invalid path"`.

2. If cfs_path does not end with `"/_node.md"` and is not exactly
   `"code-from-spec/_node.md"`,
   raise error `"invalid path"`.

3. If cfs_path is exactly `"code-from-spec/_node.md"`,
   return `"ROOT"`.

4. Remove the leading `"code-from-spec/"` prefix and the trailing
   `"/_node.md"` suffix to get the relative segment.

5. Return `"ROOT/" + <relative segment>`.

---

## function LogicalNameGetParent(logical_name: string) -> string

Returns the logical name of the parent node.
Strips any qualifier before computing the parent.

**Errors**
- `"not a ROOT reference"` — logical_name does not start with `ROOT/` and
  is not exactly `ROOT`.
- `"no parent"` — logical_name (after stripping qualifier) is exactly
  `"ROOT"`.

**Steps**

1. If logical_name does not start with `"ROOT/"` and is not exactly
   `"ROOT"`,
   raise error `"not a ROOT reference"`.

2. Strip any qualifier suffix:
   - Find the last `(` character in logical_name.
   - If found, set logical_name to the substring before that `(`.

3. If logical_name is exactly `"ROOT"`,
   raise error `"no parent"`.

4. Find the last `/` character in logical_name.
   Take the substring before that `/`.

5. Return that substring.
   (For `ROOT/x`, this returns `"ROOT"`.
    For `ROOT/x/y`, this returns `"ROOT/x"`.)

---

## function LogicalNameGetQualifier(logical_name: string) -> optional string

Extracts the parenthetical qualifier from a logical name.
Works with both `ROOT/` and `ARTIFACT/` references.

**Steps**

1. Find the last `(` character in logical_name.
   If not found, return absent.

2. Find the closing `)` character after that `(`.
   If not found, return absent.

3. Extract the substring between `(` and `)`.
   If the substring is empty, return absent.

4. Return that substring as the qualifier.

---

## function LogicalNameHasParent(logical_name: string) -> boolean

Returns true if the logical name is a `ROOT/` reference other than `ROOT`
itself.

**Steps**

1. If logical_name does not start with `"ROOT/"`,
   return false.

2. Return true.
   (Any name with the `ROOT/` prefix is a non-root `ROOT` node and has a
   parent. `"ROOT"` itself does not have the `ROOT/` prefix so it is
   already excluded.)

---

## function LogicalNameHasQualifier(logical_name: string) -> boolean

Returns true if the logical name contains a parenthetical qualifier.
Works with both `ROOT/` and `ARTIFACT/` references.

**Steps**

1. Call LogicalNameGetQualifier(logical_name).

2. If the result is absent, return false.

3. Otherwise return true.

---

## function LogicalNameIsArtifact(logical_name: string) -> boolean

Returns true if the logical name starts with `ARTIFACT/`.

**Steps**

1. If logical_name starts with `"ARTIFACT/"`,
   return true.

2. Otherwise return false.

---

## function LogicalNameGetArtifactGenerator(logical_name: string) -> string

Returns the `ROOT/` logical name of the node that generates the referenced
artifact. Strips the `ARTIFACT/` prefix and any qualifier.

**Errors**
- `"not an artifact reference"` — logical_name does not start with
  `ARTIFACT/`.

**Steps**

1. If logical_name does not start with `"ARTIFACT/"`,
   raise error `"not an artifact reference"`.

2. Remove the leading `"ARTIFACT/"` prefix to get the relative segment.

3. Strip any qualifier suffix from the relative segment:
   - Find the last `(` character in the relative segment.
   - If found, set the relative segment to the substring before that `(`.

4. Return `"ROOT/" + <relative segment>`.

---

## Error conditions summary

| Function                         | Error                      | Condition                                          |
|----------------------------------|----------------------------|----------------------------------------------------|
| LogicalNameToPath                | `"unsupported reference"`  | logical_name is not a ROOT reference               |
| LogicalNameFromPath              | `"invalid path"`           | path is not a `_node.md` under `code-from-spec/`  |
| LogicalNameGetParent             | `"not a ROOT reference"`   | logical_name is not a ROOT reference               |
| LogicalNameGetParent             | `"no parent"`              | logical_name resolves to ROOT itself               |
| LogicalNameGetArtifactGenerator  | `"not an artifact reference"` | logical_name does not start with `ARTIFACT/`    |
