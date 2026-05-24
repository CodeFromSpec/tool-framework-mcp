# ROOT/functional

Functional specifications for each component of the framework-mcp
server. Describes what each component does — inputs, outputs,
behavior, error conditions, algorithms — without prescribing
a programming language or implementation technology.

# Public

## Output format

Each leaf node generates a pseudocode file (`output.md`).
The pseudocode is a structured, language-agnostic description
of the component's logic that a downstream implementation
layer can transform into source code.

The generated file must contain:

1. **Artifact tag** — `code-from-spec: <name>@<hash>` as the
   first line, inside a markdown comment (`<!-- -->`).
2. **Function signatures** — name, parameters, return values.
   Use plain descriptive types (`string`, `list of strings`,
   `record`, `optional`), not language-specific types.
3. **Step-by-step logic** — numbered steps for each function.
   Use plain language with control flow keywords:
   `if`, `else`, `for each`, `return`, `raise error`.
4. **Error conditions** — what can go wrong and what to return.
5. **Contracts** — invariants, preconditions, postconditions.

### Style rules

- No programming language syntax — no `func`, `def`, `class`,
  `struct`, `interface`, `nil`, `null`, `None`.
- Use `"quoted text"` for literal strings (error messages, etc.).
- Use `<angle brackets>` for placeholders.
- Use indentation for nesting, not braces or keywords.
- Describe data structures as records with named fields.

### Example

```
function ParseFrontmatter(file_path) -> (frontmatter, error)

  1. Read the file at file_path.
     If the file cannot be read, raise error "cannot read file".

  2. Look for the first line containing exactly "---".
     If not found, return an empty frontmatter record.

  3. Collect lines until the next "---".
     Parse the collected lines as YAML.
     If parsing fails, raise error "malformed frontmatter".

  4. Extract known fields from the parsed YAML:
     - depends_on: list of strings
     - external: list of external records
     - input: string
     - outputs: list of output records
     Ignore unknown fields.

  5. Return the frontmatter record.
```

## Constraints

Functional specs must be language-agnostic. They describe:
- Input and output formats
- Behavior and algorithms
- Error conditions and expected responses
- Contracts and invariants

They must NOT prescribe:
- Programming language constructs (types, structs, classes)
- Package or module organization
- Library-specific APIs
- Language-specific error handling patterns

# Private

## Context

Each leaf node under this subtree specifies the behavior of one
component. The `golang/` layer consumes the generated pseudocode
to produce Go source code.
