[//]: # (code-from-spec: ROOT/golang/interfaces/os/file_writer@BU-2LwylC7VUKBN6xkikZeF_Hpw)

# Interface: `filewriter`

**Package:** `package filewriter`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"`

---

## Error Sentinels

```go
var (
    // ErrCannotCreateDirectory is returned when an intermediate directory
    // cannot be created during a write operation.
    ErrCannotCreateDirectory = errors.New("cannot create directory")

    // ErrCannotWriteFile is returned when the file cannot be written.
    ErrCannotWriteFile = errors.New("cannot write file")
)
```

---

## Functions

```go
// FileWrite writes content to the file at cfs_path as UTF-8 encoded text.
// If the file exists, it is overwritten. If it does not exist, it is created.
// Intermediate directories are created as needed.
//
// Content is written exactly as received — no normalization of line endings
// or other transformations is applied.
//
// The path is validated before writing. If validation fails, no file or
// directory is created.
//
// Returns an error if:
//   - path validation fails (errors from PathCfsToOs are propagated).
//   - an intermediate directory cannot be created (ErrCannotCreateDirectory).
//   - the file cannot be written (ErrCannotWriteFile).
func FileWrite(cfs_path *pathutils.PathCfs, content string) error
```

---

## Usage Example

```go
package main

import (
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filewriter"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
    // Write content to a file at a CFS path.
    cfsPath := &pathutils.PathCfs{Value: "output/result.txt"}
    content := "Hello, world!"

    if err := filewriter.FileWrite(cfsPath, content); err != nil {
        log.Fatalf("could not write file: %v", err)
    }
}
```
