//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
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

	platforms := []string{"linux", "darwin", "windows"}
	archs := []string{"amd64"}

	const (
		output = "bin/vaultdb"
		source = "./tools/vaultdb"
	)

	for _, platform := range platforms {
		for _, arch := range archs {
			if err := buildBinary(platform, arch, output, source, platform, arch); err != nil {
				return fmt.Errorf("failed to build for %s/%s: %w", platform, arch, err)
			}
		}
	}

	return nil
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
	os.RemoveAll("bin")
}

// Build the release binaries
func Release() error {
	mg.Deps(Clean)
	Tools()

	fmt.Println("Building...")

	platforms := []string{"linux", "darwin", "windows"}
	archs := []string{"amd64"}

	const (
		output = "bin/goschema"
		source = "."
	)

	for _, platform := range platforms {
		for _, arch := range archs {
			if err := buildBinary(platform, arch, output, source, platform, arch); err != nil {
				return fmt.Errorf("failed to build for %s/%s: %w", platform, arch, err)
			}
		}
	}

	return nil
}

func buildBinary(platform, arch, output, source string, suffixes ...string) error {
	fmt.Println("Building for", platform)

	// Get the current commit hash
	hash, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	date := time.Now().Format(time.RFC3339)

	fmt.Println("Commit hash:", strings.TrimSpace(string(hash)))
	fmt.Println("Build date:", date)

	buildFlags := fmt.Sprintf("-X main.Commit=%s -X main.Date=%s", hash, date)

	// Store the current GOOS and GOARCH
	defer func() {
		fmt.Println("Restoring GOOS and GOARCH to", runtime.GOOS, runtime.GOARCH)
		os.Setenv("GOOS", runtime.GOOS)
		os.Setenv("GOARCH", runtime.GOARCH)
	}()

	// Set the GOOS and GOARCH to the desired platform
	os.Setenv("GOOS", platform)
	os.Setenv("GOARCH", arch)

	if len(suffixes) > 0 {
		output += "-" + strings.Join(suffixes, "-")
	}

	return sh.Run("go", "build", "-o", output, "-ldflags", buildFlags, source)
}
