---
depends_on:
  - ARTIFACT/golang/implementation/internal/frontmatter/code(frontmatter)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
input: ARTIFACT/functional/tests/parsing/frontmatter(frontmatter_tests)
outputs:
  - id: frontmatter_test
    path: internal/frontmatter/frontmatter_test.go
---

# ROOT/golang/tests/internal/frontmatter
