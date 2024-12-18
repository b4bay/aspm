package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

var exitCode int

// Custom function to mock os.Exit in tests
func mockExit(code int) {
	exitCode = code
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
	os.Args = []string{"main", "-server", "https://test.com", "-key", "test-key", "-type", "git"}

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
	os.Args = []string{"main", "collect", "-server", "http://example.com", "-key", "testkey", "-type", "bin", "-scope", "/path/to/target"}

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
	os.Args = []string{"main", "gw", "-server", "https://test.com", "-key", "test-key", "-type", "bin", "./testfile"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'gw' mode") {
		t.Fatalf("Expected 'gw' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

// Test "gw" mode with error condition
func TestGWModeWithError(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "gw", "-type", "fail"}

	stdout, _ := captureOutput(func() { main() })

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for unknown mode, but got success. Output: %s", string(stdout))
	}

	if !strings.Contains(string(stdout), "Condition matched. Exiting with error") {
		t.Fatalf("Expected specific error message for missing target path. Got output: %s", string(stdout))
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

	if !strings.Contains(string(stdout), "Error: Invalid type 'invalid'") {
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
	os.Args = []string{"main", "origin", "-server", "https://test.com", "-key", "test-key", "-method", "pack", "-from", "bin", "-to", "bin", "/path/to/artefact", "/path/to/source1", "/path/to/source2"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'origin' mode") {
		t.Fatalf("Expected 'origin' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}
}

func TestOriginModeDefaults(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "origin", "/path/to/artefact", "/path/to/source"}

	stdout, stderr := captureOutput(func() { main() })

	if !strings.Contains(stdout, "Running in 'origin' mode") {
		t.Fatalf("Expected 'origin' mode to run. Got stdout: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Expected no errors. Got stderr: %s", stderr)
	}

}

func TestMissingParameters(t *testing.T) {
	Exit = mockExit
	os.Args = []string{"main", "origin"}

	stdout, _ := captureOutput(func() { main() })

	if exitCode == 0 {
		t.Fatalf("Expected non-zero exit code for 'origin' mode, but got success. Output: %s", string(stdout))
	}

	if !strings.Contains(stdout, "Error: at least artefact and one origin required") {
		t.Fatalf("Expected error message for 'origin' mode. Got output: %s", stdout)
	}

}
