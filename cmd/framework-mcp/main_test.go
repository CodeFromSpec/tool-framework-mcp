// code-from-spec: SPEC/golang/tests/server@chYkmo0LDJoLMiItulpmd4mfujI
package main_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TestMain: failed to create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binaryPath = filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TestMain: failed to build binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", out)
	}
}

func TestHelpWord(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", out)
	}
}

func TestShortHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", out)
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderr.String())
	}
}

func TestMultipleArguments(t *testing.T) {
	cmd := exec.Command(binaryPath, "foo", "bar")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1, got exit 0")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderr.String())
	}
}

type mcpSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	nextID int
}

func testStartMCPSession(t *testing.T) *mcpSession {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), binaryPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("testStartMCPSession: stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("testStartMCPSession: stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("testStartMCPSession: start: %v", err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	})

	session := &mcpSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		nextID: 1,
	}
	testMCPHandshake(t, session)
	return session
}

func testMCPHandshake(t *testing.T, s *mcpSession) {
	t.Helper()
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      s.nextID,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "0.0.1",
			},
		},
	}
	s.nextID++
	testMCPSend(t, s, initReq)
	testMCPReadResponse(t, s)

	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	testMCPSend(t, s, notif)
}

func testMCPSend(t *testing.T, s *mcpSession, msg any) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("testMCPSend: marshal: %v", err)
	}
	data = append(data, '\n')
	if _, err := s.stdin.Write(data); err != nil {
		t.Fatalf("testMCPSend: write: %v", err)
	}
}

func testMCPReadResponse(t *testing.T, s *mcpSession) map[string]any {
	t.Helper()
	if !s.stdout.Scan() {
		if err := s.stdout.Err(); err != nil {
			t.Fatalf("testMCPReadResponse: scan error: %v", err)
		}
		t.Fatal("testMCPReadResponse: EOF before response")
	}
	line := s.stdout.Text()
	var result map[string]any
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		t.Fatalf("testMCPReadResponse: unmarshal: %v (line: %s)", err, line)
	}
	return result
}

func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	session := testStartMCPSession(t)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      session.nextID,
		"method":  "tools/list",
	}
	session.nextID++
	testMCPSend(t, session, req)
	resp := testMCPReadResponse(t, session)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	var loadChainTool map[string]any
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if tool["name"] == "load_chain" {
			loadChainTool = tool
			break
		}
	}
	if loadChainTool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	meta, ok := loadChainTool["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected _meta object on load_chain, got: %v", loadChainTool["_meta"])
	}
	maxSize, ok := meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("expected anthropic/maxResultSizeChars in _meta")
	}
	maxSizeFloat, ok := maxSize.(float64)
	if !ok {
		t.Fatalf("expected anthropic/maxResultSizeChars to be numeric, got: %T %v", maxSize, maxSize)
	}
	if int(maxSizeFloat) != 500000 {
		t.Errorf("expected anthropic/maxResultSizeChars to be 500000, got: %v", maxSizeFloat)
	}
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	session := testStartMCPSession(t)

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      session.nextID,
		"method":  "tools/list",
	}
	session.nextID++
	testMCPSend(t, session, req)
	resp := testMCPReadResponse(t, session)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	toolNames := make(map[string]bool)
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := tool["name"].(string); ok {
			toolNames[name] = true
		}
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Errorf("expected tool %q in tools/list response, but it was not found", name)
		}
	}
}
