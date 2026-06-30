package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type NodeBuilder struct {
	t           *testing.T
	logicalName string
	output      *string
	input       *string
	dependsOn   []string
	public      *string
	agent       *string
	private     *string
}

func CreateSpecNode(t *testing.T, logicalName string) *NodeBuilder {
	t.Helper()
	return &NodeBuilder{t: t, logicalName: logicalName}
}

func (b *NodeBuilder) SetOutput(value string)    { b.output = &value }
func (b *NodeBuilder) SetInput(value string)     { b.input = &value }
func (b *NodeBuilder) AddDependsOn(value string) { b.dependsOn = append(b.dependsOn, value) }
func (b *NodeBuilder) SetPublic(content string)  { b.public = &content }
func (b *NodeBuilder) SetAgent(content string)   { b.agent = &content }
func (b *NodeBuilder) SetPrivate(content string) { b.private = &content }

func (b *NodeBuilder) Write() {
	b.t.Helper()
	var buf strings.Builder

	if b.output != nil || b.input != nil || len(b.dependsOn) > 0 {
		buf.WriteString("---\n")
		if len(b.dependsOn) > 0 {
			buf.WriteString("depends_on:\n")
			for _, dep := range b.dependsOn {
				buf.WriteString("  - " + dep + "\n")
			}
		}
		if b.input != nil {
			buf.WriteString("input: " + *b.input + "\n")
		}
		if b.output != nil {
			buf.WriteString("output: " + *b.output + "\n")
		}
		buf.WriteString("---\n")
	}

	buf.WriteString("# " + b.logicalName + "\n")

	if b.public != nil {
		buf.WriteString("\n# Public\n")
		buf.WriteString(*b.public + "\n")
	}
	if b.agent != nil {
		buf.WriteString("\n# Agent\n")
		buf.WriteString(*b.agent + "\n")
	}
	if b.private != nil {
		buf.WriteString("\n# Private\n")
		buf.WriteString(*b.private + "\n")
	}

	relative := strings.TrimPrefix(b.logicalName, "SPEC/")
	path := filepath.Join("code-from-spec", filepath.FromSlash(relative), "_node.md")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		b.t.Fatalf("CreateSpecNode.Write: mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0644); err != nil {
		b.t.Fatalf("CreateSpecNode.Write: write %s: %v", path, err)
	}
}

func WriteRawNode(t *testing.T, logicalName string, content string) {
	t.Helper()
	relative := strings.TrimPrefix(logicalName, "SPEC/")
	path := filepath.Join("code-from-spec", filepath.FromSlash(relative), "_node.md")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("WriteRawNode: mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteRawNode: write %s: %v", path, err)
	}
}
