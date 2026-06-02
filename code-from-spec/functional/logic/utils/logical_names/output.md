<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@SqBOYKAbQLc8vwQ5EGf3U2UtNHo -->

function LogicalNameToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - UnsupportedReference: the logical name is not a ROOT/ reference.

  1. Call LogicalNameStripQualifier(logical_name) to get stripped_name.

  2. If stripped_name does not start with "ROOT/" and is not exactly "ROOT",
     raise error "UnsupportedReference".

  3. If stripped_name is exactly "ROOT",
     return PathCfs with value "code-from-spec/_node.md".

  4. Remove the leading "ROOT/" prefix from stripped_name to get rel_path.

  5. Return PathCfs with value "code-from-spec/<rel_path>/_node.md".


function LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string
  errors:
    - InvalidPath: the path is not a _node.md file under code-from-spec/.

  1. Let path_value be cfs_path.value.

  2. If path_value does not start with "code-from-spec/",
     raise error "InvalidPath".

  3. If path_value does not end with "/_node.md" and is not exactly "code-from-spec/_node.md",
     raise error "InvalidPath".

  4. If path_value is exactly "code-from-spec/_node.md",
     return "ROOT".

  5. Remove the leading "code-from-spec/" prefix and the trailing "/_node.md" suffix
     from path_value to get rel_path.

  6. Return "ROOT/<rel_path>".


function LogicalNameGetParent(logical_name: string) -> string
  errors:
    - NoParent: the logical name is ROOT itself.
    - NotARootReference: the logical name is not a ROOT/ reference.

  1. Call LogicalNameStripQualifier(logical_name) to get stripped_name.

  2. If stripped_name does not start with "ROOT/" and is not exactly "ROOT",
     raise error "NotARootReference".

  3. If stripped_name is exactly "ROOT",
     raise error "NoParent".

  4. Remove the leading "ROOT/" prefix from stripped_name to get rel_path.

  5. Find the last "/" in rel_path.
     If not found, return "ROOT".

  6. Take the portion of rel_path before the last "/" to get parent_rel.

  7. Return "ROOT/<parent_rel>".


function LogicalNameGetQualifier(logical_name: string) -> optional string

  1. Find the last "(" in logical_name.
     If not found, return absent.

  2. Find the ")" that closes it — look for ")" after the position of "(".
     If not found, return absent.

  3. Extract the substring between "(" and ")" to get qualifier.

  4. If qualifier is empty, return absent.

  5. Return qualifier.


function LogicalNameStripQualifier(logical_name: string) -> string

  1. Find the last "(" in logical_name.
     If not found, return logical_name unchanged.

  2. Find the ")" that follows it.
     If not found, return logical_name unchanged.

  3. Return the portion of logical_name before the "(".


function LogicalNameHasParent(logical_name: string) -> boolean

  1. Call LogicalNameStripQualifier(logical_name) to get stripped_name.

  2. If stripped_name does not start with "ROOT/",
     return false.

  3. Return true.


function LogicalNameHasQualifier(logical_name: string) -> boolean

  1. Call LogicalNameGetQualifier(logical_name) to get result.

  2. If result is absent, return false.

  3. Return true.


function LogicalNameIsArtifact(logical_name: string) -> boolean

  1. If logical_name starts with "ARTIFACT/", return true.

  2. Return false.


function LogicalNameGetArtifactGenerator(logical_name: string) -> string
  errors:
    - NotAnArtifactReference: the logical name does not start with ARTIFACT/.

  1. If logical_name does not start with "ARTIFACT/",
     raise error "NotAnArtifactReference".

  2. Remove the leading "ARTIFACT/" prefix from logical_name to get rel_path.

  3. Return "ROOT/<rel_path>".
