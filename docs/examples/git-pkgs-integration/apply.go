// Example: How to add a "deps apply" command to git-pkgs
//
// This shows the integration pattern for adding the managers library
// to git-pkgs as a new subcommand.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-pkgs/managers"
	"github.com/git-pkgs/managers/definitions"
)

// ApplyOptions configures the apply command
type ApplyOptions struct {
	RepoPath   string
	DryRun     bool
	UpdateType string // "all", "patch", "minor"
	Package    string // specific package to update, empty for all
}

// ApplyResult tracks what was updated
type ApplyResult struct {
	Updated []UpdatedPackage
	Skipped []SkippedPackage
	Errors  []UpdateError
}

type UpdatedPackage struct {
	Name    string
	From    string
	To      string
	Manager string
}

type SkippedPackage struct {
	Name   string
	Reason string
}

type UpdateError struct {
	Name  string
	Error string
}

// Apply runs the dependency update process
func Apply(ctx context.Context, opts ApplyOptions) (*ApplyResult, error) {
	// Initialize the managers library
	translator, err := initTranslator()
	if err != nil {
		return nil, err
	}

	// Detect which package manager to use based on lockfiles
	manager, err := detectManagerFromLockfiles(opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("detecting package manager: %w", err)
	}

	// Get outdated packages from git-pkgs
	outdated, err := getGitPkgsOutdated(opts.RepoPath, opts.UpdateType)
	if err != nil {
		return nil, fmt.Errorf("getting outdated packages: %w", err)
	}

	// Filter to specific package if requested
	if opts.Package != "" {
		var filtered []OutdatedPackage
		for _, pkg := range outdated {
			if pkg.Name == opts.Package {
				filtered = append(filtered, pkg)
			}
		}
		outdated = filtered
	}

	result := &ApplyResult{}

	for _, pkg := range outdated {
		// Map ecosystem to our manager
		pkgManager := ecosystemToManagerWithFallback(pkg.Ecosystem, manager)
		if pkgManager == "" {
			result.Skipped = append(result.Skipped, SkippedPackage{
				Name:   pkg.Name,
				Reason: fmt.Sprintf("unsupported ecosystem: %s", pkg.Ecosystem),
			})
			continue
		}

		// Build the update command
		cmd, err := translator.BuildCommand(pkgManager, "update", managers.CommandInput{
			Args: map[string]string{"package": pkg.Name},
		})
		if err != nil {
			result.Errors = append(result.Errors, UpdateError{
				Name:  pkg.Name,
				Error: fmt.Sprintf("building command: %v", err),
			})
			continue
		}

		if opts.DryRun {
			fmt.Printf("[dry-run] Would run: %s\n", strings.Join(cmd, " "))
			result.Updated = append(result.Updated, UpdatedPackage{
				Name:    pkg.Name,
				From:    pkg.CurrentVersion,
				To:      pkg.LatestVersion,
				Manager: pkgManager,
			})
			continue
		}

		// Execute the update
		if err := executeCommand(ctx, cmd, opts.RepoPath); err != nil {
			result.Errors = append(result.Errors, UpdateError{
				Name:  pkg.Name,
				Error: err.Error(),
			})
			continue
		}

		result.Updated = append(result.Updated, UpdatedPackage{
			Name:    pkg.Name,
			From:    pkg.CurrentVersion,
			To:      pkg.LatestVersion,
			Manager: pkgManager,
		})
	}

	return result, nil
}

func initTranslator() (*managers.Translator, error) {
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		return nil, err
	}

	translator := managers.NewTranslator()
	for _, def := range defs {
		translator.Register(def)
	}
	return translator, nil
}

// detectManagerFromLockfiles finds the package manager based on lockfile presence
// This is more reliable than just using the ecosystem name
func detectManagerFromLockfiles(repoPath string) (string, error) {
	// Priority order matters - more specific lockfiles first
	checks := []struct {
		file    string
		manager string
	}{
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
		{"Gemfile.lock", "bundler"},
		{"Cargo.lock", "cargo"},
		{"go.sum", "gomod"},
		{"uv.lock", "uv"},
	}

	for _, check := range checks {
		path := filepath.Join(repoPath, check.file)
		if _, err := os.Stat(path); err == nil {
			return check.manager, nil
		}
	}

	return "", fmt.Errorf("no supported lockfile found")
}

// ecosystemToManagerWithFallback maps ecosystem names with a detected fallback
func ecosystemToManagerWithFallback(ecosystem, detected string) string {
	eco := strings.ToLower(ecosystem)

	// For npm ecosystem, use detected manager (npm/pnpm/yarn)
	if eco == "npm" {
		switch detected {
		case "npm", "pnpm", "yarn":
			return detected
		}
		return "npm"
	}

	// Direct mappings
	mapping := map[string]string{
		"rubygems": "bundler",
		"cargo":    "cargo",
		"go":       "gomod",
		"pypi":     "uv",
	}

	if m, ok := mapping[eco]; ok {
		return m
	}
	return ""
}

// getGitPkgsOutdated calls git-pkgs outdated --json and parses the output
func getGitPkgsOutdated(repoPath, updateType string) ([]OutdatedPackage, error) {
	args := []string{"outdated", "--format", "json"}
	if updateType == "patch" {
		// No filter needed, include all
	} else if updateType == "minor" {
		args = append(args, "--minor") // Skip patch-only
	}

	cmd := exec.Command("git-pkgs", args...)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		// git-pkgs outdated returns non-zero when packages are outdated
		if exitErr, ok := err.(*exec.ExitError); ok {
			output = exitErr.Stderr
		} else {
			return nil, err
		}
	}

	var result []OutdatedPackage
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parsing git-pkgs output: %w", err)
	}

	return result, nil
}

func executeCommand(ctx context.Context, args []string, dir string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
