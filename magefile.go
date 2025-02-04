//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	fmt.Println("Building...")

	// Clear the bin folder
	if err := os.RemoveAll("bin"); err != nil {
		return fmt.Errorf("failed to clear bin folder: %w", err)
	}

	// Get the current commit hash
	hash, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	date := time.Now().Format(time.RFC3339)

	fmt.Println("Commit hash:", strings.TrimSpace(string(hash)))
	fmt.Println("Build date:", date)

	buildFlags := fmt.Sprintf("-X main.Commit=%s -X main.Date=%s", hash, date)

	cmd := exec.Command("go", "build", "-o", "bin/goschema", "-ldflags", buildFlags, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Tools() error {
	fmt.Println("Building tools...")
	cmd := exec.Command("go", "build", "-o", "bin/vaultdb", "./tools/vaultdb")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("bin/goschema", "/usr/local/bin/goschema")
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
