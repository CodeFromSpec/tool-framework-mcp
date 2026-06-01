---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/scan(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/list_files(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/spec_tree/scan(spec_tree_tests)
outputs:
  - id: spectree_test
    path: internal/spectree/spectree_test.go
---

# ROOT/golang/tests/spec_tree/scan
