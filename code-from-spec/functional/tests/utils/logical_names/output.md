<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@b-toYsrf3HRP9DvI95V5iULXJuw -->

## LogicalNameToPath

### ROOT alone

Setup: none.
Action: call LogicalNameToPath with `"ROOT"`.
Expected: returns PathCfs `code-from-spec/_node.md`.

### ROOT with path

Setup: none.
Action: call LogicalNameToPath with `"ROOT/payments/processor"`.
Expected: returns PathCfs `code-from-spec/payments/processor/_node.md`.

### Strips qualifier before resolving

Setup: none.
Action: call LogicalNameToPath with `"ROOT/x/y(interface)"`.
Expected: returns PathCfs `code-from-spec/x/y/_node.md`.

### Rejects ARTIFACT reference

Setup: none.
Action: call LogicalNameToPath with `"ARTIFACT/x"`.
Expected: raises error UnsupportedReference.

### Rejects unrecognized prefix

Setup: none.
Action: call LogicalNameToPath with `"UNKNOWN/something"`.
Expected: raises error UnsupportedReference.

### Rejects empty string

Setup: none.
Action: call LogicalNameToPath with `""`.
Expected: raises error UnsupportedReference.

---

## LogicalNameFromPath

### Root node

Setup: none.
Action: call LogicalNameFromPath with PathCfs `code-from-spec/_node.md`.
Expected: returns `"ROOT"`.

### Nested node

Setup: none.
Action: call LogicalNameFromPath with PathCfs `code-from-spec/x/y/_node.md`.
Expected: returns `"ROOT/x/y"`.

### Rejects non-node path

Setup: none.
Action: call LogicalNameFromPath with PathCfs `internal/config/config.go`.
Expected: raises error InvalidPath.

### Rejects path without _node.md

Setup: none.
Action: call LogicalNameFromPath with PathCfs `code-from-spec/x/y/output.md`.
Expected: raises error InvalidPath.

---

## LogicalNameGetParent

### ROOT/x parent is ROOT

Setup: none.
Action: call LogicalNameGetParent with `"ROOT/domain"`.
Expected: returns `"ROOT"`.

### ROOT/x/y parent is ROOT/x

Setup: none.
Action: call LogicalNameGetParent with `"ROOT/domain/config"`.
Expected: returns `"ROOT/domain"`.

### Strips qualifier before computing parent

Setup: none.
Action: call LogicalNameGetParent with `"ROOT/domain/config(interface)"`.
Expected: returns `"ROOT/domain"`.

### ROOT has no parent

Setup: none.
Action: call LogicalNameGetParent with `"ROOT"`.
Expected: raises error NoParent.

### Rejects ARTIFACT reference

Setup: none.
Action: call LogicalNameGetParent with `"ARTIFACT/x"`.
Expected: raises error NotARootReference.

---

## LogicalNameGetQualifier

### Extracts qualifier from ROOT reference

Setup: none.
Action: call LogicalNameGetQualifier with `"ROOT/x/y(interface)"`.
Expected: returns `"interface"`.

### ARTIFACT without qualifier returns absent

Setup: none.
Action: call LogicalNameGetQualifier with `"ARTIFACT/x/y"`.
Expected: returns absent.

### Returns absent when no qualifier

Setup: none.
Action: call LogicalNameGetQualifier with `"ROOT/x/y"`.
Expected: returns absent.

### Returns absent for ROOT alone

Setup: none.
Action: call LogicalNameGetQualifier with `"ROOT"`.
Expected: returns absent.

---

## LogicalNameStripQualifier

### Strips qualifier from ROOT reference

Setup: none.
Action: call LogicalNameStripQualifier with `"ROOT/x/y(interface)"`.
Expected: returns `"ROOT/x/y"`.

### ARTIFACT without qualifier returns unchanged

Setup: none.
Action: call LogicalNameStripQualifier with `"ARTIFACT/x/y"`.
Expected: returns `"ARTIFACT/x/y"`.

### No qualifier returns unchanged

Setup: none.
Action: call LogicalNameStripQualifier with `"ROOT/x/y"`.
Expected: returns `"ROOT/x/y"`.

### ROOT alone returns unchanged

Setup: none.
Action: call LogicalNameStripQualifier with `"ROOT"`.
Expected: returns `"ROOT"`.

### Empty string returns unchanged

Setup: none.
Action: call LogicalNameStripQualifier with `""`.
Expected: returns `""`.

---

## LogicalNameHasParent

### ROOT alone

Setup: none.
Action: call LogicalNameHasParent with `"ROOT"`.
Expected: returns false.

### ROOT with path

Setup: none.
Action: call LogicalNameHasParent with `"ROOT/domain/config"`.
Expected: returns true.

### ROOT with qualifier

Setup: none.
Action: call LogicalNameHasParent with `"ROOT/domain/config(interface)"`.
Expected: returns true.

### ARTIFACT reference

Setup: none.
Action: call LogicalNameHasParent with `"ARTIFACT/x"`.
Expected: returns false.

### Empty string

Setup: none.
Action: call LogicalNameHasParent with `""`.
Expected: returns false.

---

## LogicalNameHasQualifier

### Without qualifier

Setup: none.
Action: call LogicalNameHasQualifier with `"ROOT/x"`.
Expected: returns false.

### With qualifier

Setup: none.
Action: call LogicalNameHasQualifier with `"ROOT/x(y)"`.
Expected: returns true.

### ARTIFACT without qualifier

Setup: none.
Action: call LogicalNameHasQualifier with `"ARTIFACT/x"`.
Expected: returns false.

### ROOT alone

Setup: none.
Action: call LogicalNameHasQualifier with `"ROOT"`.
Expected: returns false.

### Empty string

Setup: none.
Action: call LogicalNameHasQualifier with `""`.
Expected: returns false.

---

## LogicalNameIsArtifact

### ARTIFACT reference

Setup: none.
Action: call LogicalNameIsArtifact with `"ARTIFACT/x"`.
Expected: returns true.

### ROOT reference

Setup: none.
Action: call LogicalNameIsArtifact with `"ROOT/x(y)"`.
Expected: returns false.

### Empty string

Setup: none.
Action: call LogicalNameIsArtifact with `""`.
Expected: returns false.

---

## LogicalNameGetArtifactGenerator

### Simple artifact

Setup: none.
Action: call LogicalNameGetArtifactGenerator with `"ARTIFACT/x"`.
Expected: returns `"ROOT/x"`.

### Nested artifact

Setup: none.
Action: call LogicalNameGetArtifactGenerator with `"ARTIFACT/x/y/z"`.
Expected: returns `"ROOT/x/y/z"`.

### Rejects ROOT reference

Setup: none.
Action: call LogicalNameGetArtifactGenerator with `"ROOT/x(y)"`.
Expected: raises error NotAnArtifactReference.
