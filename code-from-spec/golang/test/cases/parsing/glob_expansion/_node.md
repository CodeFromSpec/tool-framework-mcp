---
depends_on:
  - SPEC/golang/test/utils/helpers
  - SPEC/golang/implementation/parsing(interface)
output: internal/parsingglobexpansiontest/parsing_glob_expansion_test.go
---

# SPEC/golang/test/cases/parsing/glob_expansion

Unit tests for `parsing.ExpandGlob`.

# Agent

## Context

`ExpandGlob` is a pure function over string lists —
no filesystem needed. Tests pass a pattern, a list of
known SPEC/ logical names, and optionally a declaring
node, then verify the returned list.

## Test cases

### Valid patterns

#### SPEC glob — matches descendants

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b", "SPEC/a/b/c",
  "SPEC/a/d"].

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes, nil).

Expected:
- Result = ["SPEC/a/b", "SPEC/a/b/c", "SPEC/a/d"]
  (sorted).

#### ARTIFACT glob — converts prefix

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b", "SPEC/a/c"].

Actions:
1. Call parsing.ExpandGlob("ARTIFACT/a/*", knownNodes,
   nil).

Expected:
- Result = ["ARTIFACT/a/b", "ARTIFACT/a/c"].

#### Glob excludes declaring node

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b", "SPEC/a/c"].
- declaringNode = "SPEC/a/b".

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes,
   &declaringNode).

Expected:
- Result = ["SPEC/a/c"]. SPEC/a/b excluded.

#### Glob excludes declaring node's ancestors

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b", "SPEC/a/b/c",
  "SPEC/a/d"].
- declaringNode = "SPEC/a/b/c".

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes,
   &declaringNode).

Expected:
- Result = ["SPEC/a/d"]. SPEC/a/b (ancestor) and
  SPEC/a/b/c (self) excluded.

#### ARTIFACT glob excludes declaring node's counterpart

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b", "SPEC/a/c"].
- declaringNode = "SPEC/a/b".

Actions:
1. Call parsing.ExpandGlob("ARTIFACT/a/*", knownNodes,
   &declaringNode).

Expected:
- Result = ["ARTIFACT/a/c"]. ARTIFACT/a/b excluded
  (counterpart of declaring node).

#### Empty match — no error

Setup:
- knownNodes = ["SPEC/a", "SPEC/b"].

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes, nil).

Expected:
- Result = empty list ([]string{}). No error.

#### Result is sorted

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/z", "SPEC/a/m",
  "SPEC/a/b"].

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes, nil).

Expected:
- Result = ["SPEC/a/b", "SPEC/a/m", "SPEC/a/z"].

#### Declaring node nil — no exclusion

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b"].

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes, nil).

Expected:
- Result = ["SPEC/a/b"].

### Invalid parameters

#### knownNodes nil — error

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", nil, nil).

Expected: Error parsing.ErrEmptyNodeList.

#### knownNodes empty — error

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", []string{}, nil).

Expected: Error parsing.ErrEmptyNodeList.

#### declaringNode not SPEC/ — error

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b"].
- declaringNode = "ARTIFACT/a".

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes,
   &declaringNode).

Expected: Error parsing.ErrInvalidName.

#### declaringNode empty string — error

Setup:
- knownNodes = ["SPEC/a", "SPEC/a/b"].
- declaringNode = "".

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*", knownNodes,
   &declaringNode).

Expected: Error parsing.ErrInvalidName.

### Invalid patterns

#### EXTERNAL glob — error

Actions:
1. Call parsing.ExpandGlob("EXTERNAL/docs/*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### VERDICT glob — error

Actions:
1. Call parsing.ExpandGlob("VERDICT/a/*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### Partial wildcard — error

Actions:
1. Call parsing.ExpandGlob("SPEC/a/foo*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### Multiple wildcards — error

Actions:
1. Call parsing.ExpandGlob("SPEC/*/a/*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### Qualifier on glob — error

Actions:
1. Call parsing.ExpandGlob("SPEC/a/*(interface)", nil,
   nil).

Expected: Error parsing.ErrInvalidGlob.

#### Missing relative path — error

Actions:
1. Call parsing.ExpandGlob("SPEC/*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### Unrecognized prefix — error

Actions:
1. Call parsing.ExpandGlob("UNKNOWN/a/*", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

#### Pattern without glob suffix — error

Actions:
1. Call parsing.ExpandGlob("SPEC/a/b", nil, nil).

Expected: Error parsing.ErrInvalidGlob.

## Go-specific guidance

- The package name is `parsingglobexpansiontest`
  (external test package).
- Use `errors.Is` for error assertions.
- Use `testutils.AssertStringSlicesEqual` or manual
  slice comparison for result assertions.
