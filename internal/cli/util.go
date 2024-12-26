package cli

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetEnvironment() map[string]string {
	var variablesMap = map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		variablesMap[pair[0]] = pair[1]
	}

	return variablesMap
}

func GetReports(paths []string) map[string]string {
	var reports = map[string]string{}
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: failed to access file %s: %v\n", path, err)
			continue
		}

		if fileInfo.IsDir() {
			fmt.Printf("Error: path %s is a directory, not a file\n", path)
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error: failed to read file %s: %v\n", path, err)
		}

		encodedContent := base64.StdEncoding.EncodeToString(content)
		reports[fileInfo.Name()] = encodedContent
	}

	return reports
}

func GetAuthorFromGit() string {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:'%an'")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
