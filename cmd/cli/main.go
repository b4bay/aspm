package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Exit = os.Exit

// Mode constants
const (
	ModeCollect = "collect"
	ModeGW      = "gw"
	ModeOrigin  = "origin"
)

// Default values for environment variables
const (
	DefaultServer = "http://localhost:8080"
	DefaultKey    = "default-key"
	DefaultType   = "git"
	DefaultMethod = "compile"
	DefaultFrom   = "git"
	DefaultTo     = "bin"
)

var (
	DefaultArtefact, _ = os.Getwd()
	DefaultScope, _    = os.Getwd()
)

var (
	DefaultModeName = "collect"
	DefaultMode     = collectMode
)

// Allowed values
var (
	AllowedTypes   = []string{"git", "bin", "fail"}
	AllowedMethods = []string{"compile", "pack"}
)

func isValidValue(value string, allowed []string) bool {
	for _, v := range allowed {
		if value == v {
			return true
		}
	}
	return false
}

func main() {
	var mode string

	// Check if mode is explicitly set as the first argument
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		mode = os.Args[1]                                      // Set mode to the first argument
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...) // Shift arguments for flag parsing
	} else {
		mode = DefaultModeName
	}

	args := os.Args[1:]

	switch mode {
	case ModeCollect:
		collectMode(args)
	case ModeGW:
		gwMode(args)
	case ModeOrigin:
		originMode(args)
	default:
		fmt.Printf("Error: Unknown mode '%s'. Supported modes are 'origin', 'collect' and 'gw'.\n", mode)
		Exit(1)
	}
}

func collectMode(args []string) {
	fs := flag.NewFlagSet(ModeCollect, flag.ExitOnError)
	server := fs.String("server", os.Getenv("SERVER"), "Server URL")
	key := fs.String("key", os.Getenv("KEY"), "Key string")
	typ := fs.String("type", DefaultType, "Type value")
	scope := fs.String("scope", DefaultScope, "Scope path")
	fs.Parse(args)

	if *typ != "" && !isValidValue(*typ, AllowedTypes) {
		fmt.Printf("Error: Invalid type '%s'\n", *typ)
		Exit(1)
	}

	unnamed := fs.Args()
	var artefact = DefaultArtefact
	if len(unnamed) > 1 {
		fmt.Println("Error: only one artefact allowed")
		Exit(1)
	}

	if len(unnamed) == 1 {
		artefact = unnamed[0]
	}

	fmt.Printf("Running in 'collect' mode: server=%s, key=%s, type=%s, scope=%s, artefact=%s\n", *server, *key, *typ, *scope, artefact)
}

func gwMode(args []string) {
	fs := flag.NewFlagSet(ModeGW, flag.ExitOnError)
	server := fs.String("server", os.Getenv("SERVER"), "Server URL")
	key := fs.String("key", os.Getenv("KEY"), "Key string")
	typ := fs.String("type", DefaultType, "Type value")
	scope := fs.String("scope", DefaultScope, "Scope path")
	fs.Parse(args)

	if *typ != "" && !isValidValue(*typ, AllowedTypes) {
		fmt.Printf("Error: Invalid type '%s'\n", *typ)
		Exit(1)
	}

	unnamed := fs.Args()
	var artefact = DefaultArtefact
	if len(unnamed) > 1 {
		fmt.Println("Error: only one artefact allowed")
		Exit(1)
	}

	if len(unnamed) == 1 {
		artefact = unnamed[0]
	}

	// Example condition for non-zero exit code
	if *typ == "fail" {
		fmt.Println("Condition matched. Exiting with error.")
		Exit(1)
	}

	fmt.Printf("Running in 'gw' mode: server=%s, key=%s, type=%s, scope=%s, artefact=%s\n", *server, *key, *typ, *scope, artefact)

}

func originMode(args []string) {
	var (
		artefact string
		sources  []string
	)

	fs := flag.NewFlagSet(ModeOrigin, flag.ExitOnError)
	server := fs.String("server", os.Getenv("SERVER"), "Server URL")
	key := fs.String("key", os.Getenv("KEY"), "Key string")
	method := fs.String("method", DefaultMethod, "Method (compile or pack)")
	from := fs.String("from", DefaultFrom, "From value")
	to := fs.String("to", DefaultTo, "To value")
	fs.Parse(args)

	unnamed := fs.Args()
	if len(unnamed) < 2 {
		fmt.Println("Error: at least artefact and one origin required")
		Exit(1)
	} else {
		artefact = unnamed[0]
		sources = unnamed[1:]
	}

	if !isValidValue(*method, AllowedMethods) {
		fmt.Printf("Error: Invalid method '%s'", *method)
		Exit(1)
	}

	if *from != "" && !isValidValue(*from, AllowedTypes) {
		fmt.Printf("Error: Invalid from '%s'", *from)
		Exit(1)
	}

	if *to != "" && !isValidValue(*to, AllowedTypes) {
		fmt.Printf("Error: Invalid to '%s'", *to)
		Exit(1)
	}

	fmt.Printf("Running in 'origin' mode: server=%s, key=%s, method=%s, from=%s, to=%s, artefact=%s, sources=%v\n", *server, *key, *method, *from, *to, artefact, sources)

	modeGit := func(path string) {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: cannot access '%s': %v\n", path, err)
			Exit(1)
		}

		if !info.IsDir() {
			fmt.Printf("Error: '%s' is not a directory\n", path)
			Exit(1)
		}

		// Check if the directory is a git repository
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Printf("Error: '%s' is not a git repository\n", path)
			Exit(1)
		}

		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error: running git command in '%s': %v\n %s\n", path, err, string(output))
			Exit(1)
		}

		fmt.Printf("%s => %s\n", path, strings.TrimSpace(string(output)))
	}

	modeBin := func(path string) {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: cannot access '%s': %v\n", path, err)
			Exit(1)
		}

		if info.IsDir() {
			fmt.Printf("Error: '%s' is not a file\n", path)
			Exit(1)
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error: cannot open file '%s': %v\n", path, err)
			Exit(1)
		}
		defer file.Close()

		hasher := sha1.New()
		if _, err := io.Copy(hasher, file); err != nil {
			fmt.Printf("Error: cannot calculate hash for file '%s': %v\n", path, err)
			Exit(1)
		}

		hash := hex.EncodeToString(hasher.Sum(nil))
		fmt.Printf("%s => %s\n", path, hash)
	}

	if *to == "git" {
		modeGit(artefact)
	} else if *to == "bin" {
		modeBin(artefact)
	}

	if *from == "git" {
		for _, source := range sources {
			modeGit(source)
		}
	} else if *from == "bin" {
		for _, source := range sources {
			modeBin(source)
		}
	}

}
