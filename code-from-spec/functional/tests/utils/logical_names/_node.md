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

#### ROOT alone

Input: `"ROOT"`.
Expect: `code-from-spec/_node.md`.

#### ROOT with path

Input: `"ROOT/payments/processor"`.
Expect: `code-from-spec/payments/processor/_node.md`.

#### Strips qualifier before resolving

Input: `"ROOT/x/y(interface)"`.
Expect: `code-from-spec/x/y/_node.md`.

#### Rejects ARTIFACT reference

Input: `"ARTIFACT/x"`.
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
Expect: `"ROOT"`.

#### Nested node

Input: `code-from-spec/x/y/_node.md`.
Expect: `"ROOT/x/y"`.

#### Rejects non-node path

Input: `internal/config/config.go`.
Expect error InvalidPath.

#### Rejects path without _node.md

Input: `code-from-spec/x/y/output.md`.
Expect error InvalidPath.

### LogicalNameGetParent

#### ROOT/x parent is ROOT

Input: `"ROOT/domain"`.
Expect: `"ROOT"`.

#### ROOT/x/y parent is ROOT/x

Input: `"ROOT/domain/config"`.
Expect: `"ROOT/domain"`.

#### Strips qualifier before computing parent

Input: `"ROOT/domain/config(interface)"`.
Expect: `"ROOT/domain"`.

#### ROOT has no parent

Input: `"ROOT"`.
Expect error NoParent.

#### Rejects ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect error NotARootReference.

### LogicalNameGetQualifier

#### Extracts qualifier from ROOT reference

Input: `"ROOT/x/y(interface)"`.
Expect: `"interface"`.

#### ARTIFACT without qualifier returns absent

Input: `"ARTIFACT/x/y"`.
Expect: absent.

#### Returns absent when no qualifier

Input: `"ROOT/x/y"`.
Expect: absent.

#### Returns absent for ROOT alone

Input: `"ROOT"`.
Expect: absent.

### LogicalNameStripQualifier

#### Strips qualifier from ROOT reference

Input: `"ROOT/x/y(interface)"`.
Expect: `"ROOT/x/y"`.

#### ARTIFACT without qualifier — returns unchanged

Input: `"ARTIFACT/x/y"`.
Expect: `"ARTIFACT/x/y"`.

#### No qualifier — returns unchanged

Input: `"ROOT/x/y"`.
Expect: `"ROOT/x/y"`.

#### ROOT alone — returns unchanged

Input: `"ROOT"`.
Expect: `"ROOT"`.

#### Empty string — returns unchanged

Input: `""`.
Expect: `""`.

### LogicalNameHasParent

#### ROOT alone

Input: `"ROOT"`.
Expect: false.

#### ROOT with path

Input: `"ROOT/domain/config"`.
Expect: true.

#### ROOT with qualifier

Input: `"ROOT/domain/config(interface)"`.
Expect: true.

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameHasQualifier

#### Without qualifier

Input: `"ROOT/x"`.
Expect: false.

#### With qualifier

Input: `"ROOT/x(y)"`.
Expect: true.

#### ARTIFACT without qualifier

Input: `"ARTIFACT/x"`.
Expect: false.

#### ROOT alone

Input: `"ROOT"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameIsArtifact

#### ARTIFACT reference

Input: `"ARTIFACT/x"`.
Expect: true.

#### ROOT reference

Input: `"ROOT/x(y)"`.
Expect: false.

#### Empty string

Input: `""`.
Expect: false.

### LogicalNameGetArtifactGenerator

#### Simple artifact

Input: `"ARTIFACT/x"`.
Expect: `"ROOT/x"`.

ARTIFACT/ references no longer have qualifiers.

#### Nested artifact

Input: `"ARTIFACT/x/y/z"`.
Expect: `"ROOT/x/y/z"`.

#### Rejects ROOT reference

Input: `"ROOT/x(y)"`.
Expect error NotAnArtifactReference.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function names from the interface:
  `LogicalNameToPath`, `LogicalNameFromPath`,
  `LogicalNameGetParent`, `LogicalNameGetQualifier`,
  `LogicalNameStripQualifier`, `LogicalNameHasParent`,
  `LogicalNameHasQualifier`, `LogicalNameIsArtifact`,
  `LogicalNameGetArtifactGenerator`.
