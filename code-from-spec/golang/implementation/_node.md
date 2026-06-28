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

### Sentinel error names in Agent sections

Agent sections reference errors by their Go sentinel
name (`ErrXxx`) instead of prose descriptions. This
ensures the generated code uses the exact sentinel
declared in the `# Public ## Interface` section and
eliminates ambiguity about which error to return.
