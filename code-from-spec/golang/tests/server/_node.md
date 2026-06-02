---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
input: ARTIFACT/golang/implementation/server(main)
outputs:
  - id: main_test
    path: cmd/framework-mcp/main_test.go
---

# ROOT/golang/tests/server

Tests for the MCP server entry point.

# Agent

## Context

Tests invoke the compiled binary as a subprocess using
`os/exec` and verify its behavior: exit codes, stderr
output, and stdout output.

The binary is built once in `TestMain` into a temp
directory. On Windows, the binary name must include the
`.exe` extension: use `runtime.GOOS == "windows"` to
detect the platform and append `.exe` to the output
path when building.

## Happy Path

### Help flag prints usage to stdout

Run the binary with `--help`.

Expect: exit 0, stdout contains the usage message.

### Help word prints usage to stdout

Run the binary with `help`.

Expect: exit 0, stdout contains the usage message.

### Short help flag prints usage to stdout

Run the binary with `-h`.

Expect: exit 0, stdout contains the usage message.

## Failure Cases

### Unrecognized argument prints usage to stderr

Run the binary with `something`.

Expect: exit 1, stderr contains the usage message.

### Multiple arguments prints usage to stderr

Run the binary with `foo bar`.

Expect: exit 1, stderr contains the usage message.

## MCP Protocol

### tools/list advertises maxResultSizeChars for load_chain

Start the binary as a subprocess. Send an MCP
`initialize` request followed by a `tools/list` request
over stdin (JSON-RPC). Parse the JSON-RPC response from
stdout.

Expect: the response contains a tool named `load_chain`
with `_meta["anthropic/maxResultSizeChars"]` equal to
`500000`.

### tools/list advertises all tools

Start the binary as a subprocess. Send an MCP
`initialize` request followed by a `tools/list` request
over stdin (JSON-RPC). Parse the JSON-RPC response from
stdout.

Expect: the response contains tools named `load_chain`,
`write_file`, `validate_specs`, `chain_hash`, and
`version`.
