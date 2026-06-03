// code-from-spec: ROOT/golang/tests/server@20pHBROsnbKM0AEsiPVAWLEmq6Q
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

var testBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "framework-mcp-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryName := "framework-mcp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	testBinaryPath = filepath.Join(dir, binaryName)

	cmd := exec.Command("go", "build", "-o", testBinaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "--help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("stdout does not contain usage message, got: %s", string(out))
	}
}

func TestHelpWordPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "help")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("stdout does not contain usage message, got: %s", string(out))
	}
}

func TestShortHelpFlagPrintsUsageToStdout(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "-h")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected exit 0, got: %v", err)
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Errorf("stdout does not contain usage message, got: %s", string(out))
	}
}

func TestUnrecognizedArgumentPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "something")
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
		t.Errorf("stderr does not contain usage message, got: %s", stderr.String())
	}
}

func TestMultipleArgumentsPrintsUsageToStderr(t *testing.T) {
	cmd := exec.Command(testBinaryPath, "foo", "bar")
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
		t.Errorf("stderr does not contain usage message, got: %s", stderr.String())
	}
}

type testMCPProcess struct {
	cmd    *exec.Cmd
	stdin  *os.File
	stdout *bufio.Reader
	nextID int
}

func testStartMCPProcess(t *testing.T) *testMCPProcess {
	t.Helper()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: create stdin pipe: %v", err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("testStartMCPProcess: create stdout pipe: %v", err)
	}

	cmd := exec.Command(testBinaryPath)
	cmd.Stdin = stdinR
	cmd.Stdout = stdoutW

	if err := cmd.Start(); err != nil {
		t.Fatalf("testStartMCPProcess: start binary: %v", err)
	}

	stdinR.Close()
	stdoutW.Close()

	p := &testMCPProcess{
		cmd:    cmd,
		stdin:  stdinW,
		stdout: bufio.NewReader(stdoutR),
		nextID: 1,
	}

	t.Cleanup(func() {
		p.stdin.Close()
		cmd.Wait()
	})

	return p
}

func (p *testMCPProcess) testSendLine(t *testing.T, line string) {
	t.Helper()
	_, err := fmt.Fprintln(p.stdin, line)
	if err != nil {
		t.Fatalf("testSendLine: %v", err)
	}
}

func (p *testMCPProcess) testReadLine(t *testing.T) string {
	t.Helper()
	line, err := p.stdout.ReadString('\n')
	if err != nil {
		t.Fatalf("testReadLine: %v", err)
	}
	return strings.TrimRight(line, "\r\n")
}

func (p *testMCPProcess) testHandshake(t *testing.T) {
	t.Helper()

	id := p.nextID
	p.nextID++

	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "0.0.1",
			},
		},
	}
	b, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("testHandshake: marshal initialize: %v", err)
	}
	p.testSendLine(t, string(b))
	p.testReadLine(t)

	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	b, err = json.Marshal(notif)
	if err != nil {
		t.Fatalf("testHandshake: marshal initialized notification: %v", err)
	}
	p.testSendLine(t, string(b))
}

func (p *testMCPProcess) testSendRequest(t *testing.T, method string, params any) map[string]any {
	t.Helper()

	id := p.nextID
	p.nextID++

	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("testSendRequest: marshal request: %v", err)
	}
	p.testSendLine(t, string(b))

	line := p.testReadLine(t)
	var resp map[string]any
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatalf("testSendRequest: unmarshal response: %v\nline: %s", err, line)
	}
	return resp
}

func TestToolsListAdvertisesMaxResultSizeCharsForLoadChain(t *testing.T) {
	p := testStartMCPProcess(t)
	p.testHandshake(t)

	resp := p.testSendRequest(t, "tools/list", nil)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("response has no result field: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("result has no tools field: %v", result)
	}

	var loadChainTool map[string]any
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if toolMap["name"] == "load_chain" {
			loadChainTool = toolMap
			break
		}
	}
	if loadChainTool == nil {
		t.Fatal("load_chain tool not found in tools/list response")
	}

	meta, ok := loadChainTool["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("load_chain tool has no _meta field: %v", loadChainTool)
	}

	maxSize, ok := meta["anthropic/maxResultSizeChars"]
	if !ok {
		t.Fatal("load_chain _meta missing anthropic/maxResultSizeChars")
	}

	maxSizeFloat, ok := maxSize.(float64)
	if !ok {
		t.Fatalf("anthropic/maxResultSizeChars is not a number: %T %v", maxSize, maxSize)
	}
	if maxSizeFloat != 500000 {
		t.Errorf("expected anthropic/maxResultSizeChars = 500000, got %v", maxSizeFloat)
	}
}

func TestToolsListAdvertisesAllTools(t *testing.T) {
	p := testStartMCPProcess(t)
	p.testHandshake(t)

	resp := p.testSendRequest(t, "tools/list", nil)

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("response has no result field: %v", resp)
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("result has no tools field: %v", result)
	}

	found := map[string]bool{}
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := toolMap["name"].(string); ok {
			found[name] = true
		}
	}

	expected := []string{"load_chain", "write_file", "validate_specs", "chain_hash", "version"}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("tool %q not found in tools/list response", name)
		}
	}
}
