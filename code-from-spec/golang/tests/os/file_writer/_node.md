---
depends_on:
  - ARTIFACT/golang/interfaces/os/file_writer(interface)
input: ARTIFACT/functional/tests/os/file_writer(file_writer_tests)
outputs:
  - id: filewriter_test
    path: internal/filewriter/filewriter_test.go
---

# ROOT/golang/tests/os/file_writer
