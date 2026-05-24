# ROOT/domain

MCP server for Code from Spec projects. Provides tools for
spec validation, artifact generation, and artifact management.

# Private

## Context

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
