---
output: internal/parsing/errors.go
---

# SPEC/golang/implementation/parsing/errors

All error sentinels for the parsing package, declared
in a single file to avoid collisions across the
independently generated files in the package.

# Agent

Generate a Go file in package `parsing` declaring all
error sentinels listed below. Use `errors.New` for
each. No logic — declarations only.

## Ownership

This file declares all error sentinels for the package.
No other file in the package may declare error
sentinels. This file has no unexported helpers.

## Declarations

```go
var (
	// Frontmatter errors
	ErrFileUnreadable = errors.New("file unreadable")
	ErrMalformedYAML  = errors.New("malformed YAML")

	// Node parsing errors
	ErrNotASpecReference                  = errors.New("not a SPEC/ reference")
	ErrHasQualifier                       = errors.New("logical name has qualifier")
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")
	ErrNodeNameDoesNotMatch               = errors.New("node name does not match")
	ErrDuplicatePublicSection             = errors.New("duplicate # Public section")
	ErrDuplicateAgentSection              = errors.New("duplicate # Agent section")
	ErrDuplicatePrivateSection            = errors.New("duplicate # Private section")
	ErrUnrecognizedSection                = errors.New("unrecognized section")
	ErrDuplicateSubsection               = errors.New("duplicate subsection")

	// Logical name errors
	ErrUnrecognizedPrefix = errors.New("unrecognized prefix")
	ErrInvalidName        = errors.New("invalid name")
	ErrNoOutput           = errors.New("no output declared")
	ErrInvalidPath        = errors.New("invalid path")
)
```
