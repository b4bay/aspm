package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var exitCode int

// Custom function to mock os.Exit in tests
func mockExit(code int) {
	exitCode = code
	return
}

type ASPMClientMock struct {
	endpoint string
	data     string
}

func (c *ASPMClientMock) Post(endpoint string, data interface{}) error {
	c.endpoint = endpoint
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	c.data = string(jsonData)
	fmt.Printf("POST to %s: %s\n", c.endpoint, c.data)
	return nil
}

// Helper function to capture stdout and stderr during test execution
func captureOutput(f func()) (string, string) {
	// Backup original stdout and stderr
	oldOut := os.Stdout
	oldErr := os.Stderr

	// Create pipes to capture outputs
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function
	f()

	// Restore original stdout and stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	// Read captured output
	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(rOut)
	errBuf.ReadFrom(rErr)

	return outBuf.String(), errBuf.String()
}

// Test default mode ("collect")
func TestDefaultMode(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "-type", "git"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'collect' mode") {
		t.Fatalf("Expected 'collect' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

// Test explicit "collect" mode
func TestCollectMode(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "collect", "-type", "bin", "/path/to/target", "/path/to/report"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'collect' mode") {
		t.Fatalf("Expected 'collect' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

// Test "gw" mode with valid input
func TestGWModeValid(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "gw", "-type", "bin", "./testfile"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'gw' mode") {
		t.Fatalf("Expected 'gw' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

// Test invalid "type" argument
func TestInvalidTypeArgument(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "collect", "-type", "invalid"}

	stdout, _ := captureOutput(func() { main() })

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for unknown mode, but got success. Output: %s", string(stdout))
	}

	if !strings.Contains(string(stdout), "Error: Invalid artefact type 'invalid'") {
		t.Fatalf("Expected error message for invalid type. Got output: %s", string(stdout))
	}
}

// Test unknown mode
func TestUnknownMode(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "unknownmode"}

	stdout, _ := captureOutput(func() { main() })

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for unknown mode, but got success. Output: %s", string(stdout))
	}

	if !strings.Contains(stdout, "Error: Unknown mode 'unknownmode'") {
		t.Fatalf("Expected error message for unknown mode. Got output: %s", stdout)
	}
}

func TestOriginMode(t *testing.T) {
	Exit = mockExit
	aspmClient = &ASPMClientMock{}

	artefactPath1 := createTempFileWithContent(t, "This is an artefact file 1.")
	defer os.Remove(artefactPath1)
	artefactPath2 := createTempFileWithContent(t, "This is an artefact file 2.")
	defer os.Remove(artefactPath2)
	artefactPath3 := createTempFileWithContent(t, "This is an artefact file 3.")
	defer os.Remove(artefactPath3)

	os.Args = []string{"main", "origin", "-method", "pack", artefactPath1, artefactPath2, artefactPath3}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'origin' mode") {
		t.Fatalf("Expected 'origin' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

func TestOriginModeValidGit(t *testing.T) {
	Exit = mockExit
	aspmClient = &ASPMClientMock{}

	// Set up a temporary Git repository for testing
	tempDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Test valid artefact as a Git repository
	os.Args = []string{"main", "origin", tempDir, tempDir}
	stdout, stderr := captureOutput(func() { main() })

	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}

	if !strings.Contains(stdout, "{\"") {
		t.Fatalf("Expected json in output, got: %s", stdout)
	}
}

func TestOriginModeInvalidGit(t *testing.T) {
	Exit = mockExit
	aspmClient = &ASPMClientMock{}

	nonGitDir := t.TempDir()
	// Test invalid artefact as non-Git repository
	os.Args = []string{"main", "origin", nonGitDir, nonGitDir}
	stdout, _ := captureOutput(func() { main() })

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for 'origin' mode, but got success. Output: %s", string(stdout))
	}

	if !strings.Contains(stdout, "not a git repository") {
		t.Fatalf("Expected error message for non-git repository. Got output: %s", stdout)
	}
}

func createTempFileWithContent(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.WriteString(content)
	tempFile.Close()
	return tempFile.Name()
}

func TestOriginMode_FileHashing(t *testing.T) {
	Exit = mockExit
	aspmClient = &ASPMClientMock{}

	// Create temporary artefact file
	artefactPath := createTempFileWithContent(t, "This is an artefact file.")
	defer os.Remove(artefactPath)

	// Create temporary source files
	source1Path := createTempFileWithContent(t, "Source file 1 content.")
	defer os.Remove(source1Path)

	source2Path := createTempFileWithContent(t, "Source file 2 content.")
	defer os.Remove(source2Path)

	os.Args = []string{"main", "origin", artefactPath, source1Path, source2Path}
	stdout, stderr := captureOutput(func() { main() })

	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}

	if !strings.Contains(stdout, "{\"") {
		t.Fatalf("Expected json in output, got: %s", stdout)
	}
}
