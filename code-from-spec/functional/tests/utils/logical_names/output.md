<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@d--5kd2sQPWiVbEHf2-P-MuCtGY -->

# Test Specification: logical_names

---

## LogicalNameToPath

### TC-01: ROOT alone

Setup: none.

Action: call `LogicalNameToPath` with `"ROOT"`.

Expected outcome: returns the path `code-from-spec/_node.md`.

---

### TC-02: ROOT with path

Setup: none.

Action: call `LogicalNameToPath` with `"ROOT/payments/processor"`.

Expected outcome: returns the path `code-from-spec/payments/processor/_node.md`.

---

### TC-03: Strips qualifier before resolving

Setup: none.

Action: call `LogicalNameToPath` with `"ROOT/x/y(interface)"`.

Expected outcome: returns the path `code-from-spec/x/y/_node.md`.

---

### TC-04: Rejects ARTIFACT reference

Setup: none.

Action: call `LogicalNameToPath` with `"ARTIFACT/x(y)"`.

Expected outcome: raises error `UnsupportedReference`.

---

### TC-05: Rejects unrecognized prefix

Setup: none.

Action: call `LogicalNameToPath` with `"UNKNOWN/something"`.

Expected outcome: raises error `UnsupportedReference`.

---

### TC-06: Rejects empty string

Setup: none.

Action: call `LogicalNameToPath` with `""`.

Expected outcome: raises error `UnsupportedReference`.

---

## LogicalNameFromPath

### TC-07: Root node

Setup: none.

Action: call `LogicalNameFromPath` with path `code-from-spec/_node.md`.

Expected outcome: returns `"ROOT"`.

---

### TC-08: Nested node

Setup: none.

Action: call `LogicalNameFromPath` with path `code-from-spec/x/y/_node.md`.

Expected outcome: returns `"ROOT/x/y"`.

---

### TC-09: Rejects non-node path

Setup: none.

Action: call `LogicalNameFromPath` with path `internal/config/config.go`.

Expected outcome: raises error `InvalidPath`.

---

### TC-10: Rejects path without _node.md

Setup: none.

Action: call `LogicalNameFromPath` with path `code-from-spec/x/y/output.md`.

Expected outcome: raises error `InvalidPath`.

---

## LogicalNameGetParent

### TC-11: ROOT/x parent is ROOT

Setup: none.

Action: call `LogicalNameGetParent` with `"ROOT/domain"`.

Expected outcome: returns `"ROOT"`.

---

### TC-12: ROOT/x/y parent is ROOT/x

Setup: none.

Action: call `LogicalNameGetParent` with `"ROOT/domain/config"`.

Expected outcome: returns `"ROOT/domain"`.

---

### TC-13: Strips qualifier before computing parent

Setup: none.

Action: call `LogicalNameGetParent` with `"ROOT/domain/config(interface)"`.

Expected outcome: returns `"ROOT/domain"`.

---

### TC-14: ROOT has no parent

Setup: none.

Action: call `LogicalNameGetParent` with `"ROOT"`.

Expected outcome: raises error `NoParent`.

---

### TC-15: Rejects ARTIFACT reference

Setup: none.

Action: call `LogicalNameGetParent` with `"ARTIFACT/x(y)"`.

Expected outcome: raises error `NotARootReference`.

---

## LogicalNameGetQualifier

### TC-16: Extracts qualifier from ROOT reference

Setup: none.

Action: call `LogicalNameGetQualifier` with `"ROOT/x/y(interface)"`.

Expected outcome: returns `"interface"`.

---

### TC-17: Extracts qualifier from ARTIFACT reference

Setup: none.

Action: call `LogicalNameGetQualifier` with `"ARTIFACT/x/y(id)"`.

Expected outcome: returns `"id"`.

---

### TC-18: Returns absent when no qualifier

Setup: none.

Action: call `LogicalNameGetQualifier` with `"ROOT/x/y"`.

Expected outcome: returns absent.

---

### TC-19: Returns absent for ROOT alone

Setup: none.

Action: call `LogicalNameGetQualifier` with `"ROOT"`.

Expected outcome: returns absent.

---

## LogicalNameStripQualifier

### TC-20: Strips qualifier from ROOT reference

Setup: none.

Action: call `LogicalNameStripQualifier` with `"ROOT/x/y(interface)"`.

Expected outcome: returns `"ROOT/x/y"`.

---

### TC-21: Strips qualifier from ARTIFACT reference

Setup: none.

Action: call `LogicalNameStripQualifier` with `"ARTIFACT/x/y(id)"`.

Expected outcome: returns `"ARTIFACT/x/y"`.

---

### TC-22: No qualifier — returns unchanged

Setup: none.

Action: call `LogicalNameStripQualifier` with `"ROOT/x/y"`.

Expected outcome: returns `"ROOT/x/y"`.

---

### TC-23: ROOT alone — returns unchanged

Setup: none.

Action: call `LogicalNameStripQualifier` with `"ROOT"`.

Expected outcome: returns `"ROOT"`.

---

### TC-24: Empty string — returns unchanged

Setup: none.

Action: call `LogicalNameStripQualifier` with `""`.

Expected outcome: returns `""`.

---

## LogicalNameHasParent

### TC-25: ROOT alone

Setup: none.

Action: call `LogicalNameHasParent` with `"ROOT"`.

Expected outcome: returns false.

---

### TC-26: ROOT with path

Setup: none.

Action: call `LogicalNameHasParent` with `"ROOT/domain/config"`.

Expected outcome: returns true.

---

### TC-27: ROOT with qualifier

Setup: none.

Action: call `LogicalNameHasParent` with `"ROOT/domain/config(interface)"`.

Expected outcome: returns true.

---

### TC-28: ARTIFACT reference

Setup: none.

Action: call `LogicalNameHasParent` with `"ARTIFACT/x(y)"`.

Expected outcome: returns false.

---

### TC-29: Empty string

Setup: none.

Action: call `LogicalNameHasParent` with `""`.

Expected outcome: returns false.

---

## LogicalNameHasQualifier

### TC-30: Without qualifier

Setup: none.

Action: call `LogicalNameHasQualifier` with `"ROOT/x"`.

Expected outcome: returns false.

---

### TC-31: With qualifier

Setup: none.

Action: call `LogicalNameHasQualifier` with `"ROOT/x(y)"`.

Expected outcome: returns true.

---

### TC-32: ARTIFACT with qualifier

Setup: none.

Action: call `LogicalNameHasQualifier` with `"ARTIFACT/x(y)"`.

Expected outcome: returns true.

---

### TC-33: ROOT alone

Setup: none.

Action: call `LogicalNameHasQualifier` with `"ROOT"`.

Expected outcome: returns false.

---

### TC-34: Empty string

Setup: none.

Action: call `LogicalNameHasQualifier` with `""`.

Expected outcome: returns false.

---

## LogicalNameIsArtifact

### TC-35: ARTIFACT reference

Setup: none.

Action: call `LogicalNameIsArtifact` with `"ARTIFACT/x(y)"`.

Expected outcome: returns true.

---

### TC-36: ROOT reference

Setup: none.

Action: call `LogicalNameIsArtifact` with `"ROOT/x(y)"`.

Expected outcome: returns false.

---

### TC-37: Empty string

Setup: none.

Action: call `LogicalNameIsArtifact` with `""`.

Expected outcome: returns false.

---

## LogicalNameGetArtifactGenerator

### TC-38: Simple artifact

Setup: none.

Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x(y)"`.

Expected outcome: returns `"ROOT/x"`.

---

### TC-39: Nested artifact

Setup: none.

Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x/y/z(id)"`.

Expected outcome: returns `"ROOT/x/y/z"`.

---

### TC-40: Rejects ROOT reference

Setup: none.

Action: call `LogicalNameGetArtifactGenerator` with `"ROOT/x(y)"`.

Expected outcome: raises error `NotAnArtifactReference`.

---

### TC-41: Artifact reference without qualifier

Setup: none.

Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x"`.

Expected outcome: returns `"ROOT/x"`.
