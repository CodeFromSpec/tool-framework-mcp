<!-- code-from-spec: SPEC/functional/logic/utils/logical_names@eTWWAItkM73cT3-I8ekwhtieCR0 -->

# logical_names

All functions are pure — no I/O.


## Records

No records are declared in this module. Types referenced from
pathutils are qualified as `pathutils.PathCfs`.


## Functions

---

### LogicalNameToPath(logical_name: string) -> pathutils.PathCfs

Converts a `SPEC/` logical name to the `PathCfs` of the
corresponding `_node.md` file.

1. Call `LogicalNameStripQualifier(logical_name)` to remove any
   parenthetical qualifier. Store the result as `stripped`.

2. If `stripped` is not exactly `"SPEC"` and does not start with
   `"SPEC/"`, raise error `UnsupportedReference`.

3. If `stripped` is exactly `"SPEC"`, return a `pathutils.PathCfs`
   with value `"code-from-spec/_node.md"`.

4. Remove the leading `"SPEC/"` prefix from `stripped`.
   Store the remainder as `relative_path`.

5. Return a `pathutils.PathCfs` with value
   `"code-from-spec/" + relative_path + "/_node.md"`.

Errors:
- `UnsupportedReference`: the logical name (after stripping qualifier)
  is not exactly `"SPEC"` and does not start with `"SPEC/"`.

---

### LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string

Derives the `SPEC/` logical name from a `_node.md` file path.

1. Let `path_value` be the `value` field of `cfs_path`.

2. If `path_value` does not end with `"/_node.md"` and is not
   exactly `"code-from-spec/_node.md"`, raise error `InvalidPath`.

3. If `path_value` does not start with `"code-from-spec/"`,
   raise error `InvalidPath`.

4. If `path_value` is exactly `"code-from-spec/_node.md"`,
   return `"SPEC"`.

5. Remove the leading `"code-from-spec/"` prefix from `path_value`.
   Remove the trailing `"/_node.md"` suffix.
   Store the remainder as `relative_path`.

6. Return `"SPEC/" + relative_path`.

Errors:
- `InvalidPath`: the path is not a `_node.md` file under
  `code-from-spec/`.

---

### LogicalNameGetParent(logical_name: string) -> string

Returns the logical name of the parent node.

1. Call `LogicalNameStripQualifier(logical_name)` to remove any
   parenthetical qualifier. Store the result as `stripped`.

2. If `stripped` is not exactly `"SPEC"` and does not start with
   `"SPEC/"`, raise error `NotASpecReference`.

3. If `stripped` is exactly `"SPEC"`, raise error `NoParent`.

4. Remove the leading `"SPEC/"` prefix from `stripped`.
   Store the remainder as `relative_path`.

5. Find the last `"/"` character in `relative_path`.

6. If no `"/"` is found, the parent is the root — return `"SPEC"`.

7. Take the substring of `relative_path` up to (but not including)
   the last `"/"`. Store as `parent_relative`.

8. Return `"SPEC/" + parent_relative`.

Errors:
- `NoParent`: the logical name (after stripping qualifier) is
  exactly `"SPEC"`.
- `NotASpecReference`: the logical name is not exactly `"SPEC"`
  and does not start with `"SPEC/"`.

---

### LogicalNameGetQualifier(logical_name: string) -> optional string

Extracts the parenthetical qualifier from a logical name.

1. Find the first `"("` character in `logical_name`.

2. If no `"("` is found, return absent.

3. Find the first `")"` character in `logical_name` after the `"("`.

4. If no `")"` is found, return absent.

5. Extract the substring between `"("` and `")"` (exclusive).
   Store as `qualifier`.

6. Return `qualifier`.

---

### LogicalNameStripQualifier(logical_name: string) -> string

Returns the logical name without the parenthetical qualifier.

1. Find the first `"("` character in `logical_name`.

2. If no `"("` is found, return `logical_name` unchanged.

3. Return the substring of `logical_name` up to (but not including)
   the `"("`.

---

### LogicalNameHasParent(logical_name: string) -> boolean

Returns true if the logical name is a `SPEC/` reference other than
`SPEC` itself.

1. Call `LogicalNameStripQualifier(logical_name)`. Store as `stripped`.

2. If `stripped` starts with `"SPEC/"`, return true.

3. Return false.

---

### LogicalNameHasQualifier(logical_name: string) -> boolean

Returns true if the logical name contains a parenthetical qualifier.

1. Call `LogicalNameGetQualifier(logical_name)`.

2. If the result is absent, return false.

3. Return true.

---

### LogicalNameIsArtifact(logical_name: string) -> boolean

Returns true if the logical name starts with `ARTIFACT/`.

1. If `logical_name` starts with `"ARTIFACT/"`, return true.

2. Return false.

---

### LogicalNameIsSpec(logical_name: string) -> boolean

Returns true if the logical name is exactly `SPEC` or starts with
`SPEC/`.

1. If `logical_name` is exactly `"SPEC"`, return true.

2. If `logical_name` starts with `"SPEC/"`, return true.

3. Return false.

---

### LogicalNameIsExternal(logical_name: string) -> boolean

Returns true if the logical name starts with `EXTERNAL/`.

1. If `logical_name` starts with `"EXTERNAL/"`, return true.

2. Return false.

---

### LogicalNameGetArtifactGenerator(logical_name: string) -> string

Returns the `SPEC/` logical name of the node that generates the
referenced artifact.

1. If `logical_name` does not start with `"ARTIFACT/"`, raise error
   `NotAnArtifactReference`.

2. Remove the leading `"ARTIFACT/"` prefix from `logical_name`.
   Store the remainder as `relative_path`.

3. Return `"SPEC/" + relative_path`.

Errors:
- `NotAnArtifactReference`: the logical name does not start with
  `"ARTIFACT/"`.

---

### LogicalNameExternalToPath(logical_name: string) -> pathutils.PathCfs

Converts an `EXTERNAL/` logical name to a `PathCfs`.

1. If `logical_name` does not start with `"EXTERNAL/"`, raise error
   `NotAnExternalReference`.

2. Remove the leading `"EXTERNAL/"` prefix from `logical_name`.
   Store the remainder as `relative_path`.

3. Return a `pathutils.PathCfs` with value `relative_path`.

Errors:
- `NotAnExternalReference`: the logical name does not start with
  `"EXTERNAL/"`.
