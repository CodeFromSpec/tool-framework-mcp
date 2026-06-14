<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@HNyT4_9XAbVSZrY-hCEHolzXiJs -->

## Test suite: logical_names

---

### LogicalNameToPath

#### Test: SPEC alone

Actions:
1. Call LogicalNameToPath with `"SPEC"`.

Expected outcome:
- Returns PathCfs `code-from-spec/_node.md`.

---

#### Test: SPEC with path

Actions:
1. Call LogicalNameToPath with `"SPEC/payments/processor"`.

Expected outcome:
- Returns PathCfs `code-from-spec/payments/processor/_node.md`.

---

#### Test: Strips qualifier before resolving

Actions:
1. Call LogicalNameToPath with `"SPEC/x/y(interface)"`.

Expected outcome:
- Returns PathCfs `code-from-spec/x/y/_node.md`.

---

#### Test: Rejects ROOT reference

Actions:
1. Call LogicalNameToPath with `"ROOT/x"`.

Expected outcome:
- Raises error UnsupportedReference.

---

#### Test: Rejects ARTIFACT reference

Actions:
1. Call LogicalNameToPath with `"ARTIFACT/x"`.

Expected outcome:
- Raises error UnsupportedReference.

---

#### Test: Rejects EXTERNAL reference

Actions:
1. Call LogicalNameToPath with `"EXTERNAL/proto/api.proto"`.

Expected outcome:
- Raises error UnsupportedReference.

---

#### Test: Rejects unrecognized prefix

Actions:
1. Call LogicalNameToPath with `"UNKNOWN/something"`.

Expected outcome:
- Raises error UnsupportedReference.

---

#### Test: Rejects empty string

Actions:
1. Call LogicalNameToPath with `""`.

Expected outcome:
- Raises error UnsupportedReference.

---

### LogicalNameFromPath

#### Test: Root node

Actions:
1. Call LogicalNameFromPath with PathCfs `code-from-spec/_node.md`.

Expected outcome:
- Returns `"SPEC"`.

---

#### Test: Nested node

Actions:
1. Call LogicalNameFromPath with PathCfs `code-from-spec/x/y/_node.md`.

Expected outcome:
- Returns `"SPEC/x/y"`.

---

#### Test: Rejects non-node path

Actions:
1. Call LogicalNameFromPath with PathCfs `internal/config/config.go`.

Expected outcome:
- Raises error InvalidPath.

---

#### Test: Rejects path without _node.md

Actions:
1. Call LogicalNameFromPath with PathCfs `code-from-spec/x/y/output.md`.

Expected outcome:
- Raises error InvalidPath.

---

### LogicalNameGetParent

#### Test: SPEC/x parent is SPEC

Actions:
1. Call LogicalNameGetParent with `"SPEC/domain"`.

Expected outcome:
- Returns `"SPEC"`.

---

#### Test: SPEC/x/y parent is SPEC/x

Actions:
1. Call LogicalNameGetParent with `"SPEC/domain/config"`.

Expected outcome:
- Returns `"SPEC/domain"`.

---

#### Test: Strips qualifier before computing parent

Actions:
1. Call LogicalNameGetParent with `"SPEC/domain/config(interface)"`.

Expected outcome:
- Returns `"SPEC/domain"`.

---

#### Test: SPEC has no parent

Actions:
1. Call LogicalNameGetParent with `"SPEC"`.

Expected outcome:
- Raises error NoParent.

---

#### Test: Rejects ROOT reference

Actions:
1. Call LogicalNameGetParent with `"ROOT/domain"`.

Expected outcome:
- Raises error NotASpecReference.

---

#### Test: Rejects ARTIFACT reference

Actions:
1. Call LogicalNameGetParent with `"ARTIFACT/x"`.

Expected outcome:
- Raises error NotASpecReference.

---

#### Test: Rejects EXTERNAL reference

Actions:
1. Call LogicalNameGetParent with `"EXTERNAL/x"`.

Expected outcome:
- Raises error NotASpecReference.

---

### LogicalNameGetQualifier

#### Test: Extracts qualifier from SPEC reference

Actions:
1. Call LogicalNameGetQualifier with `"SPEC/x/y(interface)"`.

Expected outcome:
- Returns `"interface"`.

---

#### Test: ARTIFACT without qualifier returns absent

Actions:
1. Call LogicalNameGetQualifier with `"ARTIFACT/x/y"`.

Expected outcome:
- Returns absent.

---

#### Test: EXTERNAL without qualifier returns absent

Actions:
1. Call LogicalNameGetQualifier with `"EXTERNAL/proto/api.proto"`.

Expected outcome:
- Returns absent.

---

#### Test: Returns absent when no qualifier

Actions:
1. Call LogicalNameGetQualifier with `"SPEC/x/y"`.

Expected outcome:
- Returns absent.

---

#### Test: Returns absent for SPEC alone

Actions:
1. Call LogicalNameGetQualifier with `"SPEC"`.

Expected outcome:
- Returns absent.

---

### LogicalNameStripQualifier

#### Test: Strips qualifier from SPEC reference

Actions:
1. Call LogicalNameStripQualifier with `"SPEC/x/y(interface)"`.

Expected outcome:
- Returns `"SPEC/x/y"`.

---

#### Test: ARTIFACT without qualifier returns unchanged

Actions:
1. Call LogicalNameStripQualifier with `"ARTIFACT/x/y"`.

Expected outcome:
- Returns `"ARTIFACT/x/y"`.

---

#### Test: EXTERNAL returns unchanged

Actions:
1. Call LogicalNameStripQualifier with `"EXTERNAL/proto/api.proto"`.

Expected outcome:
- Returns `"EXTERNAL/proto/api.proto"`.

---

#### Test: No qualifier returns unchanged

Actions:
1. Call LogicalNameStripQualifier with `"SPEC/x/y"`.

Expected outcome:
- Returns `"SPEC/x/y"`.

---

#### Test: SPEC alone returns unchanged

Actions:
1. Call LogicalNameStripQualifier with `"SPEC"`.

Expected outcome:
- Returns `"SPEC"`.

---

#### Test: Empty string returns unchanged

Actions:
1. Call LogicalNameStripQualifier with `""`.

Expected outcome:
- Returns `""`.

---

### LogicalNameHasParent

#### Test: SPEC alone

Actions:
1. Call LogicalNameHasParent with `"SPEC"`.

Expected outcome:
- Returns false.

---

#### Test: SPEC with path

Actions:
1. Call LogicalNameHasParent with `"SPEC/domain/config"`.

Expected outcome:
- Returns true.

---

#### Test: ARTIFACT reference

Actions:
1. Call LogicalNameHasParent with `"ARTIFACT/x"`.

Expected outcome:
- Returns false.

---

#### Test: EXTERNAL reference

Actions:
1. Call LogicalNameHasParent with `"EXTERNAL/x"`.

Expected outcome:
- Returns false.

---

#### Test: Empty string

Actions:
1. Call LogicalNameHasParent with `""`.

Expected outcome:
- Returns false.

---

### LogicalNameHasQualifier

#### Test: Without qualifier

Actions:
1. Call LogicalNameHasQualifier with `"SPEC/x"`.

Expected outcome:
- Returns false.

---

#### Test: With qualifier

Actions:
1. Call LogicalNameHasQualifier with `"SPEC/x(y)"`.

Expected outcome:
- Returns true.

---

#### Test: ARTIFACT without qualifier

Actions:
1. Call LogicalNameHasQualifier with `"ARTIFACT/x"`.

Expected outcome:
- Returns false.

---

#### Test: EXTERNAL without qualifier

Actions:
1. Call LogicalNameHasQualifier with `"EXTERNAL/x"`.

Expected outcome:
- Returns false.

---

#### Test: SPEC alone

Actions:
1. Call LogicalNameHasQualifier with `"SPEC"`.

Expected outcome:
- Returns false.

---

#### Test: Empty string

Actions:
1. Call LogicalNameHasQualifier with `""`.

Expected outcome:
- Returns false.

---

### LogicalNameIsArtifact

#### Test: ARTIFACT reference

Actions:
1. Call LogicalNameIsArtifact with `"ARTIFACT/x"`.

Expected outcome:
- Returns true.

---

#### Test: SPEC reference

Actions:
1. Call LogicalNameIsArtifact with `"SPEC/x(y)"`.

Expected outcome:
- Returns false.

---

#### Test: EXTERNAL reference

Actions:
1. Call LogicalNameIsArtifact with `"EXTERNAL/x"`.

Expected outcome:
- Returns false.

---

#### Test: Empty string

Actions:
1. Call LogicalNameIsArtifact with `""`.

Expected outcome:
- Returns false.

---

### LogicalNameIsSpec

#### Test: SPEC alone

Actions:
1. Call LogicalNameIsSpec with `"SPEC"`.

Expected outcome:
- Returns true.

---

#### Test: SPEC with path

Actions:
1. Call LogicalNameIsSpec with `"SPEC/x/y"`.

Expected outcome:
- Returns true.

---

#### Test: ROOT reference — not SPEC

Actions:
1. Call LogicalNameIsSpec with `"ROOT/x"`.

Expected outcome:
- Returns false.

---

#### Test: ARTIFACT reference

Actions:
1. Call LogicalNameIsSpec with `"ARTIFACT/x"`.

Expected outcome:
- Returns false.

---

#### Test: EXTERNAL reference

Actions:
1. Call LogicalNameIsSpec with `"EXTERNAL/x"`.

Expected outcome:
- Returns false.

---

#### Test: Empty string

Actions:
1. Call LogicalNameIsSpec with `""`.

Expected outcome:
- Returns false.

---

### LogicalNameIsExternal

#### Test: EXTERNAL reference

Actions:
1. Call LogicalNameIsExternal with `"EXTERNAL/proto/api.proto"`.

Expected outcome:
- Returns true.

---

#### Test: SPEC reference

Actions:
1. Call LogicalNameIsExternal with `"SPEC/x"`.

Expected outcome:
- Returns false.

---

#### Test: ARTIFACT reference

Actions:
1. Call LogicalNameIsExternal with `"ARTIFACT/x"`.

Expected outcome:
- Returns false.

---

#### Test: Empty string

Actions:
1. Call LogicalNameIsExternal with `""`.

Expected outcome:
- Returns false.

---

### LogicalNameGetArtifactGenerator

#### Test: Simple artifact

Actions:
1. Call LogicalNameGetArtifactGenerator with `"ARTIFACT/x"`.

Expected outcome:
- Returns `"SPEC/x"`.

---

#### Test: Nested artifact

Actions:
1. Call LogicalNameGetArtifactGenerator with `"ARTIFACT/x/y/z"`.

Expected outcome:
- Returns `"SPEC/x/y/z"`.

---

#### Test: Rejects SPEC reference

Actions:
1. Call LogicalNameGetArtifactGenerator with `"SPEC/x(y)"`.

Expected outcome:
- Raises error NotAnArtifactReference.

---

#### Test: Rejects EXTERNAL reference

Actions:
1. Call LogicalNameGetArtifactGenerator with `"EXTERNAL/x"`.

Expected outcome:
- Raises error NotAnArtifactReference.

---

### LogicalNameExternalToPath

#### Test: Simple path

Actions:
1. Call LogicalNameExternalToPath with `"EXTERNAL/proto/v1/api.proto"`.

Expected outcome:
- Returns PathCfs `proto/v1/api.proto`.

---

#### Test: Root-level file

Actions:
1. Call LogicalNameExternalToPath with `"EXTERNAL/docker-compose.yaml"`.

Expected outcome:
- Returns PathCfs `docker-compose.yaml`.

---

#### Test: Rejects SPEC reference

Actions:
1. Call LogicalNameExternalToPath with `"SPEC/x"`.

Expected outcome:
- Raises error NotAnExternalReference.

---

#### Test: Rejects ARTIFACT reference

Actions:
1. Call LogicalNameExternalToPath with `"ARTIFACT/x"`.

Expected outcome:
- Raises error NotAnExternalReference.
