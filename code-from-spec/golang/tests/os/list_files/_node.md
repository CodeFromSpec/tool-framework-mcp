---
depends_on:
  - ARTIFACT/golang/interfaces/os/list_files(interface)
input: ARTIFACT/functional/tests/os/list_files(list_files_tests)
outputs:
  - id: listfiles_test
    path: internal/listfiles/listfiles_test.go
---

# ROOT/golang/tests/os/list_files
