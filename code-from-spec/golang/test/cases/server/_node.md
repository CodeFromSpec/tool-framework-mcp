---
depends_on:
  - SPEC/golang/dependencies/mcp-go-sdk
input: ARTIFACT/golang/implementation/server
output: cmd/framework-mcp/main_test.go
---

# SPEC/golang/test/cases/server

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

The MCP stdio transport uses newline-delimited JSON —
each JSON-RPC message is one line, no embedded newlines.
The handshake sequence is:

1. Send `initialize` request (with `id`, `method`,
   `params` including `protocolVersion` and `clientInfo`).
2. Read the `initialize` response from stdout.
3. Send `notifications/initialized` notification (with
   `method` only, no `id` — it is a notification).
4. Now send further requests (e.g. `tools/list`).

All test cases in this section must follow this
handshake before sending any request.

### tools/list advertises maxResultSizeChars for load_chain

Start the binary as a subprocess. Complete the MCP
handshake, then send a `tools/list` request. Parse
the JSON-RPC response from stdout.

Expect: the response contains a tool named `load_chain`
with `_meta["anthropic/maxResultSizeChars"]` equal to
`500000`.

### tools/list advertises all tools

Start the binary as a subprocess. Complete the MCP
handshake, then send a `tools/list` request. Parse
the JSON-RPC response from stdout.

Expect: the response contains tools named `load_chain`,
`write_file`, `validate_specs`, `accept`,
`dump_chain`, `reconstruct_cache`, `prune_cache`,
and `version`.

### version tool returns version string

Start the binary as a subprocess. Complete the MCP
handshake, then send a `tools/call` request for the
`version` tool (no arguments).

Expect: the response contains a text content with at
least the default version string `"dev"`.
