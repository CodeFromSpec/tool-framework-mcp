// code-from-spec: ROOT/golang/tests/parsing/frontmatter@3LX7qeyF6d-_s9zj1DMwVakf2mE
package frontmatter_test

import (
	"errors"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatalf("testWriteFile mkdir: %v", err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatalf("testWriteFile: %v", err)
	}
}

func TestFrontmatterParse_CompleteFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - dep/one
  - dep/two
external:
  - path: ext/alpha.md
  - path: ext/beta.md
input: some-artifact
outputs:
  - id: out1
    path: gen/out1.go
  - id: out2
    path: gen/out2.go
---
body content
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "dep/one" || fm.DependsOn[1] != "dep/two" {
		t.Errorf("DependsOn = %v, want [dep/one dep/two]", fm.DependsOn)
	}
	if len(fm.External) != 2 || fm.External[0].Path != "ext/alpha.md" || fm.External[1].Path != "ext/beta.md" {
		t.Errorf("External = %v, want [{ext/alpha.md} {ext/beta.md}]", fm.External)
	}
	if fm.Input != "some-artifact" {
		t.Errorf("Input = %q, want %q", fm.Input, "some-artifact")
	}
	if len(fm.Outputs) != 2 {
		t.Fatalf("Outputs len = %d, want 2", len(fm.Outputs))
	}
	if fm.Outputs[0].ID != "out1" || fm.Outputs[0].Path != "gen/out1.go" {
		t.Errorf("Outputs[0] = %+v, want {out1 gen/out1.go}", fm.Outputs[0])
	}
	if fm.Outputs[1].ID != "out2" || fm.Outputs[1].Path != "gen/out2.go" {
		t.Errorf("Outputs[1] = %+v, want {out2 gen/out2.go}", fm.Outputs[1])
	}
}

func TestFrontmatterParse_OnlyOutputs(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
outputs:
  - id: myid
    path: my/path.go
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 1 || fm.Outputs[0].ID != "myid" || fm.Outputs[0].Path != "my/path.go" {
		t.Errorf("Outputs = %v, want [{myid my/path.go}]", fm.Outputs)
	}
}

func TestFrontmatterParse_OnlyDependsOn(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - first/dep
  - second/dep
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 2 || fm.DependsOn[0] != "first/dep" || fm.DependsOn[1] != "second/dep" {
		t.Errorf("DependsOn = %v, want [first/dep second/dep]", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_OnlyExternal(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
external:
  - path: docs/one.md
  - path: docs/two.md
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.External) != 2 || fm.External[0].Path != "docs/one.md" || fm.External[1].Path != "docs/two.md" {
		t.Errorf("External = %v, want [{docs/one.md} {docs/two.md}]", fm.External)
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_OnlyInput(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
input: my-input-artifact
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Input != "my-input-artifact" {
		t.Errorf("Input = %q, want %q", fm.Input, "my-input-artifact")
	}
	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_IgnoresUnknownFields(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - known/dep
outputs:
  - id: outid
    path: out/path.go
custom_field: ignored_value
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "known/dep" {
		t.Errorf("DependsOn = %v, want [known/dep]", fm.DependsOn)
	}
	if len(fm.Outputs) != 1 || fm.Outputs[0].ID != "outid" || fm.Outputs[0].Path != "out/path.go" {
		t.Errorf("Outputs = %v, want [{outid out/path.go}]", fm.Outputs)
	}
}

func TestFrontmatterParse_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `just body content
no delimiters here
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_EmptyFrontmatter(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
---
body here
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_OnlyFrontmatterNoBody(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - sole/dep
---
`)

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 1 || fm.DependsOn[0] != "sole/dep" {
		t.Errorf("DependsOn = %v, want [sole/dep]", fm.DependsOn)
	}
}

func TestFrontmatterParse_DelimiterWithTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", "---   \ndepends_on:\n  - something\n---\n")

	fm, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fm.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty (delimiter not recognized)", fm.DependsOn)
	}
	if len(fm.External) != 0 {
		t.Errorf("External = %v, want empty", fm.External)
	}
	if fm.Input != "" {
		t.Errorf("Input = %q, want empty", fm.Input)
	}
	if len(fm.Outputs) != 0 {
		t.Errorf("Outputs = %v, want empty", fm.Outputs)
	}
}

func TestFrontmatterParse_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "nonexistent/file.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got: %v", err)
	}
}

func TestFrontmatterParse_PropagatesPathErrors(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "../../outside"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, pathutils.ErrDirectoryTraversal) {
		t.Errorf("expected ErrDirectoryTraversal, got: %v", err)
	}
}

func TestFrontmatterParse_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on: [unclosed bracket
---
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_UnclosedFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
depends_on:
  - some/dep
just body with no closing delimiter
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_ExternalMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
external:
  - other_field: value
---
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}

func TestFrontmatterParse_OutputMissingPath(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	testWriteFile(t, "node.md", `---
outputs:
  - id: myid
---
`)

	_, err := frontmatter.FrontmatterParse(&pathutils.PathCfs{Value: "node.md"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, frontmatter.ErrMalformedYAML) {
		t.Errorf("expected ErrMalformedYAML, got: %v", err)
	}
}
