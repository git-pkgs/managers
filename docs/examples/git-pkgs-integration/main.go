// Example: git-pkgs integration with the managers library
//
// This shows how git-pkgs could add a "deps apply" command that:
// 1. Uses git-pkgs existing outdated detection (via ecosyste.ms API)
// 2. Uses the managers library to run actual update commands
//
// The managers library bridges the gap between "knowing what's outdated"
// and "actually updating it".

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/git-pkgs/managers"
	"github.com/git-pkgs/managers/definitions"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repo-path>\n", os.Args[0])
		os.Exit(1)
	}

	repoPath := os.Args[1]
	ctx := context.Background()

	if err := run(ctx, repoPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// OutdatedPackage represents what git-pkgs outdated returns
type OutdatedPackage struct {
	Name           string
	Ecosystem      string
	CurrentVersion string
	LatestVersion  string
	UpdateType     string // major, minor, patch
}

func run(ctx context.Context, repoPath string) error {
	// Load the managers library
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		return fmt.Errorf("loading definitions: %w", err)
	}

	translator := managers.NewTranslator()
	for _, def := range defs {
		translator.Register(def)
	}

	// In real git-pkgs, this would come from the existing outdated command
	// which queries the ecosyste.ms API for latest versions
	outdated := getOutdatedFromGitPkgs(repoPath)

	if len(outdated) == 0 {
		fmt.Println("All dependencies are up to date!")
		return nil
	}

	fmt.Printf("Found %d outdated dependencies\n\n", len(outdated))

	// Group by ecosystem so we can batch updates
	byEcosystem := make(map[string][]OutdatedPackage)
	for _, pkg := range outdated {
		byEcosystem[pkg.Ecosystem] = append(byEcosystem[pkg.Ecosystem], pkg)
	}

	// Apply updates for each ecosystem
	for ecosystem, packages := range byEcosystem {
		manager := ecosystemToManager(ecosystem)
		if manager == "" {
			fmt.Printf("Skipping %s packages (no manager mapping)\n", ecosystem)
			continue
		}

		fmt.Printf("Updating %d %s packages...\n", len(packages), ecosystem)

		for _, pkg := range packages {
			if err := applyUpdate(ctx, translator, manager, repoPath, pkg); err != nil {
				fmt.Printf("  Warning: failed to update %s: %v\n", pkg.Name, err)
				continue
			}
			fmt.Printf("  Updated %s: %s -> %s\n", pkg.Name, pkg.CurrentVersion, pkg.LatestVersion)
		}
	}

	return nil
}

// ecosystemToManager maps git-pkgs ecosystem names to managers library names
// git-pkgs uses ecosyste.ms ecosystem names, which may differ from our manager names
func ecosystemToManager(ecosystem string) string {
	mapping := map[string]string{
		"npm":       "npm",      // or could be pnpm/yarn based on lockfile
		"rubygems":  "bundler",
		"cargo":     "cargo",
		"go":        "gomod",
		"pypi":      "uv",       // or pip, depending on project
		"packagist": "",        // not yet supported
		"nuget":     "",        // not yet supported
		"maven":     "",        // not yet supported
		"hex":       "",        // not yet supported
	}
	return mapping[strings.ToLower(ecosystem)]
}

// applyUpdate runs the package manager update command
func applyUpdate(ctx context.Context, tr *managers.Translator, manager, repoPath string, pkg OutdatedPackage) error {
	// Build the update command using the managers library
	cmd, err := tr.BuildCommand(manager, "update", managers.CommandInput{
		Args: map[string]string{"package": pkg.Name},
	})
	if err != nil {
		return fmt.Errorf("building command: %w", err)
	}

	// Execute it
	return runCommand(ctx, cmd, repoPath)
}

// runCommand executes a command in the specified directory
func runCommand(ctx context.Context, args []string, dir string) error {
	_, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// In a real implementation, use os/exec
	// Here we just print what would run
	fmt.Printf("    Would run: %s\n", strings.Join(args, " "))
	return nil
}

// getOutdatedFromGitPkgs simulates calling git-pkgs outdated --json
// In real code, this would either:
// 1. Import and call git-pkgs/cmd.runOutdated directly
// 2. Execute `git-pkgs outdated --json` and parse output
// 3. Use the same ecosyste.ms client that git-pkgs uses
func getOutdatedFromGitPkgs(repoPath string) []OutdatedPackage {
	// Simulated data - in practice this comes from git-pkgs
	return []OutdatedPackage{
		{
			Name:           "lodash",
			Ecosystem:      "npm",
			CurrentVersion: "4.17.20",
			LatestVersion:  "4.17.21",
			UpdateType:     "patch",
		},
		{
			Name:           "rails",
			Ecosystem:      "rubygems",
			CurrentVersion: "7.0.0",
			LatestVersion:  "7.1.0",
			UpdateType:     "minor",
		},
		{
			Name:           "serde",
			Ecosystem:      "cargo",
			CurrentVersion: "1.0.190",
			LatestVersion:  "1.0.195",
			UpdateType:     "patch",
		},
	}
}
