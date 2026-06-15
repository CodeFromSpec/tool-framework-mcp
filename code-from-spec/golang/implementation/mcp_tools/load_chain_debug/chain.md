<!-- code-from-spec: SPEC/golang/implementation/mcp_tools/load_chain_debug@cZomqhQq1E9UGVpaY3yV6AsG3v0 -->
chain_hash: cZomqhQq1E9UGVpaY3yV6AsG3v0
--- context ---
## Go module
The module path is `github.com/CodeFromSpec/tool-framework-mcp/v4`.
All internal package imports must use this prefix.

## Language
Go (minimum 1.24).

## Dependencies
- Standard library unless explicitly stated otherwise.
- `github.com/modelcontextprotocol/go-sdk` — Official MCP SDK
  (stdio transport, tool registration with generics, request
  handling).

## Error handling
- **Startup errors** (unexpected arguments) — print to stderr and
  exit 1. The tool does not start if it cannot be configured.
- **Tool errors** — returned as MCP tool error responses. The tool
  continues running after a tool error.

## Project root
The tool is always executed from the project root directory.
The working directory of the process is the project root.
All relative paths — spec files, generated source files — are
resolved against it.

## Constraints
- When a functional spec lists `errors:` on a function,
  the Go implementation must include `error` as the last
  return value, following standard Go convention. This
  applies even when the function already returns multiple
  values.
- Always check pointers for nil before dereferencing,
  including struct fields that are pointers. Do not rely
  on caller guarantees — defend at the point of use.
- Every error return value must be checked.
- Always compare errors with `errors.Is` or `errors.As`,
  never with `==`. Sentinel errors may be wrapped.
- No test framework beyond the standard `testing` package.
- No configuration files.
- All test helper functions and types must be prefixed with `test`
  (e.g., `testMakeFM`, `testIntPtr`, `testCase`). This prevents
  name collisions with unexported functions and types in the
  package under test when using internal test files (same package
  as the implementation).

## Implementation rules
- Implement the pseudocode from the `input` artifact.
- Declare types, error sentinels, and function signatures
  exactly as specified in the interface artifact from
  `depends_on` — same names, same receiver types, same
  return types. The interface is the contract. The output
  file is the sole `.go` file in the package — it must
  contain all declarations from the interface.
- Use the package name declared in the interface artifact.
- Write idiomatic Go: camelCase for local variables and
  parameters, exported names for public API, receiver
  methods where the interface specifies them.
- Wrap all errors with `fmt.Errorf` using `%w` so callers
  can match with `errors.Is()`.
- Write straightforward code. Simple and readable over
  clever and compact.


## Package

```go
package chainhash
```

## Import Path

```
github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash
```

## Error Sentinels

```go
package chainhash

import "errors"

var ErrParseFailure = errors.New("parse failure")
```

## Functions

```go
package chainhash

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a content
// hash (SHA-1) for each, concatenates all content hashes as raw bytes in
// chain assembly order, and computes the final SHA-1 of the concatenation.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chain hash:", hash)
}
```


# Package `chainresolver`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver`

## Types

```go
package chainresolver

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

// ChainItem represents a single node in the resolved chain.
type ChainItem struct {
	UnqualifiedLogicalName string
	FilePath               pathutils.PathCfs
	Qualifier              *string
}

// Chain is the fully resolved chain for a target logical name.
type Chain struct {
	Ancestors    []*ChainItem
	Dependencies []*ChainItem
	Target       *ChainItem
	Input        *ChainItem
}
```

## Error Sentinels

```go
package chainresolver

import "errors"

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrUnresolvableArtifact  = errors.New("unresolvable artifact")
```

## Functions

```go
package chainresolver

// ChainResolve returns the chain for the given target logical name.
// The chain contains ancestors (root down to but not including the target),
// dependencies (sorted alphabetically by logical name), the target itself,
// and optionally the target's input.
func ChainResolve(targetLogicalName string) (*Chain, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target:", chain.Target.UnqualifiedLogicalName)
	fmt.Println("Target file:", chain.Target.FilePath.Value)

	fmt.Println("Ancestors:")
	for _, a := range chain.Ancestors {
		fmt.Println(" ", a.UnqualifiedLogicalName, a.FilePath.Value)
	}

	fmt.Println("Dependencies:")
	for _, d := range chain.Dependencies {
		qualifier := ""
		if d.Qualifier != nil {
			qualifier = "(" + *d.Qualifier + ")"
		}
		fmt.Println(" ", d.UnqualifiedLogicalName+qualifier, d.FilePath.Value)
	}

	if chain.Input != nil {
		fmt.Println("Input:", chain.Input.UnqualifiedLogicalName, chain.Input.FilePath.Value)
	}
}
```


## Package

```go
package mcploadchain
```

## Import Path

```
github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain
```

## Error Sentinels

```go
package mcploadchain

import "errors"

var ErrNoOutput          = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")
```

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the chain for the given logical name and returns
// a single string containing the chain hash and all context sections.
//
// The returned string uses the following format:
//
//	chain_hash: <27-character hash>
//	--- context ---
//	<context content>
//	--- input ---
//	<input content>
//	--- existing artifact ---
//	<existing artifact content>
//
// The "--- input ---" section is only present when the target node's
// frontmatter has a non-empty input field.
//
// The "--- existing artifact ---" section is only present when the output
// file exists on disk and is readable.
//
// Errors:
//   - ErrNoOutput: target node has no output field.
//   - ErrInvalidOutputPath: the output path fails path validation.
//   - Propagated from LogicalNameToPath, ChainResolve, ChainHashCompute,
//     NodeParse, and FileOpen.
func MCPLoadChain(logicalName string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcploadchain"
)

func main() {
	result, err := mcploadchain.MCPLoadChain("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```


# Package `filereader`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader`

## Struct Definitions

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileReader holds the state for sequential line-by-line reading of a file.
// The caller must call FileClose when done to release the underlying file handle.
type FileReader struct {
	CfsPath pathutils.PathCfs
}
```

## Error Sentinels

```go
package filereader

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrEndOfFile      = errors.New("end of file")
```

## Function Signatures

```go
package filereader

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FileOpen opens the file at cfsPath and prepares it for sequential
// line-by-line reading from the beginning. The caller must call FileClose
// when done — failing to do so leaks the file handle.
func FileOpen(cfsPath pathutils.PathCfs) (*FileReader, error)

// FileReadLine reads the next line from the reader, normalizes CRLF to LF,
// and returns the line without the line terminator. Returns ErrEndOfFile
// when there are no more lines, or after FileClose has been called.
func FileReadLine(reader *FileReader) (string, error)

// FileSkipLines reads and discards count lines from the reader without
// returning their content. Does nothing if the reader has been closed.
func FileSkipLines(reader *FileReader, count int)

// FileClose releases the file resource held by reader. After FileClose,
// FileReadLine returns ErrEndOfFile and FileSkipLines does nothing.
func FileClose(reader *FileReader)
```

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	cfsPath := pathutils.PathCfs{Value: "SPEC/myproject/some_spec.md"}

	reader, err := filereader.FileOpen(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer filereader.FileClose(reader)

	filereader.FileSkipLines(reader, 2)

	for {
		line, err := filereader.FileReadLine(reader)
		if errors.Is(err, filereader.ErrEndOfFile) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(line)
	}
}
```


# Package `pathutils`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils`

## Types

```go
package pathutils

// PathCfs is a path in the Code from Spec standard format:
// forward-slash separated, relative to the project root,
// no ".." components, no drive letters, no leading "/", no backslashes.
type PathCfs struct {
	Value string
}

// PathOs is an absolute path in the operating system's native format.
// This type is never exposed in the framework's public API.
type PathOs struct {
	Value string
}
```

## Error Sentinels

```go
package pathutils

import "errors"

var ErrCannotDetermineRoot   = errors.New("cannot determine project root")
var ErrPathEmpty             = errors.New("path is empty")
var ErrPathAbsolute          = errors.New("path must not be absolute")
var ErrPathContainsBackslash = errors.New("path must not contain backslashes")
var ErrDirectoryTraversal    = errors.New("path contains directory traversal components")
var ErrResolvesOutsideRoot   = errors.New("path resolves outside the project root")
```

## Functions

```go
package pathutils

// PathGetProjectRoot returns the project root as a PathOs,
// determined from the working directory of the process.
func PathGetProjectRoot() (*PathOs, error)

// PathValidateCfs validates that value conforms to the PathCfs format rules.
// Returns an error describing the violation if the value is not valid.
// Does not verify that the file exists or resolve symlinks.
func PathValidateCfs(value string) error

// PathCfsToOs validates cfs_path and converts it to an absolute PathOs.
// This is the single entry point for going from framework paths to OS paths.
// The target file or directory does not need to exist.
func PathCfsToOs(cfsPath *PathCfs) (*PathOs, error)

// PathOsToCfs converts an absolute PathOs to a PathCfs relative to the
// project root. Used internally by components that receive paths from the OS.
// The target file or directory does not need to exist.
func PathOsToCfs(osPath *PathOs) (*PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	root, err := pathutils.PathGetProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Project root:", root.Value)

	cfsPath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/_node.md"}

	if err := pathutils.PathValidateCfs(cfsPath.Value); err != nil {
		log.Fatal(err)
	}

	osPath, err := pathutils.PathCfsToOs(cfsPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OS path:", osPath.Value)

	roundTripped, err := pathutils.PathOsToCfs(osPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("CFS path:", roundTripped.Value)
}
```


# Package `frontmatter`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter`

## Types

```go
package frontmatter

// Frontmatter holds the parsed fields extracted from a spec node file's
// YAML front matter block.
type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}
```

## Error Sentinels

```go
package frontmatter

import "errors"

var ErrFileUnreadable = errors.New("file unreadable")
var ErrMalformedYAML  = errors.New("malformed YAML")
```

## Functions

```go
package frontmatter

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// FrontmatterParse opens the file at filePath, extracts the YAML front matter
// delimited by "---" markers, and returns the parsed Frontmatter.
// All fields default to their zero value (empty list, empty string) when
// absent from the YAML block.
func FrontmatterParse(filePath *pathutils.PathCfs) (*Frontmatter, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	path := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/_node.md"}

	fm, err := frontmatter.FrontmatterParse(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:", fm.Output)
	fmt.Println("Input:", fm.Input)
	fmt.Println("DependsOn:", fm.DependsOn)
}
```


# Package `parsenode`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode`

## Types

```go
package parsenode

// NodeSubsection represents a level-2 heading block within a section.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading block within a node file.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node holds the parsed structure of a node file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}
```

## Error Sentinels

```go
package parsenode

import "errors"

var ErrNotASpecReference                    = errors.New("logical name is not a SPEC/ reference")
var ErrHasQualifier                         = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                       = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading  = errors.New("file body has non-blank content before the first level-1 heading, or has no level-1 heading at all")
var ErrNodeNameDoesNotMatch                 = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection               = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection                = errors.New("more than one Agent section exists")
var ErrDuplicatePrivateSection              = errors.New("more than one Private section exists")
var ErrUnrecognizedSection                  = errors.New("unrecognized level-1 heading")
var ErrDuplicateSubsection                  = errors.New("two level-2 headings within the same section normalize to the same text")
```

## Functions

```go
package parsenode

// NodeParse reads and parses the node file for the given logical name.
// The logical name must be a SPEC/ reference and must not contain a parenthetical qualifier.
// Returns a fully populated Node on success, or an error describing the first violation found.
func NodeParse(logicalName string) (*Node, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
)

func main() {
	node, err := parsenode.NodeParse("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Node name section heading:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("Public section has", len(node.Public.Subsections), "subsection(s)")
		for _, sub := range node.Public.Subsections {
			fmt.Println("  Subsection:", sub.Heading)
		}
	}

	if node.Agent != nil {
		fmt.Println("Agent section content lines:", len(node.Agent.Content))
	}

	if node.Private != nil {
		fmt.Println("Private section content lines:", len(node.Private.Content))
	}
}
```


# Package `logicalnames`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames`

## Error Sentinels

```go
package logicalnames

import "errors"

var ErrUnsupportedReference   = errors.New("logical name is not a SPEC/ reference")
var ErrInvalidPath             = errors.New("path is not a _node.md file under code-from-spec/")
var ErrNoParent                = errors.New("logical name is SPEC itself")
var ErrNotASpecReference       = errors.New("logical name is not a SPEC/ reference")
var ErrNotAnArtifactReference  = errors.New("logical name does not start with ARTIFACT/")
var ErrNotAnExternalReference  = errors.New("logical name does not start with EXTERNAL/")
```

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// LogicalNameToPath converts a SPEC/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts SPEC/ references (including SPEC itself).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the SPEC/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a SPEC/ reference.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
// Only accepts SPEC/ references (including SPEC itself, which returns ErrNoParent).
// Always returns a SPEC/ reference.
func LogicalNameGetParent(logicalName string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Returns empty string and false if no qualifier is present.
// Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameGetQualifier(logicalName string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameStripQualifier(logicalName string) string

// LogicalNameHasParent returns true if the logical name is a SPEC/ reference
// other than SPEC itself. Returns false for SPEC, ARTIFACT/, EXTERNAL/,
// and unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameHasQualifier(logicalName string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool

// LogicalNameIsSpec returns true if the logical name is exactly SPEC or
// starts with SPEC/.
func LogicalNameIsSpec(logicalName string) bool

// LogicalNameIsExternal returns true if the logical name starts with EXTERNAL/.
func LogicalNameIsExternal(logicalName string) bool

// LogicalNameGetArtifactGenerator returns the SPEC/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and
// prepends SPEC/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error)

// LogicalNameExternalToPath converts an EXTERNAL/ logical name to a PathCfs.
// Strips the EXTERNAL/ prefix and returns the remainder as a PathCfs
// relative to the project root.
func LogicalNameExternalToPath(logicalName string) (*pathutils.PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	specName := "SPEC/payments/fees"

	nodePath, err := logicalnames.LogicalNameToPath(specName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node path:", nodePath.Value)

	roundTripped, err := logicalnames.LogicalNameFromPath(nodePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", roundTripped)

	parent, err := logicalnames.LogicalNameGetParent(specName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent)

	qualified := "SPEC/payments/fees(summary)"

	qualifier, ok := logicalnames.LogicalNameGetQualifier(qualified)
	if ok {
		fmt.Println("Qualifier:", qualifier)
	}

	stripped := logicalnames.LogicalNameStripQualifier(qualified)
	fmt.Println("Stripped:", stripped)

	fmt.Println("Has parent:", logicalnames.LogicalNameHasParent(specName))
	fmt.Println("Has qualifier:", logicalnames.LogicalNameHasQualifier(qualified))
	fmt.Println("Is spec:", logicalnames.LogicalNameIsSpec(specName))
	fmt.Println("Is artifact:", logicalnames.LogicalNameIsArtifact("ARTIFACT/payments/fees"))
	fmt.Println("Is external:", logicalnames.LogicalNameIsExternal("EXTERNAL/proto/v1/api.proto"))

	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/payments/fees")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator:", generator)

	extPath, err := logicalnames.LogicalNameExternalToPath("EXTERNAL/proto/v1/api.proto")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("External path:", extPath.Value)

	_ = &pathutils.PathCfs{}
}
```


# Package `textnormalization`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization`

## Overview

Package `textnormalization` provides utilities for normalizing text strings by trimming whitespace, collapsing internal spaces, and converting to lowercase. Unicode characters that have a simple lowercase folding (e.g., `Straße` → `strasse`) are handled via standard library Unicode normalization.

## Function Signatures

```go
package textnormalization

// NormalizeText trims leading and trailing whitespace, collapses all
// internal whitespace sequences to a single space, and converts the
// result to lowercase. Unicode characters with a simple lowercase
// equivalent (e.g., "Straße" → "strasse") are folded accordingly.
// An empty string returns an empty string.
func NormalizeText(rawString string) string
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/textnormalization"
)

func main() {
	fmt.Println(textnormalization.NormalizeText("  Interface  "))
	fmt.Println(textnormalization.NormalizeText("PUBLIC"))
	fmt.Println(textnormalization.NormalizeText("Straße"))
	fmt.Println(textnormalization.NormalizeText("Testes   de   aceitação"))
	fmt.Println(textnormalization.NormalizeText(""))
}
```

Expected output:

```
interface
public
strasse
testes de aceitação

```

---
output: code-from-spec/golang/implementation/mcp_tools/load_chain_debug/chain.md
---

# Agent
Save the entire content you received from `load_chain`
to the output file, verbatim. Place the artifact tag
as the first line, inside a markdown comment.
--- input ---

namespace: mcploadchain


function MCPLoadChain(logical_name: string) -> string
  errors:
    - NoOutput: target node has no output field.
    - InvalidOutputPath: the output path fails path validation.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (NodeParsing.*): propagated from NodeParse.
    - (FileReader.*): propagated from FileOpen.

  1. Call `LogicalNameToPath(logical_name)` to get the target node's file path.
     If it fails, propagate the error.

  2. Call `FrontmatterParse(target_file_path)` to read the target node's frontmatter.
     If `frontmatter.output` is empty, raise error "NoOutput".
     Call `PathValidateCfs(frontmatter.output)`.
     If it fails, raise error "InvalidOutputPath".

  3. Call `ChainResolve(logical_name)` to get the resolved `Chain`.
     If it fails, propagate the error.

  4. Call `ChainHashCompute(chain)` with the resolved chain.
     If it fails, propagate the error.
     Store the result as `chain_hash`.

  5. Build the context stream:

     Set `context_parts` to an empty list of strings.

     For each `ancestor` in `chain.ancestors` (in order):
       Call `NodeParse(ancestor.unqualified_logical_name)`.
       If `node.public` is absent or `node.public.subsections` is empty, skip.
       Otherwise:
         Build `block` by concatenating all subsections in document order:
           For each subsection in `node.public.subsections`:
             Add the subsection `raw_heading` (trailing whitespace removed).
             Add each line in `subsection.content` with leading blank lines
               after the heading removed and trailing blank lines removed.
             Ensure the block ends with exactly one LF.
           Separate consecutive subsection blocks with exactly one blank line.
         Append `block` to `context_parts`.

     For each `dep` in `chain.dependencies` (in order):
       If `LogicalNameIsArtifact(dep.unqualified_logical_name)` is true:
         Call `FileOpen(dep.file_path)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Skip the first line that contains "code-from-spec:" (the artifact tag line).
         Include all other lines.
         Call `FileClose`.
         Append the resulting text to `context_parts`.
       Else if `LogicalNameIsExternal(dep.unqualified_logical_name)` is true:
         Call `FileOpen(dep.file_path)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Call `FileClose`.
         Append the full file content to `context_parts`.
       Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` is true and `dep.qualifier` is absent:
         Call `NodeParse(dep.unqualified_logical_name)`.
         If `node.public` is absent or `node.public.subsections` is empty, skip.
         Otherwise:
           Build `block` by concatenating all subsections in document order
             (same boundary normalization rules as for ancestors).
           Append `block` to `context_parts`.
       Else if `LogicalNameIsSpec(dep.unqualified_logical_name)` is true and `dep.qualifier` is present:
         Call `NodeParse(dep.unqualified_logical_name)`.
         Compute `normalized_qualifier` = `NormalizeText(dep.qualifier)`.
         Find the subsection in `node.public.subsections` whose `heading` equals `normalized_qualifier`.
         If found:
           Build `block` from the subsection `raw_heading` (trailing whitespace removed)
             and its content (leading blank lines removed, trailing blank lines removed,
             ends with exactly one LF).
           Append `block` to `context_parts`.

     For the target node `chain.target`:
       Build a reduced frontmatter block:
         Line 1: "---"
         Line 2: "output: <frontmatter.output>"
         Line 3: "---"
       Append this block to `context_parts`.

       Call `NodeParse(chain.target.unqualified_logical_name)`.
       If `node.public` is present and `node.public.subsections` is non-empty:
         Build `block` by concatenating all subsections in document order
           (same boundary normalization rules as above).
         Append `block` to `context_parts`.

       If `node.agent` is present:
         Build `agent_block`:
           Add `node.agent.raw_heading` (trailing whitespace removed).
           Add each line in `node.agent.content`
             (leading blank lines removed, trailing blank lines removed).
           For each subsection in `node.agent.subsections`:
             Separate from previous block with exactly one blank line.
             Add the subsection `raw_heading` (trailing whitespace removed).
             Add each line in `subsection.content`
               (leading blank lines removed, trailing blank lines removed).
           Ensure the block ends with exactly one LF.
         Append `agent_block` to `context_parts`.

  6. Assemble the output string:

     Start with line: "chain_hash: <chain_hash>"
     Append line: "--- context ---"
     Append the context stream: join all entries in `context_parts`
       separated by exactly one blank line.

     If `chain.input` is present:
       Append line: "--- input ---"
       If `LogicalNameIsArtifact(chain.input.unqualified_logical_name)` is true:
         Call `FileOpen(chain.input.file_path)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Skip the first line that contains "code-from-spec:".
         Include all other lines.
         Call `FileClose`.
         Append the resulting text.
       Else (EXTERNAL/ or other):
         Call `FileOpen(chain.input.file_path)`.
         Read all lines with `FileReadLine` until `EndOfFile`.
         Call `FileClose`.
         Append the full file content.

     If the file at `frontmatter.output` exists and is readable:
       Append line: "--- existing artifact ---"
       Call `FileOpen` with the PathCfs of `frontmatter.output`.
       Read all lines with `FileReadLine` until `EndOfFile`.
       Call `FileClose`.
       Append the full file content.
       If the file does not exist or cannot be read, omit this section silently.

  7. Return the assembled output string.
