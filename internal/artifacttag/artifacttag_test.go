// code-from-spec: ROOT/golang/internal/artifact_tag/tests@bT-9KA8s65UounzYIy5d380lB6I
package artifacttag_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
)

// writeTemp creates a temporary file with the given content and returns its path.
// The file is automatically removed when the test ends.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile.go")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// ExtractArtifactTag — happy path
// ---------------------------------------------------------------------------

func TestExtractArtifactTag_SimpleTag(t *testing.T) {
	content := "// code-from-spec: some/logical/name@AAAAAAAAAAAAAAAAAAAAAAAAA00\npackage main\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "some/logical/name" {
		t.Errorf("LogicalName = %q, want %q", got.LogicalName, "some/logical/name")
	}
	if got.Hash != "AAAAAAAAAAAAAAAAAAAAAAAAA00" {
		t.Errorf("Hash = %q, want %q", got.Hash, "AAAAAAAAAAAAAAAAAAAAAAAAA00")
	}
}

func TestExtractArtifactTag_TagNotOnFirstLine(t *testing.T) {
	// Tag appears after several lines of other content.
	content := "package main\n\nimport \"fmt\"\n\n// code-from-spec: my/node@bT-9KA8s65UounzYIy5d380lB6I\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "my/node" {
		t.Errorf("LogicalName = %q, want %q", got.LogicalName, "my/node")
	}
	if got.Hash != "bT-9KA8s65UounzYIy5d380lB6I" {
		t.Errorf("Hash = %q, want %q", got.Hash, "bT-9KA8s65UounzYIy5d380lB6I")
	}
}

func TestExtractArtifactTag_HashExactly27Chars(t *testing.T) {
	// Hash is exactly 27 characters — the minimum valid length.
	hash27 := "123456789012345678901234567"
	if len(hash27) != 27 {
		t.Fatalf("test setup error: hash27 length is %d, want 27", len(hash27))
	}
	content := "// code-from-spec: node@" + hash27 + "\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Hash != hash27 {
		t.Errorf("Hash = %q, want %q", got.Hash, hash27)
	}
}

func TestExtractArtifactTag_HashLongerThan27CharsIsTruncated(t *testing.T) {
	// Extra characters after the 27-char hash are ignored.
	longHash := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1_extra"
	content := "// code-from-spec: node@" + longHash + "\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := longHash[:27]
	if got.Hash != want {
		t.Errorf("Hash = %q, want %q", got.Hash, want)
	}
}

func TestExtractArtifactTag_FirstAtUsedForSplit(t *testing.T) {
	// LogicalName itself contains '@' — the FIRST '@' must be used as the
	// split point, so the logical name is the part before it.
	content := "// code-from-spec: org@scope/node@AAAAAAAAAAAAAAAAAAAAAAAAA00\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "org" {
		t.Errorf("LogicalName = %q, want %q", got.LogicalName, "org")
	}
}

func TestExtractArtifactTag_TagEmbeddedInHashComment(t *testing.T) {
	// The prefix may appear anywhere on the line (e.g. after `#`).
	content := "# code-from-spec: root/node@bT-9KA8s65UounzYIy5d380lB6I\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "root/node" {
		t.Errorf("LogicalName = %q, want %q", got.LogicalName, "root/node")
	}
}

func TestExtractArtifactTag_OnlyFirstTagReturned(t *testing.T) {
	// Two valid tag lines — only the first must be returned.
	content := "// code-from-spec: first/node@AAAAAAAAAAAAAAAAAAAAAAAAA00\n" +
		"// code-from-spec: second/node@BBBBBBBBBBBBBBBBBBBBBBBBBBB\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "first/node" {
		t.Errorf("LogicalName = %q, want %q", got.LogicalName, "first/node")
	}
}

// ---------------------------------------------------------------------------
// ExtractArtifactTag — ErrFileUnreadable
// ---------------------------------------------------------------------------

func TestExtractArtifactTag_ErrFileUnreadable_NonExistentFile(t *testing.T) {
	_, err := artifacttag.ExtractArtifactTag("/nonexistent/path/to/file.go")
	if !errors.Is(err, artifacttag.ErrFileUnreadable) {
		t.Errorf("err = %v, want errors.Is(err, ErrFileUnreadable)", err)
	}
}

// ---------------------------------------------------------------------------
// ExtractArtifactTag — ErrNoTagFound
// ---------------------------------------------------------------------------

func TestExtractArtifactTag_ErrNoTagFound_EmptyFile(t *testing.T) {
	path := writeTemp(t, "")

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want errors.Is(err, ErrNoTagFound)", err)
	}
}

func TestExtractArtifactTag_ErrNoTagFound_NoTagInFile(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"
	path := writeTemp(t, content)

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want errors.Is(err, ErrNoTagFound)", err)
	}
}

func TestExtractArtifactTag_ErrNoTagFound_PrefixAbsentButSimilarText(t *testing.T) {
	// A line with "code-from-spec" but without the trailing ": " must not match.
	content := "// code-from-spec somenode@AAAAAAAAAAAAAAAAAAAAAAAAA00\n"
	path := writeTemp(t, content)

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrNoTagFound) {
		t.Errorf("err = %v, want errors.Is(err, ErrNoTagFound)", err)
	}
}

// ---------------------------------------------------------------------------
// ExtractArtifactTag — ErrMalformedTag
// ---------------------------------------------------------------------------

func TestExtractArtifactTag_ErrMalformedTag_MissingAt(t *testing.T) {
	// Tag prefix is present but no '@' separator.
	content := "// code-from-spec: nodewithnoathash\n"
	path := writeTemp(t, content)

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want errors.Is(err, ErrMalformedTag)", err)
	}
}

func TestExtractArtifactTag_ErrMalformedTag_HashTooShort(t *testing.T) {
	// Hash is only 26 characters — one short of the required 27.
	hash26 := "AAAAAAAAAAAAAAAAAAAAAAAAAA"
	if len(hash26) != 26 {
		t.Fatalf("test setup error: hash26 length is %d, want 26", len(hash26))
	}
	content := "// code-from-spec: node@" + hash26 + "\n"
	path := writeTemp(t, content)

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want errors.Is(err, ErrMalformedTag)", err)
	}
}

func TestExtractArtifactTag_ErrMalformedTag_EmptyHashAfterAt(t *testing.T) {
	// '@' is present but nothing follows it.
	content := "// code-from-spec: node@\n"
	path := writeTemp(t, content)

	_, err := artifacttag.ExtractArtifactTag(path)
	if !errors.Is(err, artifacttag.ErrMalformedTag) {
		t.Errorf("err = %v, want errors.Is(err, ErrMalformedTag)", err)
	}
}

func TestExtractArtifactTag_ErrMalformedTag_EmptyLogicalName(t *testing.T) {
	// '@' is right after the prefix — empty logical name, but hash is valid.
	// The implementation does not validate that LogicalName is non-empty, so
	// this should succeed and return an empty LogicalName.
	hash27 := "AAAAAAAAAAAAAAAAAAAAAAAAA00"
	content := "// code-from-spec: @" + hash27 + "\n"
	path := writeTemp(t, content)

	got, err := artifacttag.ExtractArtifactTag(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LogicalName != "" {
		t.Errorf("LogicalName = %q, want empty string", got.LogicalName)
	}
	if got.Hash != hash27 {
		t.Errorf("Hash = %q, want %q", got.Hash, hash27)
	}
}

// ---------------------------------------------------------------------------
// Sentinel identity — errors.Is must work transitively
// ---------------------------------------------------------------------------

func TestErrorSentinels_AreDistinct(t *testing.T) {
	if errors.Is(artifacttag.ErrFileUnreadable, artifacttag.ErrNoTagFound) {
		t.Error("ErrFileUnreadable should not match ErrNoTagFound")
	}
	if errors.Is(artifacttag.ErrFileUnreadable, artifacttag.ErrMalformedTag) {
		t.Error("ErrFileUnreadable should not match ErrMalformedTag")
	}
	if errors.Is(artifacttag.ErrNoTagFound, artifacttag.ErrMalformedTag) {
		t.Error("ErrNoTagFound should not match ErrMalformedTag")
	}
}
