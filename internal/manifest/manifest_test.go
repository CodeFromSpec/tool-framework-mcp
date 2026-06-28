// code-from-spec: SPEC/golang/tests/manifest@b3syEWBH4JnIOsSsazNHEO4KV9M
package manifest_test

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/manifest"
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

func testWriteManifest(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll("code-from-spec", 0o755); err != nil {
		t.Fatalf("testWriteManifest mkdir: %v", err)
	}
	if err := os.WriteFile("code-from-spec/.manifest", []byte(content), 0o644); err != nil {
		t.Fatalf("testWriteManifest write: %v", err)
	}
}

func testReadManifest(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("code-from-spec/.manifest")
	if err != nil {
		t.Fatalf("testReadManifest: %v", err)
	}
	return string(data)
}

func TestManifestOpen_Read_ExistingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:abc123;chain:hash001\n" +
		"ARTIFACT/beta;path:generated/beta.go;checksum:def456;chain:hash002\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("read")
	if err != nil {
		t.Fatalf("ManifestOpen read: %v", err)
	}

	if handle.Mode != "read" {
		t.Errorf("Mode = %q, want %q", handle.Mode, "read")
	}
	if handle.Version != "v5" {
		t.Errorf("Version = %q, want %q", handle.Version, "v5")
	}
	if len(handle.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(handle.Entries))
	}

	alpha, ok := handle.Entries["ARTIFACT/alpha"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/alpha")
	}
	if alpha.Path != "generated/alpha.go" {
		t.Errorf("alpha.Path = %q, want %q", alpha.Path, "generated/alpha.go")
	}
	if alpha.Checksum != "abc123" {
		t.Errorf("alpha.Checksum = %q, want %q", alpha.Checksum, "abc123")
	}
	if alpha.ChainHash != "hash001" {
		t.Errorf("alpha.ChainHash = %q, want %q", alpha.ChainHash, "hash001")
	}

	beta, ok := handle.Entries["ARTIFACT/beta"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/beta")
	}
	if beta.Path != "generated/beta.go" {
		t.Errorf("beta.Path = %q, want %q", beta.Path, "generated/beta.go")
	}
	if beta.Checksum != "def456" {
		t.Errorf("beta.Checksum = %q, want %q", beta.Checksum, "def456")
	}
	if beta.ChainHash != "hash002" {
		t.Errorf("beta.ChainHash = %q, want %q", beta.ChainHash, "hash002")
	}
}

func TestManifestOpen_Read_EmptyManifest(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	testWriteManifest(t, "code-from-spec: v5\n")

	handle, err := manifest.ManifestOpen("read")
	if err != nil {
		t.Fatalf("ManifestOpen read: %v", err)
	}

	if len(handle.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(handle.Entries))
	}
}

func TestManifestOpen_Read_MissingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("read")
	if err != nil {
		t.Fatalf("ManifestOpen read: %v", err)
	}

	if len(handle.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(handle.Entries))
	}

	if _, err := os.Stat("code-from-spec/.manifest"); !os.IsNotExist(err) {
		t.Error("manifest file should not exist after read of missing file")
	}
}

func TestManifestOpen_Write_LoadsExistingEntries(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:abc123;chain:hash001\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if len(handle.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(handle.Entries))
	}
	if _, ok := handle.Entries["ARTIFACT/alpha"]; !ok {
		t.Error("missing entry ARTIFACT/alpha")
	}

	if err := manifest.ManifestDiscard(handle); err != nil {
		t.Errorf("ManifestDiscard: %v", err)
	}
}

func TestManifestOpen_Write_MissingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if len(handle.Entries) != 0 {
		t.Errorf("len(Entries) = %d, want 0", len(handle.Entries))
	}

	if err := manifest.ManifestDiscard(handle); err != nil {
		t.Errorf("ManifestDiscard: %v", err)
	}
}

func TestManifestSave_CreatesFromScratch(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	handle.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "generated/beta.go",
		Checksum:  "csumbeta",
		ChainHash: "hashbeta",
	}
	handle.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{
		Path:      "generated/alpha.go",
		Checksum:  "csumalpha",
		ChainHash: "hashalpha",
	}

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave: %v", err)
	}

	got := testReadManifest(t)

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 1 || !strings.HasPrefix(lines[0], "code-from-spec:") {
		t.Errorf("missing header line, got: %q", got)
	}

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 entries), got %d: %q", len(lines), got)
	}

	if !strings.Contains(lines[1], "ARTIFACT/alpha") {
		t.Errorf("line 1 should be ARTIFACT/alpha, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "ARTIFACT/beta") {
		t.Errorf("line 2 should be ARTIFACT/beta, got %q", lines[2])
	}
}

func TestManifestSave_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:csumalpha;chain:hashalpha\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	handle.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "generated/beta.go",
		Checksum:  "csumbeta",
		ChainHash: "hashbeta",
	}

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave: %v", err)
	}

	got := testReadManifest(t)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 entries), got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[1], "ARTIFACT/alpha") {
		t.Errorf("line 1 should be ARTIFACT/alpha, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "ARTIFACT/beta") {
		t.Errorf("line 2 should be ARTIFACT/beta, got %q", lines[2])
	}
}

func TestManifestSave_ModifiedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:old-checksum;chain:hashalpha\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	entry := handle.Entries["ARTIFACT/alpha"]
	entry.Checksum = "new-checksum"
	handle.Entries["ARTIFACT/alpha"] = entry

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave: %v", err)
	}

	got := testReadManifest(t)
	if !strings.Contains(got, "new-checksum") {
		t.Errorf("expected new-checksum in manifest, got: %q", got)
	}
	if strings.Contains(got, "old-checksum") {
		t.Errorf("old-checksum should not appear in manifest, got: %q", got)
	}
}

func TestManifestSave_RemovedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:csumalpha;chain:hashalpha\n" +
		"ARTIFACT/beta;path:generated/beta.go;checksum:csumbeta;chain:hashbeta\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	delete(handle.Entries, "ARTIFACT/beta")

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave: %v", err)
	}

	got := testReadManifest(t)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 entry), got %d: %q", len(lines), got)
	}
	if !strings.Contains(lines[1], "ARTIFACT/alpha") {
		t.Errorf("line 1 should be ARTIFACT/alpha, got %q", lines[1])
	}
	if strings.Contains(got, "ARTIFACT/beta") {
		t.Errorf("ARTIFACT/beta should not appear in manifest, got: %q", got)
	}
}

func TestManifestDiscard_DoesNotModifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	content := "code-from-spec: v5\n" +
		"ARTIFACT/alpha;path:generated/alpha.go;checksum:csumalpha;chain:hashalpha\n"
	testWriteManifest(t, content)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	handle.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "generated/beta.go",
		Checksum:  "csumbeta",
		ChainHash: "hashbeta",
	}

	if err := manifest.ManifestDiscard(handle); err != nil {
		t.Fatalf("ManifestDiscard: %v", err)
	}

	got := testReadManifest(t)
	if strings.Contains(got, "ARTIFACT/beta") {
		t.Errorf("ARTIFACT/beta should not appear after discard, got: %q", got)
	}
	if !strings.Contains(got, "ARTIFACT/alpha") {
		t.Errorf("ARTIFACT/alpha should still be present, got: %q", got)
	}
}

func TestManifestSave_WrongMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("read")
	if err != nil {
		t.Fatalf("ManifestOpen read: %v", err)
	}

	err = manifest.ManifestSave(handle)
	if !errors.Is(err, manifest.ErrWrongMode) {
		t.Errorf("ManifestSave on read handle: got %v, want ErrWrongMode", err)
	}
}

func TestManifestDiscard_WrongMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("read")
	if err != nil {
		t.Fatalf("ManifestOpen read: %v", err)
	}

	err = manifest.ManifestDiscard(handle)
	if !errors.Is(err, manifest.ErrWrongMode) {
		t.Errorf("ManifestDiscard on read handle: got %v, want ErrWrongMode", err)
	}
}

func TestManifestDiscard_AfterSave(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave: %v", err)
	}

	err = manifest.ManifestDiscard(handle)
	if !errors.Is(err, manifest.ErrHandleClosed) {
		t.Errorf("ManifestDiscard after save: got %v, want ErrHandleClosed", err)
	}
}

func TestManifestSave_AfterDiscard(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if err := manifest.ManifestDiscard(handle); err != nil {
		t.Fatalf("ManifestDiscard: %v", err)
	}

	err = manifest.ManifestSave(handle)
	if !errors.Is(err, manifest.ErrHandleClosed) {
		t.Errorf("ManifestSave after discard: got %v, want ErrHandleClosed", err)
	}
}

func TestManifestSave_AfterSave(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if err := manifest.ManifestSave(handle); err != nil {
		t.Fatalf("ManifestSave first: %v", err)
	}

	err = manifest.ManifestSave(handle)
	if !errors.Is(err, manifest.ErrHandleClosed) {
		t.Errorf("second ManifestSave: got %v, want ErrHandleClosed", err)
	}
}

func TestManifestDiscard_AfterDiscard(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	handle, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	if err := manifest.ManifestDiscard(handle); err != nil {
		t.Fatalf("ManifestDiscard first: %v", err)
	}

	err = manifest.ManifestDiscard(handle)
	if !errors.Is(err, manifest.ErrHandleClosed) {
		t.Errorf("second ManifestDiscard: got %v, want ErrHandleClosed", err)
	}
}

func TestManifestOpen_InvalidMode(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	_, err := manifest.ManifestOpen("invalid")
	if !errors.Is(err, manifest.ErrInvalidMode) {
		t.Errorf("ManifestOpen invalid mode: got %v, want ErrInvalidMode", err)
	}
}

func TestManifestOpen_ConcurrentReaders(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle, err := manifest.ManifestOpen("read")
			if err != nil {
				errs <- err
				return
			}
			_ = handle
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent readers timed out — readers may be blocking each other")
	}

	close(errs)
	for err := range errs {
		t.Errorf("concurrent reader error: %v", err)
	}
}

func TestManifestOpen_WriterBlocksReader(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	writer, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen write: %v", err)
	}

	readerReady := make(chan struct{})
	readerDone := make(chan error, 1)

	go func() {
		close(readerReady)
		handle, err := manifest.ManifestOpen("read")
		if err != nil {
			readerDone <- err
			return
		}
		_ = handle
		readerDone <- nil
	}()

	<-readerReady
	time.Sleep(100 * time.Millisecond)

	if err := manifest.ManifestDiscard(writer); err != nil {
		t.Fatalf("ManifestDiscard: %v", err)
	}

	select {
	case err := <-readerDone:
		if err != nil {
			t.Errorf("reader after writer released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reader did not unblock after writer released lock")
	}
}

func TestManifestOpen_WriterBlocksWriter(t *testing.T) {
	tmpDir := t.TempDir()
	testChdir(t, tmpDir)

	first, err := manifest.ManifestOpen("write")
	if err != nil {
		t.Fatalf("ManifestOpen first write: %v", err)
	}

	secondReady := make(chan struct{})
	secondDone := make(chan error, 1)

	go func() {
		close(secondReady)
		handle, err := manifest.ManifestOpen("write")
		if err != nil {
			secondDone <- err
			return
		}
		_ = manifest.ManifestDiscard(handle)
		secondDone <- nil
	}()

	<-secondReady
	time.Sleep(100 * time.Millisecond)

	if err := manifest.ManifestDiscard(first); err != nil {
		t.Fatalf("ManifestDiscard first: %v", err)
	}

	select {
	case err := <-secondDone:
		if err != nil {
			t.Errorf("second writer after first released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer did not unblock after first writer released lock")
	}
}
