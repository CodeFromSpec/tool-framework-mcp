# SPEC/golang/implementation

Go implementation of the spec tree components.

# Public

## Implementation rules

- Write idiomatic Go: camelCase for local variables and
  parameters, exported names for public API, receiver
  methods where the interface specifies them.
- Wrap all errors with `fmt.Errorf` using `%w` so callers
  can match with `errors.Is()`.
- Write straightforward code. Simple and readable over
  clever and compact.

# Private

## Decisions

### File lock timeout (30 seconds)

All `OpenFile` calls use a timeout of 30000 ms (30
seconds). This is a safety net against deadlocks or
bugs — in normal operation, locks are acquired
instantly because contention is negligible. If a lock
is not acquired within 30 seconds, something is wrong
and failing is better than hanging indefinitely.

### Sentinel error names in Agent sections

Agent sections reference errors by their Go sentinel
name (`ErrXxx`) instead of prose descriptions. This
ensures the generated code uses the exact sentinel
declared in the `# Public ## Interface` section and
eliminates ambiguity about which error to return.
