// code-from-spec: SPEC/golang/test/cases/server@gkBv3jkiwgS6vrsq9W3uwryKVFU
package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	name := "framework-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binaryPath = filepath.Join(dir, name)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestHelpWord(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestShortHelpFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("expected stdout to contain usage message, got: %s", string(out))
	}
}

func TestUnrecognizedArgument(t *testing.T) {
	cmd := exec.Command(binaryPath, "something")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got: %d", exitErr.ExitCode())
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
		t.Fatal("expected non-zero exit code")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got: %d", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("expected stderr to contain usage message, got: %s", stderr.String())
	}
}

type mcpSession struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
	nextID int
}

func startMCPSession(t *testing.T) *mcpSession {
	t.Helper()
	cmd := exec.Command(binaryPath)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	t.Cleanup(func() {
		stdinPipe.Close()
		cmd.Wait()
	})

	s := &mcpSession{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdinPipe),
		stdout: bufio.NewReader(stdoutPipe),
		nextID: 1,
	}

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      s.nextID,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	s.nextID++
	if err := s.send(initReq); err != nil {
		t.Fatalf("failed to send initialize: %v", err)
	}
	if _, err := s.readLine(); err != nil {
		t.Fatalf("failed to read initialize response: %v", err)
	}

	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	if err := s.send(notif); err != nil {
		t.Fatalf("failed to send notifications/initialized: %v", err)
	}

	return s
}

func (s *mcpSession) send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := s.stdin.Write(data); err != nil {
		return err
	}
	if err := s.stdin.WriteByte('\n'); err != nil {
		return err
	}
	return s.stdin.Flush()
}

func (s *mcpSession) readLine() (map[string]any, error) {
	line, err := s.stdout.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w (line: %s)", err, line)
	}
	return result, nil
}

func (s *mcpSession) request(method string, params any) (map[string]any, error) {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      s.nextID,
		"method":  method,
		"params":  params,
	}
	s.nextID++
	if err := s.send(req); err != nil {
		return nil, err
	}
	return s.readLine()
}

func TestToolsListMaxResultSizeChars(t *testing.T) {
	s := startMCPSession(t)

	resp, err := s.request("tools/list", map[string]any{})
	if err != nil {
		t.Fatalf("failed to send tools/list: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if toolMap["name"] == "load_chain" {
			meta, ok := toolMap["_meta"].(map[string]any)
			if !ok {
				t.Fatalf("expected _meta for load_chain, got: %v", toolMap)
			}
			val, ok := meta["anthropic/maxResultSizeChars"]
			if !ok {
				t.Fatal("expected anthropic/maxResultSizeChars in _meta")
			}
			numVal, ok := val.(float64)
			if !ok {
				t.Fatalf("expected numeric value, got: %T %v", val, val)
			}
			if int(numVal) != 500000 {
				t.Errorf("expected 500000, got: %v", numVal)
			}
			return
		}
	}
	t.Fatal("load_chain tool not found in tools list")
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	s := startMCPSession(t)

	resp, err := s.request("tools/list", map[string]any{})
	if err != nil {
		t.Fatalf("failed to send tools/list: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("expected tools array, got: %v", result)
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "accept", "dump_chain", "version"}
	found := make(map[string]bool)
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := toolMap["name"].(string); ok {
			found[name] = true
		}
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("tool %q not found in tools list", name)
		}
	}
}

func TestVersionToolReturnsVersionString(t *testing.T) {
	s := startMCPSession(t)

	resp, err := s.request("tools/call", map[string]any{
		"name":      "version",
		"arguments": map[string]any{},
	})
	if err != nil {
		t.Fatalf("failed to call version tool: %v", err)
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result object, got: %v", resp)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("expected content array, got: %v", result)
	}
	first, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("expected content item to be object, got: %v", content[0])
	}
	text, ok := first["text"].(string)
	if !ok {
		t.Fatalf("expected text field, got: %v", first)
	}
	if !strings.Contains(text, "dev") {
		t.Errorf("expected version string to contain 'dev', got: %s", text)
	}
}
