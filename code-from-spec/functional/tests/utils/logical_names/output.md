<!-- code-from-spec: SPEC/functional/tests/utils/logical_names@KgxQvYiyLLOdJQ2h5yfcMRjZgZI -->

## Test suite: logical_names

---

### LogicalNameToPath

#### SPEC alone

Action: call LogicalNameToPath with `"SPEC"`.
Expect: returns PathCfs `code-from-spec/_node.md`.

#### SPEC with path

Action: call LogicalNameToPath with `"SPEC/payments/processor"`.
Expect: returns PathCfs `code-from-spec/payments/processor/_node.md`.

#### Strips qualifier before resolving

Action: call LogicalNameToPath with `"SPEC/x/y(interface)"`.
Expect: returns PathCfs `code-from-spec/x/y/_node.md`.

#### Rejects ROOT reference

Action: call LogicalNameToPath with `"ROOT/x"`.
Expect: raises error UnsupportedReference.

#### Rejects ARTIFACT reference

Action: call LogicalNameToPath with `"ARTIFACT/x"`.
Expect: raises error UnsupportedReference.

#### Rejects EXTERNAL reference

Action: call LogicalNameToPath with `"EXTERNAL/proto/api.proto"`.
Expect: raises error UnsupportedReference.

#### Rejects unrecognized prefix

Action: call LogicalNameToPath with `"UNKNOWN/something"`.
Expect: raises error UnsupportedReference.

#### Rejects empty string

Action: call LogicalNameToPath with `""`.
Expect: raises error UnsupportedReference.

---

### LogicalNameFromPath

#### Root node

Action: call LogicalNameFromPath with PathCfs `code-from-spec/_node.md`.
Expect: returns `"SPEC"`.

#### Nested node

Action: call LogicalNameFromPath with PathCfs `code-from-spec/x/y/_node.md`.
Expect: returns `"SPEC/x/y"`.

#### Rejects non-node path

Action: call LogicalNameFromPath with PathCfs `internal/config/config.go`.
Expect: raises error InvalidPath.

#### Rejects path without _node.md

Action: call LogicalNameFromPath with PathCfs `code-from-spec/x/y/output.md`.
Expect: raises error InvalidPath.

---

### LogicalNameGetParent

#### SPEC/x parent is SPEC

Action: call LogicalNameGetParent with `"SPEC/domain"`.
Expect: returns `"SPEC"`.

#### SPEC/x/y parent is SPEC/x

Action: call LogicalNameGetParent with `"SPEC/domain/config"`.
Expect: returns `"SPEC/domain"`.

#### Strips qualifier before computing parent

Action: call LogicalNameGetParent with `"SPEC/domain/config(interface)"`.
Expect: returns `"SPEC/domain"`.

#### SPEC has no parent

Action: call LogicalNameGetParent with `"SPEC"`.
Expect: raises error NoParent.

#### Rejects ROOT reference

Action: call LogicalNameGetParent with `"ROOT/domain"`.
Expect: raises error NotASpecReference.

#### Rejects ARTIFACT reference

Action: call LogicalNameGetParent with `"ARTIFACT/x"`.
Expect: raises error NotASpecReference.

#### Rejects EXTERNAL reference

Action: call LogicalNameGetParent with `"EXTERNAL/x"`.
Expect: raises error NotASpecReference.

---

### LogicalNameGetQualifier

#### Extracts qualifier from SPEC reference

Action: call LogicalNameGetQualifier with `"SPEC/x/y(interface)"`.
Expect: returns `"interface"`.

#### ARTIFACT without qualifier returns absent

Action: call LogicalNameGetQualifier with `"ARTIFACT/x/y"`.
Expect: returns absent.

#### EXTERNAL without qualifier returns absent

Action: call LogicalNameGetQualifier with `"EXTERNAL/proto/api.proto"`.
Expect: returns absent.

#### Returns absent when no qualifier

Action: call LogicalNameGetQualifier with `"SPEC/x/y"`.
Expect: returns absent.

#### Returns absent for SPEC alone

Action: call LogicalNameGetQualifier with `"SPEC"`.
Expect: returns absent.

---

### LogicalNameStripQualifier

#### Strips qualifier from SPEC reference

Action: call LogicalNameStripQualifier with `"SPEC/x/y(interface)"`.
Expect: returns `"SPEC/x/y"`.

#### ARTIFACT without qualifier returns unchanged

Action: call LogicalNameStripQualifier with `"ARTIFACT/x/y"`.
Expect: returns `"ARTIFACT/x/y"`.

#### EXTERNAL returns unchanged

Action: call LogicalNameStripQualifier with `"EXTERNAL/proto/api.proto"`.
Expect: returns `"EXTERNAL/proto/api.proto"`.

#### No qualifier returns unchanged

Action: call LogicalNameStripQualifier with `"SPEC/x/y"`.
Expect: returns `"SPEC/x/y"`.

#### SPEC alone returns unchanged

Action: call LogicalNameStripQualifier with `"SPEC"`.
Expect: returns `"SPEC"`.

#### Empty string returns unchanged

Action: call LogicalNameStripQualifier with `""`.
Expect: returns `""`.

---

### LogicalNameHasParent

#### SPEC alone

Action: call LogicalNameHasParent with `"SPEC"`.
Expect: returns false.

#### SPEC with path

Action: call LogicalNameHasParent with `"SPEC/domain/config"`.
Expect: returns true.

#### ARTIFACT reference

Action: call LogicalNameHasParent with `"ARTIFACT/x"`.
Expect: returns false.

#### EXTERNAL reference

Action: call LogicalNameHasParent with `"EXTERNAL/x"`.
Expect: returns false.

#### Empty string

Action: call LogicalNameHasParent with `""`.
Expect: returns false.

---

### LogicalNameHasQualifier

#### Without qualifier

Action: call LogicalNameHasQualifier with `"SPEC/x"`.
Expect: returns false.

#### With qualifier

Action: call LogicalNameHasQualifier with `"SPEC/x(y)"`.
Expect: returns true.

#### ARTIFACT without qualifier

Action: call LogicalNameHasQualifier with `"ARTIFACT/x"`.
Expect: returns false.

#### EXTERNAL without qualifier

Action: call LogicalNameHasQualifier with `"EXTERNAL/x"`.
Expect: returns false.

#### SPEC alone

Action: call LogicalNameHasQualifier with `"SPEC"`.
Expect: returns false.

#### Empty string

Action: call LogicalNameHasQualifier with `""`.
Expect: returns false.

---

### LogicalNameIsArtifact

#### ARTIFACT reference

Action: call LogicalNameIsArtifact with `"ARTIFACT/x"`.
Expect: returns true.

#### SPEC reference

Action: call LogicalNameIsArtifact with `"SPEC/x(y)"`.
Expect: returns false.

#### EXTERNAL reference

Action: call LogicalNameIsArtifact with `"EXTERNAL/x"`.
Expect: returns false.

#### Empty string

Action: call LogicalNameIsArtifact with `""`.
Expect: returns false.

---

### LogicalNameIsSpec

#### SPEC alone

Action: call LogicalNameIsSpec with `"SPEC"`.
Expect: returns true.

#### SPEC with path

Action: call LogicalNameIsSpec with `"SPEC/x/y"`.
Expect: returns true.

#### ROOT reference is not SPEC

Action: call LogicalNameIsSpec with `"ROOT/x"`.
Expect: returns false.

#### ARTIFACT reference

Action: call LogicalNameIsSpec with `"ARTIFACT/x"`.
Expect: returns false.

#### EXTERNAL reference

Action: call LogicalNameIsSpec with `"EXTERNAL/x"`.
Expect: returns false.

#### Empty string

Action: call LogicalNameIsSpec with `""`.
Expect: returns false.

---

### LogicalNameIsExternal

#### EXTERNAL reference

Action: call LogicalNameIsExternal with `"EXTERNAL/proto/api.proto"`.
Expect: returns true.

#### SPEC reference

Action: call LogicalNameIsExternal with `"SPEC/x"`.
Expect: returns false.

#### ARTIFACT reference

Action: call LogicalNameIsExternal with `"ARTIFACT/x"`.
Expect: returns false.

#### Empty string

Action: call LogicalNameIsExternal with `""`.
Expect: returns false.

---

### LogicalNameGetArtifactGenerator

#### Simple artifact

Action: call LogicalNameGetArtifactGenerator with `"ARTIFACT/x"`.
Expect: returns `"SPEC/x"`.

#### Nested artifact

Action: call LogicalNameGetArtifactGenerator with `"ARTIFACT/x/y/z"`.
Expect: returns `"SPEC/x/y/z"`.

#### Rejects SPEC reference

Action: call LogicalNameGetArtifactGenerator with `"SPEC/x(y)"`.
Expect: raises error NotAnArtifactReference.

#### Rejects EXTERNAL reference

Action: call LogicalNameGetArtifactGenerator with `"EXTERNAL/x"`.
Expect: raises error NotAnArtifactReference.

---

### LogicalNameExternalToPath

#### Simple path

Action: call LogicalNameExternalToPath with `"EXTERNAL/proto/v1/api.proto"`.
Expect: returns PathCfs `proto/v1/api.proto`.

#### Root-level file

Action: call LogicalNameExternalToPath with `"EXTERNAL/docker-compose.yaml"`.
Expect: returns PathCfs `docker-compose.yaml`.

#### Rejects SPEC reference

Action: call LogicalNameExternalToPath with `"SPEC/x"`.
Expect: raises error NotAnExternalReference.

#### Rejects ARTIFACT reference

Action: call LogicalNameExternalToPath with `"ARTIFACT/x"`.
Expect: raises error NotAnExternalReference.
