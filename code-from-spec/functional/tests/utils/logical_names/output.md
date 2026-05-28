<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@mqwXRyFfJ9E_52GQRw-hZd3-PxA -->

# Test Specification: logical_names utils

---

## LogicalNameToPath

### TC-01: ROOT alone

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT"`
- Expected outcome: returns path `code-from-spec/_node.md`

### TC-02: ROOT with path

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT/payments/processor"`
- Expected outcome: returns path `code-from-spec/payments/processor/_node.md`

### TC-03: Strips qualifier before resolving

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT/x/y(interface)"`
- Expected outcome: returns path `code-from-spec/x/y/_node.md`

### TC-04: Rejects ARTIFACT reference

- Setup: none
- Action: call `LogicalNameToPath` with `"ARTIFACT/x(y)"`
- Expected outcome: raises error `"unsupported reference"`

### TC-05: Rejects unrecognized prefix

- Setup: none
- Action: call `LogicalNameToPath` with `"UNKNOWN/something"`
- Expected outcome: raises error `"unsupported reference"`

### TC-06: Rejects empty string

- Setup: none
- Action: call `LogicalNameToPath` with `""`
- Expected outcome: raises error `"unsupported reference"`

---

## LogicalNameFromPath

### TC-07: Root node

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/_node.md`
- Expected outcome: returns `"ROOT"`

### TC-08: Nested node

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/x/y/_node.md`
- Expected outcome: returns `"ROOT/x/y"`

### TC-09: Rejects non-node path

- Setup: none
- Action: call `LogicalNameFromPath` with `internal/config/config.go`
- Expected outcome: raises error `"invalid path"`

### TC-10: Rejects path without _node.md

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/x/y/output.md`
- Expected outcome: raises error `"invalid path"`

---

## LogicalNameGetParent

### TC-11: ROOT/x parent is ROOT

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain"`
- Expected outcome: returns `"ROOT"`

### TC-12: ROOT/x/y parent is ROOT/x

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain/config"`
- Expected outcome: returns `"ROOT/domain"`

### TC-13: Strips qualifier before computing parent

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain/config(interface)"`
- Expected outcome: returns `"ROOT/domain"`

### TC-14: ROOT has no parent

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT"`
- Expected outcome: raises error `"no parent"`

### TC-15: Rejects ARTIFACT reference

- Setup: none
- Action: call `LogicalNameGetParent` with `"ARTIFACT/x(y)"`
- Expected outcome: raises error `"not a ROOT reference"`

---

## LogicalNameGetQualifier

### TC-16: Extracts qualifier from ROOT reference

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT/x/y(interface)"`
- Expected outcome: returns `"interface"`

### TC-17: Extracts qualifier from ARTIFACT reference

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ARTIFACT/x/y(id)"`
- Expected outcome: returns `"id"`

### TC-18: Returns absent when no qualifier

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT/x/y"`
- Expected outcome: returns absent

### TC-19: Returns absent for ROOT alone

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT"`
- Expected outcome: returns absent

---

## LogicalNameHasParent

### TC-20: ROOT alone

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT"`
- Expected outcome: returns false

### TC-21: ROOT with path

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT/domain/config"`
- Expected outcome: returns true

### TC-22: ROOT with qualifier

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT/domain/config(interface)"`
- Expected outcome: returns true

### TC-23: ARTIFACT reference

- Setup: none
- Action: call `LogicalNameHasParent` with `"ARTIFACT/x(y)"`
- Expected outcome: returns false

### TC-24: Empty string

- Setup: none
- Action: call `LogicalNameHasParent` with `""`
- Expected outcome: returns false

---

## LogicalNameHasQualifier

### TC-25: Without qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT/x"`
- Expected outcome: returns false

### TC-26: With qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT/x(y)"`
- Expected outcome: returns true

### TC-27: ARTIFACT with qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ARTIFACT/x(y)"`
- Expected outcome: returns true

### TC-28: ROOT alone

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT"`
- Expected outcome: returns false

### TC-29: Empty string

- Setup: none
- Action: call `LogicalNameHasQualifier` with `""`
- Expected outcome: returns false

---

## LogicalNameIsArtifact

### TC-30: ARTIFACT reference

- Setup: none
- Action: call `LogicalNameIsArtifact` with `"ARTIFACT/x(y)"`
- Expected outcome: returns true

### TC-31: ROOT reference

- Setup: none
- Action: call `LogicalNameIsArtifact` with `"ROOT/x(y)"`
- Expected outcome: returns false

### TC-32: Empty string

- Setup: none
- Action: call `LogicalNameIsArtifact` with `""`
- Expected outcome: returns false

---

## LogicalNameGetArtifactGenerator

### TC-33: Simple artifact

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x(y)"`
- Expected outcome: returns `"ROOT/x"`

### TC-34: Nested artifact

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x/y/z(id)"`
- Expected outcome: returns `"ROOT/x/y/z"`

### TC-35: Rejects ROOT reference

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ROOT/x(y)"`
- Expected outcome: raises error `"not an artifact reference"`

### TC-36: Artifact reference without qualifier

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x"`
- Expected outcome: returns `"ROOT/x"`
