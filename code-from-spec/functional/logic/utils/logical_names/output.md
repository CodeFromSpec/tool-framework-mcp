<!-- code-from-spec: ROOT/functional/utils/logical_names@xm6ZcowR5vUPONXL0WvI7hfV6vE -->

# Logical Names

All functions are pure — no I/O, no side effects.
All returned file paths use forward slashes as separators, regardless of OS.

---

## Data Structures

```
record ArtifactReference
  node_path:   string   -- the resolved _node.md file path for the node
  artifact_id: string   -- the id qualifier extracted from the logical name
```

---

## Functions

---

### function ResolvePath(logical_name) -> string

Resolves a `ROOT/` logical name to its corresponding `_node.md` file path.
Qualifiers are stripped before resolution.

Parameters:
  - logical_name: string — must start with "ROOT/"

Returns:
  - string — the file path relative to the project root, using forward slashes

Errors:
  - "unsupported reference: only ROOT/ logical names can be resolved to a path"
    raised when logical_name does not start with "ROOT"

Steps:

  1. If logical_name does not start with "ROOT", raise error
     "unsupported reference: only ROOT/ logical names can be resolved to a path".
     Note: ARTIFACT/ names require frontmatter lookup and cannot be statically resolved.

  2. Strip any qualifier from logical_name.
     A qualifier is a parenthetical suffix of the form "(z)" at the end.
     For example: "ROOT/x/y(z)" becomes "ROOT/x/y".

  3. If the stripped name is exactly "ROOT", return "code-from-spec/_node.md".

  4. Otherwise, take the portion after "ROOT/" and replace each "/" separator
     with "/" (they are already forward slashes — no transformation needed on the
     segment separators, but ensure OS path separators are not used).
     Append "/_node.md" to form the path.
     Prepend "code-from-spec/".

     Examples:
       "ROOT/x"     -> "code-from-spec/x/_node.md"
       "ROOT/x/y"   -> "code-from-spec/x/y/_node.md"
       "ROOT/x/y(z)"-> strip qualifier -> "ROOT/x/y" -> "code-from-spec/x/y/_node.md"

  5. Return the resulting path.

---

### function ResolveArtifactReference(logical_name) -> ArtifactReference

Parses an `ARTIFACT/` logical name and returns the node path and artifact id
as separate values. The node path can then be used to read the node's frontmatter
to locate the actual output file.

Parameters:
  - logical_name: string — must start with "ARTIFACT/" and contain a qualifier

Returns:
  - ArtifactReference record with fields:
      node_path:   the _node.md file path for the referenced node
      artifact_id: the qualifier extracted from the logical name

Errors:
  - "unrecognized prefix: the logical name does not start with ARTIFACT/"
    raised when logical_name does not start with "ARTIFACT/".
  - "missing qualifier: the logical name has no parenthetical qualifier"
    raised when logical_name starts with "ARTIFACT/" but contains no qualifier.

Steps:

  1. If logical_name does not start with "ARTIFACT/", raise error
     "unrecognized prefix: the logical name does not start with ARTIFACT/".

  2. Extract the qualifier from logical_name using ExtractQualifier.
     If no qualifier is present (result is absent), raise error
     "missing qualifier: the logical name has no parenthetical qualifier".

  3. Strip the qualifier from logical_name to get the bare node reference.
     For example: "ARTIFACT/x/y(id)" becomes "ARTIFACT/x/y".

  4. Take the portion after "ARTIFACT/" and build the node path:
     Prepend "code-from-spec/" and append "/_node.md".
     Ensure all path separators are forward slashes.

     Example:
       "ARTIFACT/x/y(id)" -> node_path = "code-from-spec/x/y/_node.md"
                           -> artifact_id = "id"

  5. Return an ArtifactReference with:
       node_path   = the path built in step 4
       artifact_id = the qualifier extracted in step 2

---

### function GetParent(logical_name) -> string

Returns the logical name of the parent node for a given `ROOT/` logical name.
Qualifiers are stripped before computing the parent.

Parameters:
  - logical_name: string — must start with "ROOT"

Returns:
  - string — the logical name of the parent node

Errors:
  - "no parent: the logical name is ROOT itself"
    raised when the node is ROOT (i.e., has no parent segment to remove).
  - "not a ROOT reference: the logical name is an ARTIFACT/ reference"
    raised when logical_name starts with "ARTIFACT/".

Steps:

  1. If logical_name starts with "ARTIFACT/", raise error
     "not a ROOT reference: the logical name is an ARTIFACT/ reference".

  2. Strip any qualifier from logical_name.
     For example: "ROOT/x/y(z)" becomes "ROOT/x/y".

  3. If the stripped name is exactly "ROOT", raise error
     "no parent: the logical name is ROOT itself".

  4. Find the last "/" in the stripped name.
     Remove everything from that "/" to the end.

     Examples:
       "ROOT/x"   -> last "/" is at index 4 -> remove "/x"   -> "ROOT"
       "ROOT/x/y" -> last "/" is at index 6 -> remove "/y"   -> "ROOT/x"

  5. Return the resulting string as the parent logical name.

---

### function ReverseResolve(file_path) -> string

Derives the logical name for a given `_node.md` file path.
The inverse of ResolvePath.

Parameters:
  - file_path: string — a relative file path from the project root

Returns:
  - string — the logical name corresponding to the file path

Errors:
  - "invalid path: the path is not a _node.md file under code-from-spec/"
    raised when the path does not match the expected pattern.

Steps:

  1. Normalize the path to use forward slashes (replace any backslashes with "/").

  2. If the normalized path does not start with "code-from-spec/", raise error
     "invalid path: the path is not a _node.md file under code-from-spec/".

  3. If the normalized path does not end with "/_node.md", raise error
     "invalid path: the path is not a _node.md file under code-from-spec/".

  4. If the normalized path is exactly "code-from-spec/_node.md", return "ROOT".

  5. Remove the leading "code-from-spec/" prefix and the trailing "/_node.md" suffix.
     The remaining string is the segment portion (e.g., "x/y").

  6. Prepend "ROOT/" to the segment portion.

     Examples:
       "code-from-spec/_node.md"     -> "ROOT"
       "code-from-spec/x/_node.md"   -> "ROOT/x"
       "code-from-spec/x/y/_node.md" -> "ROOT/x/y"

  7. Return the resulting logical name.

---

### function ExtractQualifier(logical_name) -> optional string

Extracts the parenthetical qualifier from a logical name, if present.
Works for both "ROOT/" and "ARTIFACT/" names.

Parameters:
  - logical_name: string

Returns:
  - optional string:
      - if a qualifier is present, return the text inside the parentheses
      - if no qualifier is present, return absent (no value)

Note: This function never raises an error. Unrecognized or malformed inputs
simply return absent.

Steps:

  1. Look for the last "(" character in logical_name.
     If not found, return absent.

  2. Look for a ")" character after the "(" found in step 1.
     If not found, return absent.

  3. If ")" is not the last character in logical_name, return absent.
     (The qualifier must close at the very end of the string.)

  4. Extract the substring between "(" and ")" (exclusive of both delimiters).

  5. If the extracted substring is empty, return absent.

  6. Return the extracted substring as the qualifier.

     Examples:
       "ROOT/x(y)"        -> "y"
       "ARTIFACT/x(id)"   -> "id"
       "ROOT/x"           -> absent
       "ROOT/x()"         -> absent  (empty qualifier)
       "ROOT/x(y)z"       -> absent  (closing paren not at end)
```
