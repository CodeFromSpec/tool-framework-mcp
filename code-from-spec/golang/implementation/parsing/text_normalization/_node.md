---
depends_on:
  - SPEC/golang/dependencies/golang-x-text
output: internal/parsing/text_normalization.go
---

# SPEC/golang/implementation/parsing/text_normalization

Normalizes text for comparison.

# Agent

Implement the NormalizeText function listed in the
Ownership section as a Go file in package `parsing`.

## Ownership

This file declares and implements:
- Functions: `NormalizeText`

The following exist in other files of this package and
can be used but must not be redeclared:
- Error sentinels — declared in `errors.go`.
- Types (`NodeFrontmatter`, `Node`, `CfsReference`,
  etc.) — declared in other files.

All unexported helpers must use the suffix `Norm`
(e.g. `collapseWhitespaceNorm`). This is mandatory to
avoid name collisions with other files in the package.

## Logic

1. If raw_string is empty, return "".

2. Trim leading and trailing whitespace characters from
   raw_string, where whitespace is defined as space
   (U+0020) and horizontal tab (U+0009) only.

3. Collapse each consecutive run of whitespace
   characters (space U+0020 and horizontal tab U+0009)
   to a single space (U+0020).

4. Apply Unicode simple case folding to the resulting
   string. This converts uppercase characters to their
   lowercase equivalents, including Unicode mappings
   (e.g., "Straße" -> "strasse").

5. Return the normalized string.

## Go-specific guidance

- Use `golang.org/x/text` for Unicode case folding and
  normalization as described in the logic.
- The package name should be `parsing`.
- Whitespace is defined strictly as U+0020 (space) and
  U+0009 (horizontal tab). Do not use standard library
  "isspace" functions that may match U+00A0 or other
  characters.
