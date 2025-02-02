//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")

	// Get the current commit hash
	hash, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	date := time.Now().Format(time.RFC3339)

	cmd := exec.Command("go", "build", "-o", "bin/goschema", fmt.Sprintf("-X main.Commit=%s -X main.Date=%s", hash, date), ".")
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
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
