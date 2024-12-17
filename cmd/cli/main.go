package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Predefined allowed values for "type"
var allowedTypes = map[string]bool{
	"git":  true,
	"bin":  true,
	"fail": true,
}

var Exit = os.Exit

func main() {
	// Default mode is "collect" if no arguments are provided
	mode := "collect"

	// Check if mode is explicitly set as the first argument
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		mode = os.Args[1]                                      // Set mode to the first argument
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...) // Shift arguments for flag parsing
	}

	// Define command-line flags
	serverEnv := os.Getenv("SERVER_URL")
	keyEnv := os.Getenv("KEY")

	var (
		server string
		key    string
		typ    string
		target string
	)

	// Create a flag set for parsing arguments
	fs := flag.NewFlagSet(mode, flag.ExitOnError)
	fs.StringVar(&server, "server", serverEnv, "Server URL (default from SERVER_URL env)")
	fs.StringVar(&key, "key", keyEnv, "Authentication key (default from KEY env)")
	fs.StringVar(&typ, "type", "git", "Type of operation (bin or git, default git)")
	workDir, _ := os.Getwd()
	fs.StringVar(&target, "target", workDir, "Path to a file or directory on the filesystem, default '.'")

	// Parse flags
	err := fs.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("Error: Failed to parse arguments")
		Exit(1)
	}

	// Validate "type" flag if provided
	if typ != "" && !allowedTypes[typ] {
		fmt.Printf("Error: Invalid type '%s'. Allowed values are: bin, git, fail\n", typ)
		Exit(1)
	}

	// Validate "target" flag (optional, but must be valid if provided)
	if target != "" {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			fmt.Printf("Error: Target path '%s' does not exist\n", target)
			Exit(1)
		}
	}

	// Handle modes
	switch mode {
	case "collect":
		handleCollectMode(server, key, typ, target)
	case "gw":
		handleGWMode(server, key, typ, target)
	default:
		fmt.Printf("Error: Unknown mode '%s'. Supported modes are 'collect' and 'gw'.\n", mode)
		Exit(1)
	}
}

// handleCollectMode handles the "collect" mode logic
func handleCollectMode(server, key, typ, target string) {
	fmt.Println("Running in 'collect' mode")
	fmt.Printf("Server: %s\n", server)
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Type: %s\n", typ)
	fmt.Printf("Target: %s\n", target)

	// Example: Process the target path
	if target != "" {
		absPath, _ := filepath.Abs(target)
		fmt.Printf("Processing target: %s\n", absPath)
	}
	fmt.Println("Collect mode completed successfully")
}

// handleGWMode handles the "gw" mode logic
func handleGWMode(server, key, typ, target string) {
	fmt.Println("Running in 'gw' mode")
	fmt.Printf("Server: %s\n", server)
	fmt.Printf("Key: %s\n", key)
	fmt.Printf("Type: %s\n", typ)
	fmt.Printf("Target: %s\n", target)

	// Example: Simulate a condition for a non-zero exit code
	if typ == "fail" {
		fmt.Println("Error: Just fail in 'fail' mode")
		Exit(2) // Non-zero exit code
	}

	fmt.Println("GW mode completed successfully")
}
