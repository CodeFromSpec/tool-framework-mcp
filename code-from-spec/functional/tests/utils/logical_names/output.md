<!-- code-from-spec: ROOT/functional/tests/utils/logical_names@lwqioTmUSQCDLK7eOlJt0d5hKvk -->

# Test Specification: Logical Names

## LogicalNameToPath

### ROOT alone

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `"ROOT"`.
2. Expect result `code-from-spec/_node.md`.

---

### ROOT with path

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `"ROOT/payments/processor"`.
2. Expect result `code-from-spec/payments/processor/_node.md`.

---

### Strips qualifier before resolving

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `"ROOT/x/y(interface)"`.
2. Expect result `code-from-spec/x/y/_node.md`.

---

### Rejects ARTIFACT reference

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `"ARTIFACT/x"`.
2. Expect error UnsupportedReference.

---

### Rejects unrecognized prefix

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `"UNKNOWN/something"`.
2. Expect error UnsupportedReference.

---

### Rejects empty string

Setup: None.

Actions:
1. Call `LogicalNameToPath` with input `""`.
2. Expect error UnsupportedReference.

---

## LogicalNameFromPath

### Root node

Setup: None.

Actions:
1. Call `LogicalNameFromPath` with input `code-from-spec/_node.md`.
2. Expect result `"ROOT"`.

---

### Nested node

Setup: None.

Actions:
1. Call `LogicalNameFromPath` with input `code-from-spec/x/y/_node.md`.
2. Expect result `"ROOT/x/y"`.

---

### Rejects non-node path

Setup: None.

Actions:
1. Call `LogicalNameFromPath` with input `internal/config/config.go`.
2. Expect error InvalidPath.

---

### Rejects path without _node.md

Setup: None.

Actions:
1. Call `LogicalNameFromPath` with input `code-from-spec/x/y/output.md`.
2. Expect error InvalidPath.

---

## LogicalNameGetParent

### ROOT/x parent is ROOT

Setup: None.

Actions:
1. Call `LogicalNameGetParent` with input `"ROOT/domain"`.
2. Expect result `"ROOT"`.

---

### ROOT/x/y parent is ROOT/x

Setup: None.

Actions:
1. Call `LogicalNameGetParent` with input `"ROOT/domain/config"`.
2. Expect result `"ROOT/domain"`.

---

### Strips qualifier before computing parent

Setup: None.

Actions:
1. Call `LogicalNameGetParent` with input `"ROOT/domain/config(interface)"`.
2. Expect result `"ROOT/domain"`.

---

### ROOT has no parent

Setup: None.

Actions:
1. Call `LogicalNameGetParent` with input `"ROOT"`.
2. Expect error NoParent.

---

### Rejects ARTIFACT reference

Setup: None.

Actions:
1. Call `LogicalNameGetParent` with input `"ARTIFACT/x"`.
2. Expect error NotARootReference.

---

## LogicalNameGetQualifier

### Extracts qualifier from ROOT reference

Setup: None.

Actions:
1. Call `LogicalNameGetQualifier` with input `"ROOT/x/y(interface)"`.
2. Expect result `"interface"`.

---

### ARTIFACT without qualifier returns absent

Setup: None.

Actions:
1. Call `LogicalNameGetQualifier` with input `"ARTIFACT/x/y"`.
2. Expect result absent.

---

### Returns absent when no qualifier

Setup: None.

Actions:
1. Call `LogicalNameGetQualifier` with input `"ROOT/x/y"`.
2. Expect result absent.

---

### Returns absent for ROOT alone

Setup: None.

Actions:
1. Call `LogicalNameGetQualifier` with input `"ROOT"`.
2. Expect result absent.

---

## LogicalNameStripQualifier

### Strips qualifier from ROOT reference

Setup: None.

Actions:
1. Call `LogicalNameStripQualifier` with input `"ROOT/x/y(interface)"`.
2. Expect result `"ROOT/x/y"`.

---

### ARTIFACT without qualifier — returns unchanged

Setup: None.

Actions:
1. Call `LogicalNameStripQualifier` with input `"ARTIFACT/x/y"`.
2. Expect result `"ARTIFACT/x/y"`.

---

### No qualifier — returns unchanged

Setup: None.

Actions:
1. Call `LogicalNameStripQualifier` with input `"ROOT/x/y"`.
2. Expect result `"ROOT/x/y"`.

---

### ROOT alone — returns unchanged

Setup: None.

Actions:
1. Call `LogicalNameStripQualifier` with input `"ROOT"`.
2. Expect result `"ROOT"`.

---

### Empty string — returns unchanged

Setup: None.

Actions:
1. Call `LogicalNameStripQualifier` with input `""`.
2. Expect result `""`.

---

## LogicalNameHasParent

### ROOT alone

Setup: None.

Actions:
1. Call `LogicalNameHasParent` with input `"ROOT"`.
2. Expect result false.

---

### ROOT with path

Setup: None.

Actions:
1. Call `LogicalNameHasParent` with input `"ROOT/domain/config"`.
2. Expect result true.

---

### ROOT with qualifier

Setup: None.

Actions:
1. Call `LogicalNameHasParent` with input `"ROOT/domain/config(interface)"`.
2. Expect result true.

---

### ARTIFACT reference

Setup: None.

Actions:
1. Call `LogicalNameHasParent` with input `"ARTIFACT/x"`.
2. Expect result false.

---

### Empty string

Setup: None.

Actions:
1. Call `LogicalNameHasParent` with input `""`.
2. Expect result false.

---

## LogicalNameHasQualifier

### Without qualifier

Setup: None.

Actions:
1. Call `LogicalNameHasQualifier` with input `"ROOT/x"`.
2. Expect result false.

---

### With qualifier

Setup: None.

Actions:
1. Call `LogicalNameHasQualifier` with input `"ROOT/x(y)"`.
2. Expect result true.

---

### ARTIFACT without qualifier

Setup: None.

Actions:
1. Call `LogicalNameHasQualifier` with input `"ARTIFACT/x"`.
2. Expect result false.

---

### ROOT alone

Setup: None.

Actions:
1. Call `LogicalNameHasQualifier` with input `"ROOT"`.
2. Expect result false.

---

### Empty string

Setup: None.

Actions:
1. Call `LogicalNameHasQualifier` with input `""`.
2. Expect result false.

---

## LogicalNameIsArtifact

### ARTIFACT reference

Setup: None.

Actions:
1. Call `LogicalNameIsArtifact` with input `"ARTIFACT/x"`.
2. Expect result true.

---

### ROOT reference

Setup: None.

Actions:
1. Call `LogicalNameIsArtifact` with input `"ROOT/x(y)"`.
2. Expect result false.

---

### Empty string

Setup: None.

Actions:
1. Call `LogicalNameIsArtifact` with input `""`.
2. Expect result false.

---

## LogicalNameGetArtifactGenerator

### Simple artifact

Setup: None.

Actions:
1. Call `LogicalNameGetArtifactGenerator` with input `"ARTIFACT/x"`.
2. Expect result `"ROOT/x"`.

---

### Nested artifact

Setup: None.

Actions:
1. Call `LogicalNameGetArtifactGenerator` with input `"ARTIFACT/x/y/z"`.
2. Expect result `"ROOT/x/y/z"`.

---

### Rejects ROOT reference

Setup: None.

Actions:
1. Call `LogicalNameGetArtifactGenerator` with input `"ROOT/x(y)"`.
2. Expect error NotAnArtifactReference.
