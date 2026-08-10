package mcpwriteverdict_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpcreatetoken"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpwriteverdict"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/subagenttoken"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
)

func TestMCPWriteVerdict(t *testing.T) {
	t.Run("write verdict with pass result", func(t *testing.T) {
		testutils.Chdir(t)

		root := testutils.CreateSpecNode(t, "SPEC/root")
		root.Write()

		b := testutils.CreateSpecNode(t, "SPEC/root/v")
		b.SetType("verdict")
		b.SetOutput("code-from-spec/root/v/verdict.md")
		b.Write()

		token, err := mcpcreatetoken.MCPCreateToken("SPEC/root/v")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		result, err := mcpwriteverdict.MCPWriteVerdict(token, true, "All checks passed.\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "wrote code-from-spec/root/v/verdict.md" {
			t.Errorf("got %q, want %q", result, "wrote code-from-spec/root/v/verdict.md")
		}

		content, err := os.ReadFile("code-from-spec/root/v/verdict.md")
		if err != nil {
			t.Fatalf("reading verdict file: %v", err)
		}
		if string(content) != "All checks passed.\n" {
			t.Errorf("file content: got %q, want %q", string(content), "All checks passed.\n")
		}

		m, err := manifest.OpenManifest(true)
		if err != nil {
			t.Fatalf("opening manifest: %v", err)
		}
		entry, ok := m.Entries["VERDICT/root/v"]
		if !ok {
			t.Fatal("manifest entry VERDICT/root/v not found")
		}
		if entry.Result != "pass" {
			t.Errorf("Result: got %q, want %q", entry.Result, "pass")
		}
		if entry.Checksum == "" {
			t.Error("Checksum is empty")
		}
		if entry.ChainHash == "" {
			t.Error("ChainHash is empty")
		}
	})

	t.Run("write verdict with fail result", func(t *testing.T) {
		testutils.Chdir(t)

		root := testutils.CreateSpecNode(t, "SPEC/root")
		root.Write()

		b := testutils.CreateSpecNode(t, "SPEC/root/v")
		b.SetType("verdict")
		b.SetOutput("code-from-spec/root/v/verdict.md")
		b.Write()

		token, err := mcpcreatetoken.MCPCreateToken("SPEC/root/v")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		result, err := mcpwriteverdict.MCPWriteVerdict(token, false, "Found issues.\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "wrote code-from-spec/root/v/verdict.md" {
			t.Errorf("got %q, want %q", result, "wrote code-from-spec/root/v/verdict.md")
		}

		m, err := manifest.OpenManifest(true)
		if err != nil {
			t.Fatalf("opening manifest: %v", err)
		}
		entry, ok := m.Entries["VERDICT/root/v"]
		if !ok {
			t.Fatal("manifest entry VERDICT/root/v not found")
		}
		if entry.Result != "fail" {
			t.Errorf("Result: got %q, want %q", entry.Result, "fail")
		}
	})

	t.Run("write verdict with default output path", func(t *testing.T) {
		testutils.Chdir(t)

		root := testutils.CreateSpecNode(t, "SPEC/root")
		root.Write()

		b := testutils.CreateSpecNode(t, "SPEC/root/v")
		b.SetType("verdict")
		b.Write()

		token, err := mcpcreatetoken.MCPCreateToken("SPEC/root/v")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		result, err := mcpwriteverdict.MCPWriteVerdict(token, true, "OK.\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "wrote code-from-spec/root/v/verdict.md" {
			t.Errorf("got %q, want %q", result, "wrote code-from-spec/root/v/verdict.md")
		}

		if _, err := os.Stat("code-from-spec/root/v/verdict.md"); err != nil {
			t.Errorf("verdict file does not exist: %v", err)
		}
	})

	t.Run("node has no type - ErrNoOutput", func(t *testing.T) {
		testutils.Chdir(t)

		root := testutils.CreateSpecNode(t, "SPEC/root")
		root.Write()

		b := testutils.CreateSpecNode(t, "SPEC/root/a")
		b.Write()

		token, err := mcpcreatetoken.MCPCreateToken("SPEC/root/a")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		_, err = mcpwriteverdict.MCPWriteVerdict(token, true, "content")
		if !errors.Is(err, mcpwriteverdict.ErrNoOutput) {
			t.Errorf("got %v, want ErrNoOutput", err)
		}
	})

	t.Run("node has type artifact - ErrNotAVerdict", func(t *testing.T) {
		testutils.Chdir(t)

		root := testutils.CreateSpecNode(t, "SPEC/root")
		root.Write()

		b := testutils.CreateSpecNode(t, "SPEC/root/a")
		b.SetType("artifact")
		b.SetOutput("internal/x.go")
		b.Write()

		token, err := mcpcreatetoken.MCPCreateToken("SPEC/root/a")
		if err != nil {
			t.Fatalf("unexpected error creating token: %v", err)
		}

		_, err = mcpwriteverdict.MCPWriteVerdict(token, true, "content")
		if !errors.Is(err, mcpwriteverdict.ErrNotAVerdict) {
			t.Errorf("got %v, want ErrNotAVerdict", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		testutils.Chdir(t)

		_, err := mcpwriteverdict.MCPWriteVerdict("bad-token", true, "content")
		if !errors.Is(err, subagenttoken.ErrInvalidToken) {
			t.Errorf("got %v, want ErrInvalidToken", err)
		}
	})
}
