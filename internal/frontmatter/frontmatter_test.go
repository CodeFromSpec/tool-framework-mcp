// code-from-spec: ROOT/golang/tests/parsing/frontmatter@gDM--fw5q3Gx8lYZT7oin57aKnY
package frontmatter_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
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

func testWriteFile(t *testing.T, name string, content string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestFrontmatterParse_TC01_AllFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - SPEC/payments/fees
  - ARTIFACT/generated/output.go
  - EXTERNAL/proto/api.proto
input: some input value
output: path/to/output.go
---
# Body content
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 3 {
		t.Fatalf("expected 3 depends_on entries, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "SPEC/payments/fees" {
		t.Errorf("expected SPEC/payments/fees, got %s", fm.DependsOn[0])
	}
	if fm.DependsOn[1] != "ARTIFACT/generated/output.go" {
		t.Errorf("expected ARTIFACT/generated/output.go, got %s", fm.DependsOn[1])
	}
	if fm.DependsOn[2] != "EXTERNAL/proto/api.proto" {
		t.Errorf("expected EXTERNAL/proto/api.proto, got %s", fm.DependsOn[2])
	}
	if fm.Input != "some input value" {
		t.Errorf("expected input 'some input value', got %q", fm.Input)
	}
	if fm.Output != "path/to/output.go" {
		t.Errorf("expected output 'path/to/output.go', got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC02_OnlyOutput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: path/to/output.go
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "path/to/output.go" {
		t.Errorf("expected output 'path/to/output.go', got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC03_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - SPEC/a
  - SPEC/b
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on entries, got %d", len(fm.DependsOn))
	}
	if fm.DependsOn[0] != "SPEC/a" {
		t.Errorf("expected SPEC/a, got %s", fm.DependsOn[0])
	}
	if fm.DependsOn[1] != "SPEC/b" {
		t.Errorf("expected SPEC/b, got %s", fm.DependsOn[1])
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC04_ExternalInDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - EXTERNAL/proto/api.proto
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "EXTERNAL/proto/api.proto" {
		t.Errorf("expected [EXTERNAL/proto/api.proto], got %v", fm.DependsOn)
	}
}

func TestFrontmatterParse_TC05_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
input: my input text
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Input != "my input text" {
		t.Errorf("expected input 'my input text', got %q", fm.Input)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC06_UnknownFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - SPEC/x
input: text
output: out.go
custom_field: some value
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "SPEC/x" {
		t.Errorf("expected [SPEC/x], got %v", fm.DependsOn)
	}
	if fm.Input != "text" {
		t.Errorf("expected input 'text', got %q", fm.Input)
	}
	if fm.Output != "out.go" {
		t.Errorf("expected output 'out.go', got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC07_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `# Just body content
No frontmatter here.
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC08_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC09_FrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: result.go
---`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Output != "result.go" {
		t.Errorf("expected output 'result.go', got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC10_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", "---   \noutput: result.go\n---\n")

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
	if fm.Output != "" {
		t.Errorf("expected empty output, got %q", fm.Output)
	}
}

func TestFrontmatterParse_TC11_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFrontmatterParse_TC12_PathTraversalError(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got %v", err)
	}
}

func TestFrontmatterParse_TC13_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
: invalid: yaml: content
  - broken
---
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_TC14_UnclosedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
output: result.go
input: some text
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got %v", err)
	}
}

func TestFrontmatterParse_TC15_ExternalFieldIgnored(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - SPEC/x
output: out.go
external: some legacy value
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "SPEC/x" {
		t.Errorf("expected [SPEC/x], got %v", fm.DependsOn)
	}
	if fm.Output != "out.go" {
		t.Errorf("expected output 'out.go', got %q", fm.Output)
	}
	if fm.Input != "" {
		t.Errorf("expected empty input, got %q", fm.Input)
	}
}
