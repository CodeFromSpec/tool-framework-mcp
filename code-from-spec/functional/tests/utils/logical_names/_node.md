---
depends_on:
  - ROOT/functional/logic/utils/logical_names(interface)
outputs:
  - id: logical_names_tests
    path: code-from-spec/functional/tests/utils/logical_names/output.md
---

# ROOT/functional/tests/utils/logical_names

Test cases for the logical names component.

# Public

## Test cases

### PathFromLogicalName

#### ROOT

Input: "ROOT". Expect path = "code-from-spec/_node.md",
ok = true.

#### ROOT with path

Input: "ROOT/payments/processor". Expect path =
"code-from-spec/payments/processor/_node.md", ok = true.

#### ROOT with qualifier

Input: "ROOT/payments/processor(interface)". Expect path =
"code-from-spec/payments/processor/_node.md", ok = true.

#### ROOT with qualifier -- strips qualifier from path

Input: "ROOT/x(y)". Expect path =
"code-from-spec/x/_node.md", ok = true.

#### ARTIFACT reference returns false

Input: "ARTIFACT/x(y)". Expect path = "", ok = false.

#### Unrecognized prefix

Input: "UNKNOWN/something". Expect path = "", ok = false.

#### Empty string

Input: "". Expect path = "", ok = false.

### HasParent

#### ROOT

Input: "ROOT". Expect has parent = false, ok = true.

#### ROOT with path

Input: "ROOT/domain/config". Expect has parent = true,
ok = true.

#### ROOT with qualifier

Input: "ROOT/domain/config(interface)". Expect has parent =
true, ok = true.

#### ARTIFACT returns false false

Input: "ARTIFACT/x(y)". Expect has parent = false,
ok = false.

#### Empty string

Input: "". Expect has parent = false, ok = false.

### ParentLogicalName

#### ROOT/x -- parent is ROOT

Input: "ROOT/domain". Expect parent = "ROOT", ok = true.

#### ROOT/x/y -- parent is ROOT/x

Input: "ROOT/domain/config". Expect parent = "ROOT/domain",
ok = true.

#### ROOT/x/y(z) -- parent is ROOT/x

Input: "ROOT/domain/config(interface)". Expect parent =
"ROOT/domain", ok = true.

#### ROOT has no parent

Input: "ROOT". Expect parent = "", ok = false.

#### Empty string invalid

Input: "". Expect parent = "", ok = false.

### HasQualifier

#### ROOT without qualifier

Input: "ROOT/x". Expect has qualifier = false, ok = true.

#### ROOT with qualifier

Input: "ROOT/x(y)". Expect has qualifier = true, ok = true.

#### ARTIFACT with qualifier

Input: "ARTIFACT/x(y)". Expect has qualifier = true,
ok = true.

#### ROOT alone

Input: "ROOT". Expect has qualifier = false, ok = true.

#### Empty string

Input: "". Expect has qualifier = false, ok = false.

### QualifierName

#### ROOT with qualifier

Input: "ROOT/x(y)". Expect qualifier = "y", ok = true.

#### ROOT with nested path and qualifier

Input: "ROOT/x/y(interface)". Expect qualifier =
"interface", ok = true.

#### ARTIFACT with qualifier

Input: "ARTIFACT/x(y)". Expect qualifier = "y", ok = true.

#### ROOT without qualifier

Input: "ROOT/x". Expect qualifier = "", ok = false.

#### ROOT alone

Input: "ROOT". Expect qualifier = "", ok = false.

#### Empty string

Input: "". Expect qualifier = "", ok = false.

### IsArtifactRef

#### ARTIFACT reference

Input: "ARTIFACT/x(y)". Expect true.

#### ROOT reference

Input: "ROOT/x(y)". Expect false.

#### Empty string

Input: "". Expect false.

### ArtifactRefParts

#### ARTIFACT/x(y)

Input: "ARTIFACT/x(y)". Expect path =
"code-from-spec/x/_node.md", id = "y", ok = true.

#### ARTIFACT/x/y(z)

Input: "ARTIFACT/x/y(z)". Expect path =
"code-from-spec/x/y/_node.md", id = "z", ok = true.

#### ARTIFACT without qualifier returns false

Input: "ARTIFACT/x". Expect path = "", id = "", ok = false.

#### ROOT reference returns false

Input: "ROOT/x(y)". Expect path = "", id = "", ok = false.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
