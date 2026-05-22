# ROOT/functional/tools

Shared behavior for all tools exposed by the MCP server.

# Public

## Response format

### Success

A text response containing the result.

### Error

A text response marked as an error, containing an actionable
message that identifies what went wrong and what the caller
can do about it.

## Contracts

- Tools are stateless — each call resolves its own inputs
  independently.
- Multiple tools may execute concurrently.
- Expected error conditions are returned as tool errors, not
  as fatal failures.
