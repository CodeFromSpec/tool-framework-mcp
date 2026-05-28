---
depends_on:
  - ARTIFACT/golang/interfaces/utils/spec_tree(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/os/list_files(interface)
  - ARTIFACT/golang/interfaces/utils/logical_names(interface)
input: ARTIFACT/functional/tests/utils/spec_tree(spec_tree_tests)
outputs:
  - id: spectree_test
    path: internal/spectree/spectree_test.go
---

# ROOT/golang/tests/utils/spec_tree
