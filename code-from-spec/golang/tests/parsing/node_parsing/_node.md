---
depends_on:
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/parsing/node_parsing(node_parsing_tests)
outputs:
  - id: parsenode_test
    path: internal/parsenode/parsenode_test.go
---

# ROOT/golang/tests/parsing/node_parsing
