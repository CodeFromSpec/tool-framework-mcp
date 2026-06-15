// code-from-spec: SPEC/golang/tests/parsing/frontmatter@fYPwti2Vrgj01U19aH68YUr9cSI
package frontmatter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"
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

func TestFrontmatterParse_TC1_AllFields(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\ndepends_on:\n  - \"SPEC/some/node\"\n  - \"ARTIFACT/some/output\"\n  - \"EXTERNAL/proto/api.proto\"\ninput: \"some/input/path\"\noutput: \"some/output/path\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantDeps := []string{"SPEC/some/node", "ARTIFACT/some/output", "EXTERNAL/proto/api.proto"}
	if len(fm.DependsOn) != len(wantDeps) {
		t.Fatalf("DependsOn length: got %d, want %d", len(fm.DependsOn), len(wantDeps))
	}
	for i, dep := range wantDeps {
		if fm.DependsOn[i] != dep {
			t.Errorf("DependsOn[%d]: got %q, want %q", i, fm.DependsOn[i], dep)
		}
	}
	if fm.Input != "some/input/path" {
		t.Errorf("Input: got %q, want %q", fm.Input, "some/input/path")
	}
	if fm.Output != "some/output/path" {
		t.Errorf("Output: got %q, want %q", fm.Output, "some/output/path")
	}
}

func TestFrontmatterParse_TC2_OnlyOutput(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\noutput: \"only/output/path\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "only/output/path" {
		t.Errorf("Output: got %q, want %q", fm.Output, "only/output/path")
	}
}

func TestFrontmatterParse_TC3_OnlyDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\ndepends_on:\n  - \"SPEC/some/dep\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "SPEC/some/dep" {
		t.Errorf("DependsOn: got %v, want [SPEC/some/dep]", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_TC4_ExternalInDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\ndepends_on:\n  - \"EXTERNAL/proto/api.proto\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("DependsOn: got %v, want [EXTERNAL/proto/api.proto]", fm.DependsOn)
	}
}

func TestFrontmatterParse_TC5_OnlyInput(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\ninput: \"some/input/file\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "some/input/file" {
		t.Errorf("Input: got %q, want %q", fm.Input, "some/input/file")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_TC6_IgnoresUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\noutput: \"some/output/path\"\ncustom_field: \"unexpected value\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "some/output/path" {
		t.Errorf("Output: got %q, want %q", fm.Output, "some/output/path")
	}
}

func TestFrontmatterParse_TC7_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "# Just a heading\n\nSome body content.\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_TC8_EmptyFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_TC9_OnlyFrontmatterNoBody(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\noutput: \"some/output/path\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "some/output/path" {
		t.Errorf("Output: got %q, want %q", fm.Output, "some/output/path")
	}
}

func TestFrontmatterParse_TC10_DelimiterWithTrailingWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---   \noutput: \"some/output/path\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("Output: got %q, want empty", fm.Output)
	}
}

func TestFrontmatterParse_TC11_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFrontmatterParse_TC12_InvalidPathDirectoryTraversal(t *testing.T) {
	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestFrontmatterParse_TC13_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\ndepends_on: [unclosed bracket\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_TC14_UnclosedFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\noutput: \"some/output/path\"\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_TC15_ExternalFieldIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "---\noutput: \"some/output/path\"\nexternal: \"some/external/ref\"\n---\n"
	if err := os.WriteFile("file.md", []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Output != "some/output/path" {
		t.Errorf("Output: got %q, want %q", fm.Output, "some/output/path")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn: got %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input: got %q, want empty", fm.Input)
	}
}
