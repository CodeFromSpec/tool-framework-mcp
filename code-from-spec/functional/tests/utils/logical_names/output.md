<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@JTj8knDgfRpOusK77S61liPkfTo -->

# Test Specification: logical_names utilities

---

## LogicalNameToPath

### Test: ROOT alone

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT"`
- Expected outcome: returns `code-from-spec/_node.md`

---

### Test: ROOT with path

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT/payments/processor"`
- Expected outcome: returns `code-from-spec/payments/processor/_node.md`

---

### Test: Strips qualifier before resolving

- Setup: none
- Action: call `LogicalNameToPath` with `"ROOT/x/y(interface)"`
- Expected outcome: returns `code-from-spec/x/y/_node.md`

---

### Test: Rejects ARTIFACT reference

- Setup: none
- Action: call `LogicalNameToPath` with `"ARTIFACT/x(y)"`
- Expected outcome: raises error `"unsupported reference"`

---

### Test: Rejects unrecognized prefix

- Setup: none
- Action: call `LogicalNameToPath` with `"UNKNOWN/something"`
- Expected outcome: raises error `"unsupported reference"`

---

### Test: Rejects empty string

- Setup: none
- Action: call `LogicalNameToPath` with `""`
- Expected outcome: raises error `"unsupported reference"`

---

## LogicalNameFromPath

### Test: Root node

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/_node.md`
- Expected outcome: returns `"ROOT"`

---

### Test: Nested node

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/x/y/_node.md`
- Expected outcome: returns `"ROOT/x/y"`

---

### Test: Rejects non-node path

- Setup: none
- Action: call `LogicalNameFromPath` with `internal/config/config.go`
- Expected outcome: raises error `"invalid path"`

---

### Test: Rejects path without _node.md

- Setup: none
- Action: call `LogicalNameFromPath` with `code-from-spec/x/y/output.md`
- Expected outcome: raises error `"invalid path"`

---

## LogicalNameGetParent

### Test: ROOT/x parent is ROOT

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain"`
- Expected outcome: returns `"ROOT"`

---

### Test: ROOT/x/y parent is ROOT/x

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain/config"`
- Expected outcome: returns `"ROOT/domain"`

---

### Test: Strips qualifier before computing parent

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT/domain/config(interface)"`
- Expected outcome: returns `"ROOT/domain"`

---

### Test: ROOT has no parent

- Setup: none
- Action: call `LogicalNameGetParent` with `"ROOT"`
- Expected outcome: raises error `"no parent"`

---

### Test: Rejects ARTIFACT reference

- Setup: none
- Action: call `LogicalNameGetParent` with `"ARTIFACT/x(y)"`
- Expected outcome: raises error `"not a ROOT reference"`

---

## LogicalNameGetQualifier

### Test: Extracts qualifier from ROOT reference

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT/x/y(interface)"`
- Expected outcome: returns `"interface"`

---

### Test: Extracts qualifier from ARTIFACT reference

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ARTIFACT/x/y(id)"`
- Expected outcome: returns `"id"`

---

### Test: Returns absent when no qualifier

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT/x/y"`
- Expected outcome: returns absent

---

### Test: Returns absent for ROOT alone

- Setup: none
- Action: call `LogicalNameGetQualifier` with `"ROOT"`
- Expected outcome: returns absent

---

## LogicalNameStripQualifier

### Test: Strips qualifier from ROOT reference

- Setup: none
- Action: call `LogicalNameStripQualifier` with `"ROOT/x/y(interface)"`
- Expected outcome: returns `"ROOT/x/y"`

---

### Test: Strips qualifier from ARTIFACT reference

- Setup: none
- Action: call `LogicalNameStripQualifier` with `"ARTIFACT/x/y(id)"`
- Expected outcome: returns `"ARTIFACT/x/y"`

---

### Test: No qualifier — returns unchanged

- Setup: none
- Action: call `LogicalNameStripQualifier` with `"ROOT/x/y"`
- Expected outcome: returns `"ROOT/x/y"`

---

### Test: ROOT alone — returns unchanged

- Setup: none
- Action: call `LogicalNameStripQualifier` with `"ROOT"`
- Expected outcome: returns `"ROOT"`

---

### Test: Empty string — returns unchanged

- Setup: none
- Action: call `LogicalNameStripQualifier` with `""`
- Expected outcome: returns `""`

---

## LogicalNameHasParent

### Test: ROOT alone

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT"`
- Expected outcome: returns false

---

### Test: ROOT with path

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT/domain/config"`
- Expected outcome: returns true

---

### Test: ROOT with qualifier

- Setup: none
- Action: call `LogicalNameHasParent` with `"ROOT/domain/config(interface)"`
- Expected outcome: returns true

---

### Test: ARTIFACT reference

- Setup: none
- Action: call `LogicalNameHasParent` with `"ARTIFACT/x(y)"`
- Expected outcome: returns false

---

### Test: Empty string

- Setup: none
- Action: call `LogicalNameHasParent` with `""`
- Expected outcome: returns false

---

## LogicalNameHasQualifier

### Test: Without qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT/x"`
- Expected outcome: returns false

---

### Test: With qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT/x(y)"`
- Expected outcome: returns true

---

### Test: ARTIFACT with qualifier

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ARTIFACT/x(y)"`
- Expected outcome: returns true

---

### Test: ROOT alone

- Setup: none
- Action: call `LogicalNameHasQualifier` with `"ROOT"`
- Expected outcome: returns false

---

### Test: Empty string

- Setup: none
- Action: call `LogicalNameHasQualifier` with `""`
- Expected outcome: returns false

---

## LogicalNameIsArtifact

### Test: ARTIFACT reference

- Setup: none
- Action: call `LogicalNameIsArtifact` with `"ARTIFACT/x(y)"`
- Expected outcome: returns true

---

### Test: ROOT reference

- Setup: none
- Action: call `LogicalNameIsArtifact` with `"ROOT/x(y)"`
- Expected outcome: returns false

---

### Test: Empty string

- Setup: none
- Action: call `LogicalNameIsArtifact` with `""`
- Expected outcome: returns false

---

## LogicalNameGetArtifactGenerator

### Test: Simple artifact

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x(y)"`
- Expected outcome: returns `"ROOT/x"`

---

### Test: Nested artifact

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x/y/z(id)"`
- Expected outcome: returns `"ROOT/x/y/z"`

---

### Test: Rejects ROOT reference

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ROOT/x(y)"`
- Expected outcome: raises error `"not an artifact reference"`

---

### Test: Artifact reference without qualifier

- Setup: none
- Action: call `LogicalNameGetArtifactGenerator` with `"ARTIFACT/x"`
- Expected outcome: returns `"ROOT/x"`
