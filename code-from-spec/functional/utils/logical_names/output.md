<!-- code-from-spec: ROOT/functional/utils/logical_names@H5zBporTd2Jjlm03M3c5_FNa8zw -->

# Functional Spec: Logical Name Utilities

All functions are pure string manipulation — no I/O, no side effects.
All returned file paths use forward slashes as separators regardless of OS.


## record ArtifactReference

Fields:
  - node_path: string   — the path to the _node.md file for the referenced node
  - artifact_id: string — the artifact id within that node


## function ResolvePath(logical_name) -> string

Resolves a ROOT/ logical name to its corresponding _node.md file path.

Errors:
  - "unsupported reference: only ROOT/ logical names can be resolved to a path"
    when logical_name does not start with "ROOT/"

Steps:

  1. If logical_name does not start with "ROOT/", raise error
     "unsupported reference: only ROOT/ logical names can be resolved to a path".
     Note: the bare "ROOT" case is handled by step 3 below.

  2. Strip any parenthetical qualifier from logical_name.
     A qualifier is the last "(…)" segment appended to the name.
     Example: "ROOT/x/y(z)" → "ROOT/x/y".

  3. Map the stripped name to a file path:
     - If the stripped name is exactly "ROOT", return "code-from-spec/_node.md".
     - Otherwise remove the leading "ROOT/" prefix, then replace every "/"
       separator with "/" (already forward-slash), and append "/_node.md".
       Prepend "code-from-spec/".
       Example: "ROOT/x/y" → "code-from-spec/x/y/_node.md".

  4. Return the resulting path string.


## function ResolveArtifactReference(logical_name) -> ArtifactReference

Parses an ARTIFACT/ logical name into its node path and artifact id.

Errors:
  - "unrecognized prefix: the logical name does not start with ARTIFACT/"
    when logical_name does not start with "ARTIFACT/".
  - "missing qualifier: the logical name has no parenthetical qualifier"
    when no "(…)" qualifier is present.

Steps:

  1. If logical_name does not start with "ARTIFACT/", raise error
     "unrecognized prefix: the logical name does not start with ARTIFACT/".

  2. Extract the qualifier from logical_name using ExtractQualifier.
     If ExtractQualifier returns no qualifier, raise error
     "missing qualifier: the logical name has no parenthetical qualifier".

  3. Strip the qualifier (including the surrounding parentheses) from
     logical_name to get the bare node name.
     Example: "ARTIFACT/x/y(id)" → "ARTIFACT/x/y".

  4. Remove the leading "ARTIFACT/" prefix to get the relative node segment.
     Example: "ARTIFACT/x/y" → "x/y".

  5. Build the node_path by prepending "code-from-spec/" and appending
     "/_node.md" with forward slashes.
     Example: "x/y" → "code-from-spec/x/y/_node.md".

  6. Return an ArtifactReference record with:
     - node_path set to the value from step 5.
     - artifact_id set to the qualifier from step 2.


## function GetParent(logical_name) -> string

Returns the parent logical name of a ROOT/ node.

Errors:
  - "no parent: the logical name is ROOT itself"
    when logical_name (after qualifier stripping) is exactly "ROOT".
  - "not a ROOT reference: the logical name is an ARTIFACT/ reference"
    when logical_name starts with "ARTIFACT/".

Steps:

  1. If logical_name starts with "ARTIFACT/", raise error
     "not a ROOT reference: the logical name is an ARTIFACT/ reference".

  2. Strip any parenthetical qualifier from logical_name.
     Example: "ROOT/x/y(z)" → "ROOT/x/y".

  3. If the stripped name is exactly "ROOT", raise error
     "no parent: the logical name is ROOT itself".

  4. Find the position of the last "/" in the stripped name.

  5. Take the substring up to (but not including) that last "/".
     Example: "ROOT/x/y" → last "/" is before "y" → parent is "ROOT/x".
     Example: "ROOT/x"   → last "/" is before "x" → parent is "ROOT".

  6. Return the parent string.


## function ReverseResolve(file_path) -> string

Derives the logical name for a _node.md file path.

Errors:
  - "invalid path: the path is not a _node.md file under code-from-spec/"
    when the path does not match the expected structure.

Steps:

  1. Normalize the path to use forward slashes.
     Replace any backslash characters with forward slashes.

  2. Check that the path ends with "/_node.md" or is exactly
     "code-from-spec/_node.md".
     Also check that the path starts with "code-from-spec/".
     If either check fails, raise error
     "invalid path: the path is not a _node.md file under code-from-spec/".

  3. If the normalized path is exactly "code-from-spec/_node.md",
     return "ROOT".

  4. Remove the leading "code-from-spec/" prefix and the trailing
     "/_node.md" suffix.
     The remainder is the relative node segment (e.g., "x/y").

  5. Prepend "ROOT/" to the relative segment and return.
     Example: "x/y" → "ROOT/x/y".


## function ExtractQualifier(logical_name) -> optional string

Returns the qualifier inside the last parenthetical "(…)" of a logical name,
or no value if none is present.

Steps:

  1. Search logical_name for the last occurrence of "(".

  2. If no "(" is found, return no value (qualifier absent).

  3. Check that the name ends with ")".
     If it does not end with ")", return no value (malformed, treated as absent).

  4. Extract the substring between the last "(" and the final ")".
     If that substring is empty, return no value.

  5. Return the extracted substring as the qualifier.
