package manifest_test

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/manifest"
	"github.com/CodeFromSpec/tool-framework-mcp/v5/internal/testutils"
)

func writeManifestFile(t *testing.T, lines []string) {
	t.Helper()
	if err := os.MkdirAll("code-from-spec", 0755); err != nil {
		t.Fatalf("failed to create code-from-spec dir: %v", err)
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile("code-from-spec/.manifest", []byte(content), 0644); err != nil {
		t.Fatalf("failed to write manifest file: %v", err)
	}
}

func readManifestFile(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("code-from-spec/.manifest")
	if err != nil {
		t.Fatalf("failed to read manifest file: %v", err)
	}
	raw := string(data)
	var lines []string
	start := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] == '\n' {
			lines = append(lines, raw[start:i])
			start = i + 1
		}
	}
	if start < len(raw) {
		lines = append(lines, raw[start:])
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func TestOpenManifest_ReadOnly_ExistingManifest(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/foo/bar;path:internal/foo/bar.go;checksum:aaaaaaaaaaaaaaaaaaaaaaaaaaaa1;chain:aaaaaaaaaaaaaaaaaaaaaaaaaaaa2",
		"ARTIFACT/foo/baz;path:internal/foo/baz.go;checksum:bbbbbbbbbbbbbbbbbbbbbbbbbbbb1;chain:bbbbbbbbbbbbbbbbbbbbbbbbbbbb2",
	})

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Version != "v5" {
		t.Errorf("expected Version v5, got %q", m.Version)
	}
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}

	e1, ok := m.Entries["ARTIFACT/foo/bar"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/foo/bar")
	}
	if e1.Path != "internal/foo/bar.go" {
		t.Errorf("unexpected path: %q", e1.Path)
	}
	if e1.Checksum != "aaaaaaaaaaaaaaaaaaaaaaaaaaaa1" {
		t.Errorf("unexpected checksum: %q", e1.Checksum)
	}
	if e1.ChainHash != "aaaaaaaaaaaaaaaaaaaaaaaaaaaa2" {
		t.Errorf("unexpected chain hash: %q", e1.ChainHash)
	}

	e2, ok := m.Entries["ARTIFACT/foo/baz"]
	if !ok {
		t.Fatal("missing entry ARTIFACT/foo/baz")
	}
	if e2.Path != "internal/foo/baz.go" {
		t.Errorf("unexpected path: %q", e2.Path)
	}
	if e2.Checksum != "bbbbbbbbbbbbbbbbbbbbbbbbbbbb1" {
		t.Errorf("unexpected checksum: %q", e2.Checksum)
	}
	if e2.ChainHash != "bbbbbbbbbbbbbbbbbbbbbbbbbbbb2" {
		t.Errorf("unexpected chain hash: %q", e2.ChainHash)
	}
}

func TestOpenManifest_ReadOnly_HeaderOnly(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
	})

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(m.Entries))
	}
}

func TestOpenManifest_ReadOnly_MissingManifest(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(m.Entries))
	}

	if _, statErr := os.Stat("code-from-spec/.manifest"); !os.IsNotExist(statErr) {
		t.Error("manifest file should not have been created")
	}
}

func TestOpenManifest_Writable_LoadsExistingEntries(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:cccccccccccccccccccccccccc1;chain:cccccccccccccccccccccccccc2",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	if _, ok := m.Entries["ARTIFACT/alpha"]; !ok {
		t.Error("missing entry ARTIFACT/alpha")
	}

	if err := m.Discard(); err != nil {
		t.Errorf("Discard returned error: %v", err)
	}
}

func TestOpenManifest_Writable_MissingManifest(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(m.Entries))
	}

	if err := m.Discard(); err != nil {
		t.Errorf("Discard returned error: %v", err)
	}
}

func TestSave_CreatesManifestFromScratch(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = m.Discard() }()

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta.go",
		Checksum:  "betaChecksum111111111111111",
		ChainHash: "betaChain1111111111111111111",
	}
	m.Entries["ARTIFACT/alpha"] = manifest.ManifestEntry{
		Path:      "internal/alpha.go",
		Checksum:  "alphaChecksum11111111111111",
		ChainHash: "alphaChain111111111111111111",
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "code-from-spec: v5" {
		t.Errorf("unexpected header: %q", lines[0])
	}
	expected1 := "ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111"
	expected2 := "ARTIFACT/beta;path:internal/beta.go;checksum:betaChecksum111111111111111;chain:betaChain1111111111111111111"
	if lines[1] != expected1 {
		t.Errorf("unexpected line 1: %q", lines[1])
	}
	if lines[2] != expected2 {
		t.Errorf("unexpected line 2: %q", lines[2])
	}
}

func TestSave_OverwritesExistingManifest(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = m.Discard() }()

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta.go",
		Checksum:  "betaChecksum111111111111111",
		ChainHash: "betaChain1111111111111111111",
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[1] != "ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111" {
		t.Errorf("unexpected line 1: %q", lines[1])
	}
	if lines[2] != "ARTIFACT/beta;path:internal/beta.go;checksum:betaChecksum111111111111111;chain:betaChain1111111111111111111" {
		t.Errorf("unexpected line 2: %q", lines[2])
	}
}

func TestSave_ModifiedEntry(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:old-checksum111111111111111;chain:alphaChain111111111111111111",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = m.Discard() }()

	entry := m.Entries["ARTIFACT/alpha"]
	entry.Checksum = "new-checksum111111111111111"
	m.Entries["ARTIFACT/alpha"] = entry

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[1] != "ARTIFACT/alpha;path:internal/alpha.go;checksum:new-checksum111111111111111;chain:alphaChain111111111111111111" {
		t.Errorf("unexpected line 1: %q", lines[1])
	}
}

func TestSave_RemovedEntry(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111",
		"ARTIFACT/beta;path:internal/beta.go;checksum:betaChecksum111111111111111;chain:betaChain1111111111111111111",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = m.Discard() }()

	delete(m.Entries, "ARTIFACT/beta")

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[1] != "ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111" {
		t.Errorf("unexpected line: %q", lines[1])
	}
}

func TestSave_EmptyEntries(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111",
		"ARTIFACT/beta;path:internal/beta.go;checksum:betaChecksum111111111111111;chain:betaChain1111111111111111111",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = m.Discard() }()

	m.Entries = map[string]manifest.ManifestEntry{}

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %v", len(lines), lines)
	}
	if lines[0] != "code-from-spec: v5" {
		t.Errorf("unexpected header: %q", lines[0])
	}
}

func TestDiscard_DoesNotModifyFile(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"code-from-spec: v5",
		"ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111",
	})

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Entries["ARTIFACT/beta"] = manifest.ManifestEntry{
		Path:      "internal/beta.go",
		Checksum:  "betaChecksum111111111111111",
		ChainHash: "betaChain1111111111111111111",
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("Discard returned error: %v", err)
	}

	lines := readManifestFile(t)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[1] != "ARTIFACT/alpha;path:internal/alpha.go;checksum:alphaChecksum11111111111111;chain:alphaChain111111111111111111" {
		t.Errorf("unexpected line: %q", lines[1])
	}
	for _, l := range lines {
		if l == "ARTIFACT/beta;path:internal/beta.go;checksum:betaChecksum111111111111111;chain:betaChain1111111111111111111" {
			t.Error("beta entry should not be present after Discard")
		}
	}
}

func TestOpenManifest_InvalidHeader(t *testing.T) {
	testutils.Chdir(t)

	writeManifestFile(t, []string{
		"invalid-header",
	})

	_, err := manifest.OpenManifest(true)
	if !errors.Is(err, manifest.ErrManifestFormatError) {
		t.Errorf("expected ErrManifestFormatError, got %v", err)
	}
}

func TestReadOnly_SaveReturnsErrReadOnly(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Save(); !errors.Is(err, manifest.ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestReadOnly_DiscardReturnsErrReadOnly(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Discard(); !errors.Is(err, manifest.ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestClosed_DiscardAfterSave(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if err := m.Discard(); !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestClosed_SaveAfterDiscard(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("Discard returned error: %v", err)
	}

	if err := m.Save(); !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestClosed_SaveAfterSave(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Save(); err != nil {
		t.Fatalf("first Save returned error: %v", err)
	}

	if err := m.Save(); !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestClosed_DiscardAfterDiscard(t *testing.T) {
	testutils.Chdir(t)

	m, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Discard(); err != nil {
		t.Fatalf("first Discard returned error: %v", err)
	}

	if err := m.Discard(); !errors.Is(err, manifest.ErrManifestClosed) {
		t.Errorf("expected ErrManifestClosed, got %v", err)
	}
}

func TestConcurrency_ConcurrentReadersDoNotBlock(t *testing.T) {
	testutils.Chdir(t)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manifest.OpenManifest(true)
			if err != nil {
				errCh <- err
			}
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
		t.Fatal("concurrent readers timed out — possible deadlock")
	}

	close(errCh)
	for err := range errCh {
		t.Errorf("reader error: %v", err)
	}
}

func TestConcurrency_WriterBlocksReader(t *testing.T) {
	testutils.Chdir(t)

	writer, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error opening writer: %v", err)
	}

	readerStarted := make(chan struct{})
	readerDone := make(chan error, 1)

	go func() {
		close(readerStarted)
		_, err := manifest.OpenManifest(true)
		readerDone <- err
	}()

	<-readerStarted
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-readerDone:
		t.Fatalf("reader should be blocked, but returned with err=%v", err)
	default:
	}

	if err := writer.Discard(); err != nil {
		t.Fatalf("writer Discard returned error: %v", err)
	}

	select {
	case err := <-readerDone:
		if err != nil {
			t.Errorf("reader returned error after lock released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reader timed out after writer released lock")
	}
}

func TestConcurrency_WriterBlocksWriter(t *testing.T) {
	testutils.Chdir(t)

	writer1, err := manifest.OpenManifest(false)
	if err != nil {
		t.Fatalf("unexpected error opening first writer: %v", err)
	}

	writer2Started := make(chan struct{})
	writer2Done := make(chan error, 1)

	go func() {
		close(writer2Started)
		m, err := manifest.OpenManifest(false)
		if err != nil {
			writer2Done <- err
			return
		}
		writer2Done <- m.Discard()
	}()

	<-writer2Started
	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-writer2Done:
		t.Fatalf("second writer should be blocked, but returned with err=%v", err)
	default:
	}

	if err := writer1.Save(); err != nil {
		t.Fatalf("first writer Save returned error: %v", err)
	}

	select {
	case err := <-writer2Done:
		if err != nil {
			t.Errorf("second writer returned error after lock released: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second writer timed out after first writer released lock")
	}
}
