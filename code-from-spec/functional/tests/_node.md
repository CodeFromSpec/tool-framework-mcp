# ROOT/functional/tests

Functional test specifications — test cases, inputs,
expected outputs, and error conditions for each component.
Language-agnostic: describes what to test, not how.

# Public

## Rules

- Use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: setup (if needed), actions (what
  functions to call), and expected outcome.
- Do not prescribe how to create test files, temp
  directories, or assert results — those are
  implementation details for the language layer.
