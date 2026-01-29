// Example: Simple dependabot-like tool using the managers library
//
// This example shows how to build a cron-based dependency updater that:
// 1. Detects the package manager for a project
// 2. Checks for outdated dependencies
// 3. Updates each dependency in a separate branch
// 4. Creates a pull request for each update
//
// Run with: go run main.go /path/to/repo
//
// For cron usage, add to crontab:
//   0 9 * * 1 /path/to/dependabot-cron /path/to/repo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

func run(ctx context.Context, repoPath string) error {
	// Load definitions and create translator
	defs, err := definitions.LoadEmbedded()
	if err != nil {
		return fmt.Errorf("loading definitions: %w", err)
	}

	translator := managers.NewTranslator()
	for _, def := range defs {
		translator.Register(def)
	}

	// Detect package manager
	managerName, err := detectManager(repoPath)
	if err != nil {
		return fmt.Errorf("detecting manager: %w", err)
	}
	fmt.Printf("Detected package manager: %s\n", managerName)

	// Get outdated dependencies
	outdated, err := getOutdated(ctx, translator, managerName, repoPath)
	if err != nil {
		return fmt.Errorf("checking outdated: %w", err)
	}

	if len(outdated) == 0 {
		fmt.Println("All dependencies are up to date!")
		return nil
	}

	fmt.Printf("Found %d outdated dependencies\n", len(outdated))

	// Update each dependency
	for _, dep := range outdated {
		if err := updateDependency(ctx, translator, managerName, repoPath, dep); err != nil {
			fmt.Printf("Warning: failed to update %s: %v\n", dep.Name, err)
			continue
		}
	}

	return nil
}

// Dependency represents an outdated package
type Dependency struct {
	Name    string
	Current string
	Latest  string
}

// detectManager finds the package manager for a repository
// In a real implementation, this would use the detector from the managers library
func detectManager(repoPath string) (string, error) {
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
		path := repoPath + "/" + check.file
		if _, err := os.Stat(path); err == nil {
			return check.manager, nil
		}
	}

	return "", fmt.Errorf("no supported package manager found")
}

// getOutdated returns a list of outdated dependencies
func getOutdated(ctx context.Context, tr *managers.Translator, managerName, repoPath string) ([]Dependency, error) {
	// Build the outdated command
	cmd, err := tr.BuildCommand(managerName, "outdated", managers.CommandInput{})
	if err != nil {
		return nil, err
	}

	// Execute it
	result, err := runCommand(ctx, cmd, repoPath)
	if err != nil {
		// Many package managers return non-zero when there are outdated deps
		// so we check if we got output
		if result == nil || len(result.Stdout) == 0 {
			return nil, err
		}
	}

	// Parse the output (this is manager-specific)
	return parseOutdated(managerName, result.Stdout)
}

// parseOutdated parses the outdated command output
// Each manager has different JSON formats
func parseOutdated(managerName string, output []byte) ([]Dependency, error) {
	var deps []Dependency

	switch managerName {
	case "npm", "pnpm", "yarn":
		// npm outdated --json returns: {"package": {"current": "1.0", "latest": "2.0"}}
		var data map[string]struct {
			Current string `json:"current"`
			Latest  string `json:"latest"`
		}
		if err := json.Unmarshal(output, &data); err != nil {
			return nil, err
		}
		for name, info := range data {
			deps = append(deps, Dependency{
				Name:    name,
				Current: info.Current,
				Latest:  info.Latest,
			})
		}

	case "bundler":
		// bundler outdated --parseable returns: gem (current < latest)
		// Not JSON, need to parse text
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			// Parse "gem (current < latest)" format
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				deps = append(deps, Dependency{Name: parts[0]})
			}
		}

	case "cargo":
		// cargo doesn't have native outdated, would need cargo-outdated
		return nil, fmt.Errorf("cargo outdated requires cargo-outdated extension")

	case "gomod":
		// go list -m -u -json all returns JSON lines
		decoder := json.NewDecoder(bytes.NewReader(output))
		for decoder.More() {
			var mod struct {
				Path    string `json:"Path"`
				Version string `json:"Version"`
				Update  *struct {
					Version string `json:"Version"`
				} `json:"Update"`
			}
			if err := decoder.Decode(&mod); err != nil {
				continue
			}
			if mod.Update != nil {
				deps = append(deps, Dependency{
					Name:    mod.Path,
					Current: mod.Version,
					Latest:  mod.Update.Version,
				})
			}
		}

	case "uv":
		// uv tree --outdated output needs parsing
		// For now, return empty - would need proper implementation
		return nil, fmt.Errorf("uv outdated parsing not yet implemented")
	}

	return deps, nil
}

// updateDependency updates a single dependency in its own branch
func updateDependency(ctx context.Context, tr *managers.Translator, managerName, repoPath string, dep Dependency) error {
	branchName := fmt.Sprintf("deps/%s-%s", dep.Name, dep.Latest)
	fmt.Printf("Updating %s to %s (branch: %s)\n", dep.Name, dep.Latest, branchName)

	// Create a new branch
	if err := gitCommand(repoPath, "checkout", "-b", branchName); err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}
	defer func() { _ = gitCommand(repoPath, "checkout", "-") }() // Return to original branch

	// Build and run the update command
	cmd, err := tr.BuildCommand(managerName, "update", managers.CommandInput{
		Args: map[string]string{"package": dep.Name},
	})
	if err != nil {
		return err
	}

	if _, err := runCommand(ctx, cmd, repoPath); err != nil {
		return fmt.Errorf("running update: %w", err)
	}

	// Commit the changes
	if err := gitCommand(repoPath, "add", "."); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("Update %s to %s", dep.Name, dep.Latest)
	if err := gitCommand(repoPath, "commit", "-m", commitMsg); err != nil {
		return err
	}

	// Push and create PR (using gh cli)
	if err := gitCommand(repoPath, "push", "-u", "origin", branchName); err != nil {
		return fmt.Errorf("pushing branch: %w", err)
	}

	prBody := fmt.Sprintf("Updates %s from %s to %s\n\nGenerated by dependabot-cron", dep.Name, dep.Current, dep.Latest)
	if err := ghCommand(repoPath, "pr", "create", "--title", commitMsg, "--body", prBody); err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	fmt.Printf("Created PR for %s\n", dep.Name)
	return nil
}

// CommandResult holds the output of a command
type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// runCommand executes a command and returns the result
func runCommand(ctx context.Context, args []string, dir string) (*CommandResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &CommandResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: cmd.ProcessState.ExitCode(),
	}

	return result, err
}

// gitCommand runs a git command in the repo
func gitCommand(repoPath string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ghCommand runs a GitHub CLI command
func ghCommand(repoPath string, args ...string) error {
	cmd := exec.Command("gh", args...)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
