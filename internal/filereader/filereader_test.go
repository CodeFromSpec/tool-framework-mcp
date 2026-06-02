// code-from-spec: ROOT/golang/tests/os/file_reader@a2Khx6txHzmfV_4PJF-vSkgRUVI
package filereader_test

import (
	"errors"
	"os"
	"testing"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/filereader"
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

func TestFileReader_OpensAndReadsAllLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta\ngamma\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta", "gamma"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_NormalizesCRLF(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\r\nbeta\r\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\nbeta"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_SkipLinesAdvancesReader(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\nthree\nfour\nfive\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	filereader.FileSkipLines(r, 2)

	got, err := filereader.FileReadLine(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "three" {
		t.Errorf("got %q, want %q", got, "three")
	}
}

func TestFileReader_SkipLinesPastEndOfFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	filereader.FileSkipLines(r, 10)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_PreservesLeadingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("  alpha\n    beta\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"  alpha", "    beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesTrailingWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha  \nbeta   \n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha  ", "beta   "} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesInternalWhitespace(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha   beta\none\ttwo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha   beta", "one\ttwo"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesEmptyLines(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n\n\nbeta\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"alpha", "", "", "beta"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_PreservesNonASCII(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("café\n日本語\n🎉🚀\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	for _, want := range []string{"café", "日本語", "🎉🚀"} {
		got, err := filereader.FileReadLine(r)
		if err != nil {
			t.Fatalf("FileReadLine: %v", err)
		}
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFileReader_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_SingleLineWithoutNewline(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	defer filereader.FileClose(r)

	got, err := filereader.FileReadLine(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	_, err := filereader.FileOpen(&pathutils.PathCfs{Value: "nonexistent/file.txt"})
	if !errors.Is(err, filereader.ErrFileUnreadable) {
		t.Errorf("expected ErrFileUnreadable, got %v", err)
	}
}

func TestFileReader_ReadAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}

	filereader.FileClose(r)

	_, err = filereader.FileReadLine(r)
	if !errors.Is(err, filereader.ErrEndOfFile) {
		t.Errorf("expected ErrEndOfFile, got %v", err)
	}
}

func TestFileReader_SkipAfterClose(t *testing.T) {
	dir := t.TempDir()
	testChdir(t, dir)

	if err := os.WriteFile("file.txt", []byte("alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}

	r, err := filereader.FileOpen(&pathutils.PathCfs{Value: "file.txt"})
	if err != nil {
		t.Fatal(err)
	}

	filereader.FileClose(r)

	filereader.FileSkipLines(r, 1)
}
