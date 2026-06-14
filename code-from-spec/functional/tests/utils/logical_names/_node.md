---
depends_on:
  - ROOT/functional/logic/utils/logical_names(interface)
output: code-from-spec/functional/tests/utils/logical_names/output.md
---

# ROOT/functional/tests/utils/logical_names

Test cases for the logical names component.

# Public

## Test cases

### LogicalNameToPath

#### SPEC alone

Input: `"SPEC"`.
Expect: `code-from-spec/_node.md`.

#### SPEC with path

Input: `"SPEC/payments/processor"`.
Expect: `code-from-spec/payments/processor/_node.md`.

#### Strips qualifier before resolving

Input: `"SPEC/x/y(interface)"`.
Expect: `code-from-spec/x/y/_node.md`.

#### ROOT alone (backward compatibility)

Input: `"ROOT"`.
Expect: `code-from-spec/_node.md`.

#### ROOT with path (backward compatibility)

Input: `"ROOT/payments/processor"`.
Expect: `code-from-spec/payments/processor/_node.md`.

#### Rejects ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect error UnsupportedReference.

#### Rejects EXTERNAL reference

Input: `"EXTERNAL/proto/api.proto"`.
Expect error UnsupportedReference.

#### Rejects unrecognized prefix

Input: `"UNKNOWN/something"`.
Expect error UnsupportedReference.

#### Rejects empty string

Input: `""`.
Expect error UnsupportedReference.

### LogicalNameFromPath

#### Root node

Input: `code-from-spec/_node.md`.
Expect: `"SPEC"`.

#### Nested node

Input: `code-from-spec/x/y/_node.md`.
Expect: `"SPEC/x/y"`.

#### Rejects non-node path

Input: `internal/config/config.go`.
Expect error InvalidPath.

#### Rejects path without _node.md

Input: `code-from-spec/x/y/output.md`.
Expect error InvalidPath.

### LogicalNameGetParent

#### SPEC/x parent is SPEC

Input: `"SPEC/domain"`.
Expect: `"SPEC"`.

#### SPEC/x/y parent is SPEC/x

Input: `"SPEC/domain/config"`.
Expect: `"SPEC/domain"`.

#### Strips qualifier before computing parent

Input: `"SPEC/domain/config(interface)"`.
Expect: `"SPEC/domain"`.

#### SPEC has no parent

Input: `"SPEC"`.
Expect error NoParent.

#### ROOT/x parent returns SPEC parent (backward compat)

Input: `"ROOT/domain"`.
Expect: `"SPEC"`.

#### ROOT has no parent (backward compat)

Input: `"ROOT"`.
Expect error NoParent.

#### Rejects ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect error NotASpecReference.

#### Rejects EXTERNAL reference

Input: `"EXTERNAL/x"`.
Expect error NotASpecReference.

### LogicalNameGetQualifier

#### Extracts qualifier from SPEC reference

Input: `"SPEC/x/y(interface)"`.
Expect: `"interface"`.

#### Extracts qualifier from ROOT reference (backward compat)

Input: `"ROOT/x/y(interface)"`.
Expect: `"interface"`.

#### ARTIFACT without qualifier returns absent

Input: `"ARTIFACT/x/y"`.
Expect: absent.

#### EXTERNAL without qualifier returns absent

Input: `"EXTERNAL/proto/api.proto"`.
Expect: absent.

#### Returns absent when no qualifier

Input: `"SPEC/x/y"`.
Expect: absent.

#### Returns absent for SPEC alone

Input: `"SPEC"`.
Expect: absent.

### LogicalNameStripQualifier

#### Strips qualifier from SPEC reference

Input: `"SPEC/x/y(interface)"`.
Expect: `"SPEC/x/y"`.

#### Strips qualifier from ROOT reference (preserves prefix)

Input: `"ROOT/x/y(interface)"`.
Expect: `"ROOT/x/y"`.

#### ARTIFACT without qualifier — returns unchanged

Input: `"ARTIFACT/x/y"`.
Expect: `"ARTIFACT/x/y"`.

#### EXTERNAL — returns unchanged

Input: `"EXTERNAL/proto/api.proto"`.
Expect: `"EXTERNAL/proto/api.proto"`.

#### No qualifier — returns unchanged

Input: `"SPEC/x/y"`.
Expect: `"SPEC/x/y"`.

#### SPEC alone — returns unchanged

Input: `"SPEC"`.
Expect: `"SPEC"`.

#### Empty string — returns unchanged

Input: `""`.
Expect: `""`.

### LogicalNameHasParent

#### SPEC alone

Input: `"SPEC"`.
Expect: false.

#### SPEC with path

Input: `"SPEC/domain/config"`.
Expect: true.

#### ROOT alone (backward compat)

Input: `"ROOT"`.
Expect: false.

#### ROOT with path (backward compat)

Input: `"ROOT/domain/config"`.
Expect: true.

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: false.

#### EXTERNAL reference

Input: `"EXTERNAL/x"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameHasQualifier

#### Without qualifier

Input: `"SPEC/x"`.
Expect: false.

#### With qualifier

Input: `"SPEC/x(y)"`.
Expect: true.

#### ARTIFACT without qualifier

Input: `"ARTIFACT/x"`.
Expect: false.

#### EXTERNAL without qualifier

Input: `"EXTERNAL/x"`.
Expect: false.

#### SPEC alone

Input: `"SPEC"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameIsArtifact

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: true.

#### SPEC reference

Input: `"SPEC/x(y)"`.
Expect: false.

#### EXTERNAL reference

Input: `"EXTERNAL/x"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameIsSpec

#### SPEC alone

Input: `"SPEC"`.
Expect: true.

#### SPEC with path

Input: `"SPEC/x/y"`.
Expect: true.

#### ROOT alone (backward compat)

Input: `"ROOT"`.
Expect: true.

#### ROOT with path (backward compat)

Input: `"ROOT/x/y"`.
Expect: true.

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: false.

#### EXTERNAL reference

Input: `"EXTERNAL/x"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameIsExternal

#### EXTERNAL reference

Input: `"EXTERNAL/proto/api.proto"`.
Expect: true.

#### SPEC reference

Input: `"SPEC/x"`.
Expect: false.

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameGetArtifactGenerator

#### Simple artifact

Input: `"ARTIFACT/x"`.
Expect: `"SPEC/x"`.

#### Nested artifact

Input: `"ARTIFACT/x/y/z"`.
Expect: `"SPEC/x/y/z"`.

#### Rejects SPEC reference

Input: `"SPEC/x(y)"`.
Expect error NotAnArtifactReference.

#### Rejects EXTERNAL reference

Input: `"EXTERNAL/x"`.
Expect error NotAnArtifactReference.

### LogicalNameExternalToPath

#### Simple path

Input: `"EXTERNAL/proto/v1/api.proto"`.
Expect: `proto/v1/api.proto`.

#### Root-level file

Input: `"EXTERNAL/docker-compose.yaml"`.
Expect: `docker-compose.yaml`.

#### Rejects SPEC reference

Input: `"SPEC/x"`.
Expect error NotAnExternalReference.

#### Rejects ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect error NotAnExternalReference.

### LogicalNameNormalize

#### ROOT to SPEC

Input: `"ROOT/x/y"`.
Expect: `"SPEC/x/y"`.

#### ROOT bare to SPEC

Input: `"ROOT"`.
Expect: `"SPEC"`.

#### SPEC unchanged

Input: `"SPEC/x/y"`.
Expect: `"SPEC/x/y"`.

#### ARTIFACT unchanged

Input: `"ARTIFACT/x"`.
Expect: `"ARTIFACT/x"`.

#### EXTERNAL unchanged

Input: `"EXTERNAL/x"`.
Expect: `"EXTERNAL/x"`.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface:
  `LogicalNameToPath`, `LogicalNameFromPath`,
  `LogicalNameGetParent`, `LogicalNameGetQualifier`,
  `LogicalNameStripQualifier`, `LogicalNameHasParent`,
  `LogicalNameHasQualifier`, `LogicalNameIsArtifact`,
  `LogicalNameIsSpec`, `LogicalNameIsExternal`,
  `LogicalNameGetArtifactGenerator`,
  `LogicalNameExternalToPath`, `LogicalNameNormalize`.
