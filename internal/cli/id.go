package cli

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Exit = os.Exit

func IdGit(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New("not a directory")
	}

	// Check if the directory is a git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "", errors.New("not a git repository")
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func IdBin(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		fmt.Printf("Error: '%s' is not a file\n", path)
		return "", errors.New("not a file")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	return strings.TrimSpace(hash), nil
}

func NameGit(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New("not a directory")
	}

	// Check if the directory is a git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "", errors.New("not a git repository")
	}

	// Gitlab cannot work with `git symbolic-ref HEAD` without additional checkout
	// https://stackoverflow.com/questions/69267025/detached-head-in-gitlab-ci-pipeline-how-to-push-correctly/69268083#69268083
	// Replaced with CI_COMMIT_REF_NAME approach
	/*
		cmd := exec.Command("git", "symbolic-ref", "HEAD")
		cmd.Dir = path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
	*/
	output := os.Getenv("CI_COMMIT_REF_NAME")

	return strings.TrimSpace(string(output)), nil
}

func NameBin(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		fmt.Printf("Error: '%s' is not a file\n", path)
		return "", errors.New("not a file")
	}

	name := filepath.Base(path)
	return name, nil
}
