<!-- code-from-spec: ROOT/functional/utils/logical_names@PENDING -->

record ArtifactReference
  node_path: string
  artifact_id: string

function ResolvePath(logical_name) -> string

  1. If logical_name starts with "ARTIFACT/",
     raise error "unsupported reference".

  2. Strip the parenthetical qualifier if present.
     If logical_name contains "(", take the substring before
     the first "(" as the base name. Otherwise the base name
     is logical_name itself.

  3. If the base name is exactly "ROOT",
     return "code-from-spec/_node.md".

  4. Remove the "ROOT/" prefix from the base name to get the
     relative segment.

  5. Return "code-from-spec/" + relative segment + "/_node.md".

function ResolveArtifactReference(logical_name) -> ArtifactReference

  1. If logical_name does not start with "ARTIFACT/",
     raise error "unrecognized prefix".

  2. If logical_name does not contain "(", raise error
     "missing qualifier".

  3. Extract the qualifier: the text between the last "("
     and the closing ")". This is the artifact_id.

  4. Extract the node path: take the substring after
     "ARTIFACT/" and before the last "(". Prepend "ROOT/"
     to form the full node logical name. This is the node_path.

  5. Return an ArtifactReference with node_path and artifact_id.

function GetParent(logical_name) -> string

  1. If logical_name starts with "ARTIFACT/",
     raise error "not a ROOT reference".

  2. Strip the parenthetical qualifier if present.
     If logical_name contains "(", take the substring before
     the first "(" as the base name. Otherwise the base name
     is logical_name itself.

  3. If the base name is exactly "ROOT",
     raise error "no parent".

  4. Find the last "/" in the base name.
     Return the substring up to and not including that "/".

function ReverseResolve(file_path) -> string

  1. Normalize the file_path to use forward slashes.

  2. If the file_path does not end with "/_node.md" or is not
     under "code-from-spec/", raise error "invalid path".

  3. Remove the trailing "/_node.md" from the path.

  4. Remove the leading "code-from-spec/" prefix to get the
     relative segment.

  5. If the relative segment is empty (the path was
     "code-from-spec/_node.md"), return "ROOT".

  6. Return "ROOT/" + relative segment.

function ExtractQualifier(logical_name) -> optional string

  1. If logical_name contains "(":
     extract the text between the last "(" and the closing ")".
     Return that text.

  2. Otherwise, return nothing (no qualifier present).
