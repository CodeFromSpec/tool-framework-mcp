# ROOT/functional

MCP server for Code from Spec projects. Provides tools for
spec validation, artifact generation, and artifact management.

Functional specifications for each component — inputs, outputs,
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

### Style rules

- No programming language syntax — no `func`, `def`, `class`,
  `struct`, `interface`, `nil`, `null`, `None`.
- Use `"quoted text"` for literal strings (error messages, etc.).
- Use `<angle brackets>` for placeholders.
- Use indentation for nesting, not braces or keywords.
- Describe data structures as records with named fields.

### Example

```
function ParseFrontmatter(file_path) -> frontmatter

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

This server is the single point of interaction between agents
and the spec tree. It controls what agents can read and where
they can write, making the correct workflow the only possible
workflow.

For code generation, this means confinement: given unrestricted
file access, a subagent will compensate for perceived gaps in
its context by exploring the repository rather than stopping to
report ambiguity. This produces hallucinated or inconsistent
output. The server prevents this by exposing only the assembled
chain and restricting writes to declared outputs.

For validation, the server provides a single tool that checks
the entire spec tree for format errors, circular references,
and artifact staleness — replacing the need for a separate
external tool.

## Contracts

### Invocation

```
framework-mcp
```

Any argument causes the tool to print a usage message and exit.
`--help`, `-h`, and `help` exit 0; any other argument exits 1.

### Project root

The working directory of the MCP server process is the
project root. All relative paths — `outputs`, `external`,
artifact file paths — are resolved relative to this
directory. The server does not search for the project root;
it is the caller's responsibility to start the server from
the correct location.

### Distribution

The binary may be placed inside the host project repository at a
path chosen by that project. No installation on the machine is
required.

### Deployment

The server is registered once in the project's Claude Code
configuration (`.claude/settings.json`):

```json
{
  "mcpServers": {
    "framework-mcp": {
      "type": "stdio",
      "command": "<path-to-framework-mcp>"
    }
  }
}
```

Once configured, the server is available to all sessions and
subagents in that project. No per-invocation setup or teardown
is needed.

### Concurrency

Multiple instances may run in parallel without conflict. Each is
an independent OS process with its own state.

## Tools

| Tool | Purpose |
|---|---|
| `load_chain` | Load the spec chain for a node, including the chain hash |
| `write_file` | Write a generated file to disk, validated against `outputs` |
| `validate_specs` | Validate format, circular references, and artifact staleness |
| `hash_fragment` | Calculate hash of a file line range for `external:` fragments |

## Decisions

### Confinement is the caller's responsibility

The server exposes every tool it has to every connection. If a
subagent should only use a subset, the orchestrator must enforce
that by configuring the subagent itself — not by asking the server
to hide tools. This keeps the server simple and the tool surface
predictable.

### Minimal tool surface

Purpose-built tools combined with caller-side restriction of
which tools a subagent can call constrain the agent's action
space, making correct behavior more likely by construction.

### Validation is built in

The `validate_specs` tool replaces the need for a separate
`staleness-check` binary. Having validation inside the same
server simplifies deployment — one binary instead of two.
