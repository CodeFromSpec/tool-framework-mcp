---
depends_on:
  - ROOT/golang/internal/frontmatter
input: ARTIFACT/golang/internal/frontmatter/code(frontmatter)
outputs:
  - id: frontmatter_test
    path: internal/frontmatter/frontmatter_test.go
---

# ROOT/golang/internal/frontmatter/tests

Test cases for the frontmatter package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated
temporary directory. Test files are created with
controlled frontmatter content. `ParseFrontmatter` is
called with the path to each test file.

## Happy Path

### Parses complete frontmatter (all fields)

Create a file with all fields:

```
---
depends_on:
  - ROOT/other
  - ROOT/architecture/backend
external:
  - path: CODE_FROM_SPEC.md
    fragments:
      - description: v3 format
        lines: "10-25"
        hash: abc123
input: ARTIFACT/some/artifact(id)
outputs:
  - id: config
    path: internal/config/config.go
  - id: config_test
    path: internal/config/config_test.go
---
```

Expect:
- `DependsOn` = `["ROOT/other", "ROOT/architecture/backend"]`
- `External` has one entry with `Path` = `"CODE_FROM_SPEC.md"` and one fragment
- `Input` = `"ARTIFACT/some/artifact(id)"`
- `Outputs` has two entries:
  - `ID` = `"config"`, `Path` = `"internal/config/config.go"`
  - `ID` = `"config_test"`, `Path` = `"internal/config/config_test.go"`

### Parses frontmatter with only outputs

Create a file with only `outputs`:

```
---
outputs:
  - id: main
    path: cmd/main.go
---
```

Expect:
- `DependsOn` = nil
- `External` = nil
- `Input` = `""`
- `Outputs` has one entry: `ID` = `"main"`, `Path` = `"cmd/main.go"`
- No error.

### Parses frontmatter with only depends_on

Create a file with only `depends_on`:

```
---
depends_on:
  - ROOT/other/node
  - ROOT/another/node
---
```

Expect:
- `DependsOn` = `["ROOT/other/node", "ROOT/another/node"]`
- `External` = nil
- `Input` = `""`
- `Outputs` = nil
- No error.

### Parses frontmatter with external and fragments

Create a file with `external` entries including fragments:

```
---
external:
  - path: CODE_FROM_SPEC.md
    fragments:
      - description: frontmatter format
        lines: "10-20"
        hash: def456
      - description: tree structure
        lines: "30-50"
        hash: ghi789
  - path: README.md
---
```

Expect:
- `External` has two entries.
- First entry: `Path` = `"CODE_FROM_SPEC.md"`, `Fragments` has two items with correct `Description`, `Lines`, and `Hash`.
- Second entry: `Path` = `"README.md"`, `Fragments` = nil.

### Parses frontmatter with input field

Create a file with `input`:

```
---
input: ARTIFACT/golang/internal/frontmatter(frontmatter)
---
```

Expect:
- `Input` = `"ARTIFACT/golang/internal/frontmatter(frontmatter)"`
- All other fields are nil/empty.
- No error.

### Ignores unknown frontmatter fields

Create a file with extra fields:

```
---
depends_on:
  - ROOT/other
some_future_field: hello
another: 42
---
```

Expect no error. Known fields parsed correctly.
Unknown fields ignored.

### File with no frontmatter returns empty Frontmatter

Create a file with no `---` at all:

```
Just some text without frontmatter.
```

Expect:
- No error.
- Result is an empty `Frontmatter` (all fields nil/zero).

## Edge Cases

### Empty frontmatter

Create a file with:

```
---
---
```

Expect no error. Result is an empty `Frontmatter` (all fields nil/zero).

### File with only frontmatter, nothing after

Create a file with:

```
---
depends_on:
  - ROOT/other
---
```

Expect no error. `DependsOn` = `["ROOT/other"]`. Body is not read.

## Failure Cases

### File does not exist

Call `ParseFrontmatter` with a non-existent path.
Expect `errors.Is(err, ErrRead)`.

### Malformed YAML in frontmatter

Create a file with invalid YAML between delimiters:

```
---
depends_on: [invalid
---
```

Expect `errors.Is(err, ErrFrontmatterParse)`.
