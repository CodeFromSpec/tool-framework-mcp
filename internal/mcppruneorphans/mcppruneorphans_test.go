package mcppruneorphans_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcppruneorphans"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/oslayer"
	"github.com/CodeFromSpec/tool-framework-mcp/v6/internal/testutils"
)

func TestMCPPruneOrphans(t *testing.T) {
	t.Run("removes orphan entry and deletes artifact file", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/alpha")
		b.SetType("artifact")
		b.SetOutput("out/alpha.go")
		b.SetPublic("## Alpha\ncontent")
		b.Write()

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{Path: "out/alpha.go"}
		m.Entries["ARTIFACT/removed"] = manifest.ManifestEntry{Path: "out/removed.go"}
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		f, err := oslayer.OpenFile("out/removed.go", "overwrite", 0)
		if err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		f.Close()

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		if !strings.Contains(summary, "pruned ARTIFACT/removed (out/removed.go) — deleted from disk") {
			t.Errorf("summary missing pruned line: %q", summary)
		}
		if !strings.Contains(summary, "pruned orphans: 1 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}

		m2, err := manifest.OpenManifest(true)
		if err != nil {
			t.Fatalf("OpenManifest read-only: %v", err)
		}
		if _, ok := m2.Entries["ARTIFACT/alpha"]; !ok {
			t.Error("ARTIFACT/alpha should still exist in manifest")
		}
		if _, ok := m2.Entries["ARTIFACT/removed"]; ok {
			t.Error("ARTIFACT/removed should not exist in manifest")
		}

		_, err = oslayer.OpenFile("out/removed.go", "read", 0)
		if !errors.Is(err, oslayer.ErrFileUnreadable) {
			t.Errorf("expected ErrFileUnreadable, got %v", err)
		}
	})

	t.Run("removes orphan when artifact file does not exist on disk", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/alpha")
		b.SetType("artifact")
		b.SetOutput("out/alpha.go")
		b.Write()

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{Path: "out/alpha.go"}
		m.Entries["ARTIFACT/gone"] = manifest.ManifestEntry{Path: "out/gone.go"}
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		if !strings.Contains(summary, "pruned ARTIFACT/gone (out/gone.go) — not found on disk") {
			t.Errorf("summary missing pruned line: %q", summary)
		}
		if !strings.Contains(summary, "pruned orphans: 1 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}

		m2, err := manifest.OpenManifest(true)
		if err != nil {
			t.Fatalf("OpenManifest read-only: %v", err)
		}
		if _, ok := m2.Entries["ARTIFACT/gone"]; ok {
			t.Error("ARTIFACT/gone should not exist in manifest")
		}
	})

	t.Run("no orphans zero entries removed", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/alpha")
		b.SetType("artifact")
		b.SetOutput("out/alpha.go")
		b.Write()

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{Path: "out/alpha.go"}
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		if !strings.Contains(summary, "pruned orphans: 0 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}
	})

	t.Run("orphan because node has no type", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/alpha")
		b.SetType("artifact")
		b.SetOutput("out/alpha.go")
		b.Write()

		b2 := testutils.CreateSpecNode(t, "SPEC/docs-only")
		b2.Write()

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{Path: "out/alpha.go"}
		m.Entries["ARTIFACT/docs-only"] = manifest.ManifestEntry{Path: "out/docs.go"}
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		f, err := oslayer.OpenFile("out/docs.go", "overwrite", 0)
		if err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		f.Close()

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		if !strings.Contains(summary, "pruned ARTIFACT/docs-only (out/docs.go) — deleted from disk") {
			t.Errorf("summary missing pruned line: %q", summary)
		}
		if !strings.Contains(summary, "pruned orphans: 1 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}

		_, err = oslayer.OpenFile("out/docs.go", "read", 0)
		if !errors.Is(err, oslayer.ErrFileUnreadable) {
			t.Errorf("expected ErrFileUnreadable, got %v", err)
		}
	})

	t.Run("empty manifest no errors", func(t *testing.T) {
		testutils.Chdir(t)

		b := testutils.CreateSpecNode(t, "SPEC/alpha")
		b.SetType("artifact")
		b.SetOutput("out/alpha.go")
		b.Write()

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		if !strings.Contains(summary, "pruned orphans: 0 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}
	})

	t.Run("multiple orphans pruned in alphabetical order", func(t *testing.T) {
		testutils.Chdir(t)

		if err := os.MkdirAll("code-from-spec", 0755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}

		m, err := manifest.OpenManifest(false)
		if err != nil {
			t.Fatalf("OpenManifest: %v", err)
		}
		defer func() { _ = m.Discard() }()
		m.Entries["ARTIFACT/zebra"] = manifest.ManifestEntry{Path: "out/z.go"}
		m.Entries["ARTIFACT/apple"] = manifest.ManifestEntry{Path: "out/a.go"}
		if err := m.Save(); err != nil {
			t.Fatalf("Save: %v", err)
		}

		fa, err := oslayer.OpenFile("out/a.go", "overwrite", 0)
		if err != nil {
			t.Fatalf("OpenFile a.go: %v", err)
		}
		fa.Close()

		fz, err := oslayer.OpenFile("out/z.go", "overwrite", 0)
		if err != nil {
			t.Fatalf("OpenFile z.go: %v", err)
		}
		fz.Close()

		summary, err := mcppruneorphans.MCPPruneOrphans()
		if err != nil {
			t.Fatalf("MCPPruneOrphans: %v", err)
		}

		appleIdx := strings.Index(summary, "ARTIFACT/apple")
		zebraIdx := strings.Index(summary, "ARTIFACT/zebra")
		if appleIdx == -1 {
			t.Error("summary missing ARTIFACT/apple")
		}
		if zebraIdx == -1 {
			t.Error("summary missing ARTIFACT/zebra")
		}
		if appleIdx != -1 && zebraIdx != -1 && appleIdx > zebraIdx {
			t.Error("ARTIFACT/apple should appear before ARTIFACT/zebra in summary")
		}
		if !strings.Contains(summary, "pruned orphans: 2 entries removed") {
			t.Errorf("summary missing count line: %q", summary)
		}
	})
}
