---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/parsing/artifact_tag/output.md
---

# SPEC/golang/interfaces/parsing/artifact_tag

Extracts the artifact tag from generated files for
staleness detection.

# Public

## Package

`package artifacttag`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag"`

## Interface

```go
type ArtifactTag struct {
	LogicalName string
	Hash        string
}

func ArtifactTagExtract(filePath pathutils.PathCfs) (*ArtifactTag, error)
```

### Artifact tag format

Generated files contain the string:

```
code-from-spec: <logical-name>@<hash>
```

The tag may appear inside any comment syntax. The tool
scans each line for the pattern regardless of context.

### Errors

- `ErrNoTagFound`: the file has no `code-from-spec:`
  substring.
- `ErrMalformedTag`: the tag exists but cannot be
  parsed (no @, empty name, wrong hash length).
- Propagated errors from `file` package.

# Agent

Generate an interface specification document listing
the package, import path, struct definition, and
function signatures.
