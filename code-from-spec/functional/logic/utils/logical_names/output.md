<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@eB8htIL5nYO2MkzggiLgi72X6kY -->

# Logical Names

All functions are pure — no I/O, no filesystem access.

---

## Records

```
record pathutils.PathCfs
  value: string
```

---

## Functions

### LogicalNameToPath

```
function LogicalNameToPath(logical_name: string) -> pathutils.PathCfs
```

Converts a `ROOT/` logical name to the `PathCfs` of the corresponding
`_node.md` file. Strips any qualifier before resolving.

Steps:

  1. If logical_name does not start with "ROOT" (i.e., neither "ROOT"
     nor "ROOT/..."), raise error UnsupportedReference:
     "logical name must be a ROOT/ reference".

  2. Call LogicalNameStripQualifier(logical_name) to get the bare name.

  3. If the bare name is exactly "ROOT":
     Return pathutils.PathCfs with value "code-from-spec/_node.md".

  4. Verify the bare name starts with "ROOT/".
     If not, raise error UnsupportedReference:
     "logical name must be a ROOT/ reference".

  5. Strip the "ROOT/" prefix from the bare name to get the relative segment.
     For example, "ROOT/x/y" → "x/y".

  6. Return pathutils.PathCfs with value
     "code-from-spec/<relative segment>/_node.md".

Errors:
  - UnsupportedReference: the logical name is not a ROOT/ reference
    (neither "ROOT" nor "ROOT/...").
```

---

### LogicalNameFromPath

```
function LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string
```

Derives the `ROOT/` logical name from a `_node.md` file path.
The inverse of `LogicalNameToPath`. Always returns a `ROOT/` reference.

Steps:

  1. Let path_value be cfs_path.value.

  2. If path_value does not start with "code-from-spec/", raise error
     InvalidPath: "path is not under code-from-spec/".

  3. If path_value does not end with "/_node.md" and is not exactly
     "code-from-spec/_node.md", raise error InvalidPath:
     "path is not a _node.md file".

  4. If path_value is exactly "code-from-spec/_node.md":
     Return "ROOT".

  5. Strip the leading "code-from-spec/" prefix and trailing "/_node.md"
     suffix from path_value to get the middle segment.
     For example, "code-from-spec/x/y/_node.md" → "x/y".

  6. Return "ROOT/<middle segment>".

Errors:
  - InvalidPath: the path is not a _node.md file under code-from-spec/.
```

---

### LogicalNameGetParent

```
function LogicalNameGetParent(logical_name: string) -> string
```

Returns the logical name of the parent node.
Strips any qualifier before computing the parent.

Steps:

  1. If logical_name does not start with "ROOT" (i.e., neither "ROOT"
     nor "ROOT/..."), raise error NotARootReference:
     "logical name must be a ROOT/ reference".

  2. Call LogicalNameStripQualifier(logical_name) to get the bare name.

  3. If the bare name is exactly "ROOT":
     Raise error NoParent: "ROOT has no parent".

  4. If the bare name does not contain "/" after "ROOT":
     Raise error NotARootReference:
     "logical name must be a ROOT/ reference".

  5. Find the last "/" in the bare name.
     Let parent be the substring before that last "/".

  6. If parent is "ROOT":
     Return "ROOT".

  7. Return parent.

Errors:
  - NoParent: the logical name is ROOT itself.
  - NotARootReference: the logical name is not a ROOT/ reference
    (neither "ROOT" nor "ROOT/...").
```

---

### LogicalNameGetQualifier

```
function LogicalNameGetQualifier(logical_name: string) -> optional string
```

Extracts the parenthetical qualifier from a logical name.
Returns absent if no qualifier is present.
Works with both `ROOT/` and `ARTIFACT/` references.

Steps:

  1. Search logical_name for the pattern "(<qualifier>)" at the end,
     where <qualifier> is one or more characters inside parentheses
     and the closing ")" is the last character of logical_name.

  2. If such a pattern is found:
     Return the content between the parentheses (not including "(" or ")").

  3. Otherwise:
     Return absent.

Errors: none.
```

---

### LogicalNameStripQualifier

```
function LogicalNameStripQualifier(logical_name: string) -> string
```

Returns the logical name without the parenthetical qualifier.
If no qualifier is present, returns the input unchanged.
Works with both `ROOT/` and `ARTIFACT/` references.

Steps:

  1. Search logical_name for the pattern "(<qualifier>)" at the end,
     where ")" is the last character.

  2. If such a pattern is found:
     Return the substring of logical_name before the opening "(".

  3. Otherwise:
     Return logical_name unchanged.

Errors: none.
```

---

### LogicalNameHasParent

```
function LogicalNameHasParent(logical_name: string) -> boolean
```

Returns true if the logical name is a `ROOT/` reference other than
`ROOT` itself. Returns false for `ROOT`, `ARTIFACT/` references,
and unrecognized prefixes.

Steps:

  1. Call LogicalNameStripQualifier(logical_name) to get the bare name.

  2. If bare name is exactly "ROOT":
     Return false.

  3. If bare name starts with "ROOT/":
     Return true.

  4. Return false.

Errors: none.
```

---

### LogicalNameHasQualifier

```
function LogicalNameHasQualifier(logical_name: string) -> boolean
```

Returns true if the logical name contains a parenthetical qualifier.
Works with both `ROOT/` and `ARTIFACT/` references.

Steps:

  1. Call LogicalNameGetQualifier(logical_name).

  2. If the result is not absent:
     Return true.

  3. Return false.

Errors: none.
```

---

### LogicalNameIsArtifact

```
function LogicalNameIsArtifact(logical_name: string) -> boolean
```

Returns true if the logical name starts with "ARTIFACT/".

Steps:

  1. If logical_name starts with "ARTIFACT/":
     Return true.

  2. Return false.

Errors: none.
```

---

### LogicalNameGetArtifactGenerator

```
function LogicalNameGetArtifactGenerator(logical_name: string) -> string
```

Returns the `ROOT/` logical name of the node that generates the
referenced artifact. Strips the `ARTIFACT/` prefix and any qualifier.

Steps:

  1. If logical_name does not start with "ARTIFACT/":
     Raise error NotAnArtifactReference:
     "logical name does not start with ARTIFACT/".

  2. Call LogicalNameStripQualifier(logical_name) to get the bare name.

  3. Strip the "ARTIFACT/" prefix from the bare name to get the segment.
     For example, "ARTIFACT/x/y" → "x/y".

  4. Return "ROOT/<segment>".

Errors:
  - NotAnArtifactReference: the logical name does not start with ARTIFACT/.
```
