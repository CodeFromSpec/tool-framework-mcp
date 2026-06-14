<!-- code-from-spec: ROOT/functional/logic/utils/logical_names@yYxk_riqK8gCTamx67MIcrSATC4 -->

function LogicalNameToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - UnsupportedReference: the logical name is not a SPEC/ reference

  1. Strip any qualifier from logical_name using LogicalNameStripQualifier.

  2. If the stripped name is exactly "SPEC":
       return PathCfs with value "code-from-spec/_node.md".

  3. If the stripped name starts with "SPEC/":
       Extract the part after "SPEC/" as <relative>.
       Return PathCfs with value "code-from-spec/<relative>/_node.md".

  4. Otherwise, raise error "UnsupportedReference".


function LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string
  errors:
    - InvalidPath: the path is not a _node.md file under code-from-spec/

  1. Let <path_value> be cfs_path.value.

  2. If <path_value> does not end with "_node.md":
       raise error "InvalidPath".

  3. If <path_value> is exactly "code-from-spec/_node.md":
       return "SPEC".

  4. If <path_value> starts with "code-from-spec/" and ends with "/_node.md":
       Extract the part between "code-from-spec/" and "/_node.md" as <relative>.
       Return "SPEC/<relative>".

  5. Otherwise, raise error "InvalidPath".


function LogicalNameGetParent(logical_name: string) -> string
  errors:
    - NoParent: the logical name is SPEC itself
    - NotASpecReference: the logical name is not a SPEC/ reference

  1. Strip any qualifier from logical_name using LogicalNameStripQualifier.
     Let <stripped> be the result.

  2. If <stripped> is not exactly "SPEC" and does not start with "SPEC/":
       raise error "NotASpecReference".

  3. If <stripped> is exactly "SPEC":
       raise error "NoParent".

  4. Extract the part after "SPEC/" as <relative>.

  5. Find the last "/" in <relative>.
     If no "/" is found:
       return "SPEC".
     Else:
       Let <parent_relative> be everything in <relative> before the last "/".
       Return "SPEC/<parent_relative>".


function LogicalNameGetQualifier(logical_name: string) -> optional string

  1. Find the last "(" in logical_name.
     If not found, return absent.

  2. Find the ")" that follows the last "(".
     If not found, return absent.

  3. Extract the text between "(" and ")" as <qualifier>.

  4. If ")" is not the last character of logical_name, return absent.

  5. Return <qualifier>.


function LogicalNameStripQualifier(logical_name: string) -> string

  1. Find the last "(" in logical_name.
     If not found, return logical_name unchanged.

  2. Check that logical_name ends with ")".
     If it does not, return logical_name unchanged.

  3. Return everything before the last "(".


function LogicalNameHasParent(logical_name: string) -> boolean

  1. Strip any qualifier using LogicalNameStripQualifier.
     Let <stripped> be the result.

  2. If <stripped> starts with "SPEC/" and the part after "SPEC/" is non-empty:
       return true.

  3. Return false.


function LogicalNameHasQualifier(logical_name: string) -> boolean

  1. Call LogicalNameGetQualifier(logical_name).
     If the result is absent, return false.
     Otherwise return true.


function LogicalNameIsArtifact(logical_name: string) -> boolean

  1. If logical_name starts with "ARTIFACT/":
       return true.
  2. Return false.


function LogicalNameIsSpec(logical_name: string) -> boolean

  1. If logical_name is exactly "SPEC" or starts with "SPEC/":
       return true.
  2. Return false.


function LogicalNameIsExternal(logical_name: string) -> boolean

  1. If logical_name starts with "EXTERNAL/":
       return true.
  2. Return false.


function LogicalNameGetArtifactGenerator(logical_name: string) -> string
  errors:
    - NotAnArtifactReference: the logical name does not start with ARTIFACT/

  1. If logical_name does not start with "ARTIFACT/":
       raise error "NotAnArtifactReference".

  2. Extract the part after "ARTIFACT/" as <relative>.

  3. Return "SPEC/<relative>".


function LogicalNameExternalToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - NotAnExternalReference: the logical name does not start with EXTERNAL/

  1. If logical_name does not start with "EXTERNAL/":
       raise error "NotAnExternalReference".

  2. Extract the part after "EXTERNAL/" as <relative>.

  3. Return PathCfs with value <relative>.
