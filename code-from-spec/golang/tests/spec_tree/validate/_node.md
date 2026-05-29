---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/utils/text_normalization(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
input: ARTIFACT/functional/tests/spec_tree/validate(format_validation_tests)
outputs:
  - id: spectreevalidate_test
    path: internal/spectreevalidate/spectreevalidate_test.go
---

# ROOT/golang/tests/spec_tree/validate
