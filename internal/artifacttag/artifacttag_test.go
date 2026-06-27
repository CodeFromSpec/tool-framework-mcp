// code-from-spec: SPEC/golang/tests/parsing/artifact_tag@AIk5dXsXveMQapj03hGtyPlO5zw
package artifacttag_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/file"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func testChdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("testChdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("testChdir cleanup: %v", err)
		}
	})
}

func TestArtifactTagExtract(t *testing.T) {
	type testCase struct {
		name            string
		fileContent     *string
		filePath        string
		wantLogicalName string
		wantHash        string
		wantErr         error
	}

	strPtr := func(s string) *string { return &s }

	tests := []testCase{
		{
			name:            "TC-01: extracts tag from slash-slash comment",
			fileContent:     strPtr("// code-from-spec: ROOT/golang/implementation/internal/foo/code(bar)@abcdefghijklmnopqrstuvwxyza\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/golang/implementation/internal/foo/code(bar)",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			name:            "TC-02: extracts tag from hash comment",
			fileContent:     strPtr("# code-from-spec: ROOT/some/node(id)@123456789012345678901234567\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/some/node(id)",
			wantHash:        "123456789012345678901234567",
		},
		{
			name:            "TC-03: extracts tag from HTML comment",
			fileContent:     strPtr("<!-- code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza -->\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/docs/readme",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			name:            "TC-04: stops reading at first match",
			fileContent:     strPtr("// code-from-spec: ROOT/first/node@abcdefghijklmnopqrstuvwxyza\n// code-from-spec: ROOT/second/node@abcdefghijklmnopqrstuvwxyzb\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/first/node",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			name:            "TC-05: tag on non-first line",
			fileContent:     strPtr("line one\nline two\n// code-from-spec: ROOT/docs/readme@abcdefghijklmnopqrstuvwxyza\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/docs/readme",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			name:            "TC-06: extra whitespace before logical name",
			fileContent:     strPtr("// code-from-spec:   ROOT/x(y)@abcdefghijklmnopqrstuvwxyza\n"),
			filePath:        "artifact.txt",
			wantLogicalName: "ROOT/x(y)",
			wantHash:        "abcdefghijklmnopqrstuvwxyza",
		},
		{
			name:        "TC-07: empty file",
			fileContent: strPtr(""),
			filePath:    "artifact.txt",
			wantErr:     artifacttag.ErrNoTagFound,
		},
		{
			name:     "TC-08: file does not exist",
			filePath: "nonexistent/file.txt",
			wantErr:  file.ErrFileUnreadable,
		},
		{
			name:     "TC-09: propagates path errors",
			filePath: "../../outside",
			wantErr:  pathutils.ErrDirectoryTraversal,
		},
		{
			name:        "TC-10: no tag in file",
			fileContent: strPtr("This file has no artifact tag.\nJust regular content.\n"),
			filePath:    "artifact.txt",
			wantErr:     artifacttag.ErrNoTagFound,
		},
		{
			name:        "TC-11: malformed tag — no @ separator",
			fileContent: strPtr("// code-from-spec: ROOT/foo/bar\n"),
			filePath:    "artifact.txt",
			wantErr:     artifacttag.ErrMalformedTag,
		},
		{
			name:        "TC-12: malformed tag — empty logical name",
			fileContent: strPtr("// code-from-spec: @abcdefghijklmnopqrstuvwxyza\n"),
			filePath:    "artifact.txt",
			wantErr:     artifacttag.ErrMalformedTag,
		},
		{
			name:        "TC-13: malformed tag — wrong hash length",
			fileContent: strPtr("// code-from-spec: ROOT/foo(bar)@short\n"),
			filePath:    "artifact.txt",
			wantErr:     artifacttag.ErrMalformedTag,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testChdir(t, tempDir)

			if tc.fileContent != nil {
				if err := os.WriteFile(tc.filePath, []byte(*tc.fileContent), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			}

			tag, err := artifacttag.ArtifactTagExtract(&pathutils.PathCfs{Value: tc.filePath})

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tag == nil {
				t.Fatal("expected non-nil ArtifactTag, got nil")
			}
			if tag.LogicalName != tc.wantLogicalName {
				t.Errorf("LogicalName: got %q, want %q", tag.LogicalName, tc.wantLogicalName)
			}
			if tag.Hash != tc.wantHash {
				t.Errorf("Hash: got %q, want %q", tag.Hash, tc.wantHash)
			}
		})
	}
}
