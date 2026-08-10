---
output: internal/parsing/glob_expansion.go
---

# SPEC/golang/implementation/parsing/glob_expansion

Expands glob references (`SPEC/x/*`, `ARTIFACT/x/*`,
`VERDICT/x/*`) in `imports`, `input`, and `wait_on`
fields to the sorted list of concrete logical names
matching the pattern.

# Agent

Implement the function listed in the Ownership section
as a Go file in package `parsing`.

## Ownership

This file declares and implements:
- Function: `ExpandGlob`

The following exist in other files of this package and
can be used but must not be redeclared:
- Error sentinels — declared in `errors.go`.

To avoid name collisions with other files in this
package, all identifiers you declare beyond the ones
listed in the Ownership section (functions, variables,
types) must use the suffix `GE`.

## Logic

### ExpandGlob(pattern string, knownNodes []string, declaringNode *string) -> ([]string, error)

1. **Validate parameters.**

   If `knownNodes` is nil or empty, return
   ErrEmptyNodeList.

   If `declaringNode` is non-nil and does not start
   with `SPEC/`, return ErrInvalidName.

2. **Validate glob syntax.**

   If pattern does not end with `/*`, return
   ErrInvalidGlob.

   If pattern contains `(`, return ErrInvalidGlob
   (qualifier on glob).

   Let `basePath` = pattern with trailing `/*` removed.
   E.g., `SPEC/payments/*` becomes `SPEC/payments`.

   Classify `basePath` by prefix:
   - If it starts with `SPEC/`: let `prefix` = `SPEC/`,
     `relative` = basePath with `SPEC/` removed.
   - Else if it starts with `ARTIFACT/`: let
     `prefix` = `ARTIFACT/`,
     `relative` = basePath with `ARTIFACT/` removed.
   - Else if it starts with `VERDICT/`: let
     `prefix` = `VERDICT/`,
     `relative` = basePath with `VERDICT/` removed.
   - Else: return ErrInvalidGlob.
     This rejects `EXTERNAL/` globs and unrecognized
     prefixes.

   If `relative` is empty, return ErrInvalidGlob.

   If `relative` contains `*`, return ErrInvalidGlob
   (multiple wildcards or partial glob in earlier
   segments).

3. **Compute match prefix.**

   Let `matchPrefix` = `SPEC/` + `relative` + `/`.

   All glob prefixes match against `SPEC/` logical
   names. `ARTIFACT/` globs produce `ARTIFACT/` names
   and `VERDICT/` globs produce `VERDICT/` names in
   the output. `SPEC/` globs produce `SPEC/` names.

4. **Expand.**

   Initialize an empty result list.

   For each name in `knownNodes`:
     If name starts with `matchPrefix`:
       If prefix is `SPEC/`:
         Add name to results.
       Else (prefix is `ARTIFACT/` or `VERDICT/`):
         Let `specRelative` = name with `SPEC/` removed.
         Add `prefix` + `specRelative` to results.

5. **Exclude declaring node and ancestors.**

   If `declaringNode` is non-nil:
     Let `declRelative` = *declaringNode with `SPEC/`
     removed. (The declaring node is always a SPEC/
     node.)

     Compute the names to exclude: the declaring node
     and each ancestor. For `a/b/c`, the set is
     `{a/b/c, a/b, a}`. Build this by repeatedly
     stripping the last `/`-delimited segment.

     For each excluded relative path, build the
     excluded name using `prefix` (the glob's prefix,
     not `SPEC/`):
       excludedName = prefix + excludedRelative.

     Remove all excluded names from results.

6. **Sort and return.**

   Sort results alphabetically.
   Return results. An empty list is not an error.

## Go-specific guidance

- Use `strings.HasSuffix`, `strings.HasPrefix`,
  `strings.TrimSuffix`, `strings.TrimPrefix`,
  `strings.Contains`, `strings.LastIndex` for string
  operations.
- Use `sort.Strings` for sorting.
- Error sentinels are declared in `errors.go` — do
  not redeclare them here.
- The package name should be `parsing`.
