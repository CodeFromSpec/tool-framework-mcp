<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@JTj8knDgfRpOusK77S61liPkfTo -->

# Test Specification: logical_names

---

## LogicalNameToPath

---

### TC-01: ROOT alone

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `"ROOT"`.

**Expected outcome:** Returns the path `code-from-spec/_node.md`.

---

### TC-02: ROOT with path

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `"ROOT/payments/processor"`.

**Expected outcome:** Returns the path `code-from-spec/payments/processor/_node.md`.

---

### TC-03: Strips qualifier before resolving

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `"ROOT/x/y(interface)"`.

**Expected outcome:** Returns the path `code-from-spec/x/y/_node.md`.

---

### TC-04: Rejects ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Raises error `"unsupported reference"`.

---

### TC-05: Rejects unrecognized prefix

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `"UNKNOWN/something"`.

**Expected outcome:** Raises error `"unsupported reference"`.

---

### TC-06: Rejects empty string

**Setup:** none

**Action:** Call `LogicalNameToPath` with input `""`.

**Expected outcome:** Raises error `"unsupported reference"`.

---

## LogicalNameFromPath

---

### TC-07: Root node

**Setup:** none

**Action:** Call `LogicalNameFromPath` with path `code-from-spec/_node.md`.

**Expected outcome:** Returns `"ROOT"`.

---

### TC-08: Nested node

**Setup:** none

**Action:** Call `LogicalNameFromPath` with path `code-from-spec/x/y/_node.md`.

**Expected outcome:** Returns `"ROOT/x/y"`.

---

### TC-09: Rejects non-node path

**Setup:** none

**Action:** Call `LogicalNameFromPath` with path `internal/config/config.go`.

**Expected outcome:** Raises error `"invalid path"`.

---

### TC-10: Rejects path without _node.md

**Setup:** none

**Action:** Call `LogicalNameFromPath` with path `code-from-spec/x/y/output.md`.

**Expected outcome:** Raises error `"invalid path"`.

---

## LogicalNameGetParent

---

### TC-11: ROOT/x parent is ROOT

**Setup:** none

**Action:** Call `LogicalNameGetParent` with input `"ROOT/domain"`.

**Expected outcome:** Returns `"ROOT"`.

---

### TC-12: ROOT/x/y parent is ROOT/x

**Setup:** none

**Action:** Call `LogicalNameGetParent` with input `"ROOT/domain/config"`.

**Expected outcome:** Returns `"ROOT/domain"`.

---

### TC-13: Strips qualifier before computing parent

**Setup:** none

**Action:** Call `LogicalNameGetParent` with input `"ROOT/domain/config(interface)"`.

**Expected outcome:** Returns `"ROOT/domain"`.

---

### TC-14: ROOT has no parent

**Setup:** none

**Action:** Call `LogicalNameGetParent` with input `"ROOT"`.

**Expected outcome:** Raises error `"no parent"`.

---

### TC-15: Rejects ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameGetParent` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Raises error `"not a ROOT reference"`.

---

## LogicalNameGetQualifier

---

### TC-16: Extracts qualifier from ROOT reference

**Setup:** none

**Action:** Call `LogicalNameGetQualifier` with input `"ROOT/x/y(interface)"`.

**Expected outcome:** Returns `"interface"`.

---

### TC-17: Extracts qualifier from ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameGetQualifier` with input `"ARTIFACT/x/y(id)"`.

**Expected outcome:** Returns `"id"`.

---

### TC-18: Returns absent when no qualifier

**Setup:** none

**Action:** Call `LogicalNameGetQualifier` with input `"ROOT/x/y"`.

**Expected outcome:** Returns absent.

---

### TC-19: Returns absent for ROOT alone

**Setup:** none

**Action:** Call `LogicalNameGetQualifier` with input `"ROOT"`.

**Expected outcome:** Returns absent.

---

## LogicalNameStripQualifier

---

### TC-20: Strips qualifier from ROOT reference

**Setup:** none

**Action:** Call `LogicalNameStripQualifier` with input `"ROOT/x/y(interface)"`.

**Expected outcome:** Returns `"ROOT/x/y"`.

---

### TC-21: Strips qualifier from ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameStripQualifier` with input `"ARTIFACT/x/y(id)"`.

**Expected outcome:** Returns `"ARTIFACT/x/y"`.

---

### TC-22: No qualifier — returns unchanged

**Setup:** none

**Action:** Call `LogicalNameStripQualifier` with input `"ROOT/x/y"`.

**Expected outcome:** Returns `"ROOT/x/y"`.

---

### TC-23: ROOT alone — returns unchanged

**Setup:** none

**Action:** Call `LogicalNameStripQualifier` with input `"ROOT"`.

**Expected outcome:** Returns `"ROOT"`.

---

### TC-24: Empty string — returns unchanged

**Setup:** none

**Action:** Call `LogicalNameStripQualifier` with input `""`.

**Expected outcome:** Returns `""`.

---

## LogicalNameHasParent

---

### TC-25: ROOT alone

**Setup:** none

**Action:** Call `LogicalNameHasParent` with input `"ROOT"`.

**Expected outcome:** Returns false.

---

### TC-26: ROOT with path

**Setup:** none

**Action:** Call `LogicalNameHasParent` with input `"ROOT/domain/config"`.

**Expected outcome:** Returns true.

---

### TC-27: ROOT with qualifier

**Setup:** none

**Action:** Call `LogicalNameHasParent` with input `"ROOT/domain/config(interface)"`.

**Expected outcome:** Returns true.

---

### TC-28: ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameHasParent` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Returns false.

---

### TC-29: Empty string

**Setup:** none

**Action:** Call `LogicalNameHasParent` with input `""`.

**Expected outcome:** Returns false.

---

## LogicalNameHasQualifier

---

### TC-30: Without qualifier

**Setup:** none

**Action:** Call `LogicalNameHasQualifier` with input `"ROOT/x"`.

**Expected outcome:** Returns false.

---

### TC-31: With qualifier

**Setup:** none

**Action:** Call `LogicalNameHasQualifier` with input `"ROOT/x(y)"`.

**Expected outcome:** Returns true.

---

### TC-32: ARTIFACT with qualifier

**Setup:** none

**Action:** Call `LogicalNameHasQualifier` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Returns true.

---

### TC-33: ROOT alone

**Setup:** none

**Action:** Call `LogicalNameHasQualifier` with input `"ROOT"`.

**Expected outcome:** Returns false.

---

### TC-34: Empty string

**Setup:** none

**Action:** Call `LogicalNameHasQualifier` with input `""`.

**Expected outcome:** Returns false.

---

## LogicalNameIsArtifact

---

### TC-35: ARTIFACT reference

**Setup:** none

**Action:** Call `LogicalNameIsArtifact` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Returns true.

---

### TC-36: ROOT reference

**Setup:** none

**Action:** Call `LogicalNameIsArtifact` with input `"ROOT/x(y)"`.

**Expected outcome:** Returns false.

---

### TC-37: Empty string

**Setup:** none

**Action:** Call `LogicalNameIsArtifact` with input `""`.

**Expected outcome:** Returns false.

---

## LogicalNameGetArtifactGenerator

---

### TC-38: Simple artifact

**Setup:** none

**Action:** Call `LogicalNameGetArtifactGenerator` with input `"ARTIFACT/x(y)"`.

**Expected outcome:** Returns `"ROOT/x"`.

---

### TC-39: Nested artifact

**Setup:** none

**Action:** Call `LogicalNameGetArtifactGenerator` with input `"ARTIFACT/x/y/z(id)"`.

**Expected outcome:** Returns `"ROOT/x/y/z"`.

---

### TC-40: Rejects ROOT reference

**Setup:** none

**Action:** Call `LogicalNameGetArtifactGenerator` with input `"ROOT/x(y)"`.

**Expected outcome:** Raises error `"not an artifact reference"`.

---

### TC-41: Artifact reference without qualifier

**Setup:** none

**Action:** Call `LogicalNameGetArtifactGenerator` with input `"ARTIFACT/x"`.

**Expected outcome:** Returns `"ROOT/x"`.
