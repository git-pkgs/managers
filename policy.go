package managers

import (
	"context"
)

// Policy defines an interface for checks that run before package operations.
// Policies can inspect the operation details and either allow or deny execution.
type Policy interface {
	// Name returns a unique identifier for this policy.
	Name() string

	// Check evaluates the policy against the given operation.
	// Returns a PolicyResult indicating whether the operation should proceed.
	Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error)
}

// PolicyOperation contains details about the operation being checked.
type PolicyOperation struct {
	// Manager is the package manager name (e.g., "npm", "bundler").
	Manager string

	// Operation is the command being run (e.g., "add", "install", "update").
	Operation string

	// Packages is the list of packages being operated on.
	// Empty for operations like "install" that don't target specific packages.
	Packages []string

	// Args contains the raw arguments passed to the command.
	Args map[string]string

	// Flags contains the flags passed to the command.
	Flags map[string]any

	// WorkingDir is the directory where the operation will run.
	WorkingDir string

	// Command is the fully constructed command that will be executed.
	Command []string
}

// PolicyResult contains the outcome of a policy check.
type PolicyResult struct {
	// Allowed indicates whether the operation should proceed.
	Allowed bool

	// Reason explains why the operation was allowed or denied.
	Reason string

	// Warnings contains non-blocking issues that should be reported.
	Warnings []string

	// Metadata contains policy-specific data for programmatic access.
	Metadata map[string]any
}

// PolicyMode determines how policy violations are handled.
type PolicyMode int

const (
	// PolicyEnforce blocks operations that fail policy checks.
	PolicyEnforce PolicyMode = iota

	// PolicyWarn logs warnings but allows operations to proceed.
	PolicyWarn

	// PolicyDisabled skips all policy checks.
	PolicyDisabled
)

func (m PolicyMode) String() string {
	switch m {
	case PolicyEnforce:
		return "enforce"
	case PolicyWarn:
		return "warn"
	case PolicyDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// PolicyRunner wraps a Runner and applies policies before execution.
type PolicyRunner struct {
	inner    Runner
	policies []Policy
	mode     PolicyMode
	handler  PolicyHandler
}

// PolicyHandler receives policy check results for logging or custom handling.
type PolicyHandler interface {
	OnPolicyResult(op *PolicyOperation, policy Policy, result *PolicyResult)
}

// PolicyRunnerOption configures a PolicyRunner.
type PolicyRunnerOption func(*PolicyRunner)

// WithPolicies adds policies to the runner.
func WithPolicies(policies ...Policy) PolicyRunnerOption {
	return func(pr *PolicyRunner) {
		pr.policies = append(pr.policies, policies...)
	}
}

// WithPolicyMode sets the enforcement mode.
func WithPolicyMode(mode PolicyMode) PolicyRunnerOption {
	return func(pr *PolicyRunner) {
		pr.mode = mode
	}
}

// WithPolicyHandler sets a handler for policy results.
func WithPolicyHandler(handler PolicyHandler) PolicyRunnerOption {
	return func(pr *PolicyRunner) {
		pr.handler = handler
	}
}

// NewPolicyRunner creates a Runner that applies policies before execution.
func NewPolicyRunner(inner Runner, opts ...PolicyRunnerOption) *PolicyRunner {
	pr := &PolicyRunner{
		inner:    inner,
		policies: make([]Policy, 0),
		mode:     PolicyEnforce,
	}
	for _, opt := range opts {
		opt(pr)
	}
	return pr
}

// AddPolicy registers a policy to be checked before operations.
func (pr *PolicyRunner) AddPolicy(p Policy) {
	pr.policies = append(pr.policies, p)
}

// Run executes the command after checking all registered policies.
func (pr *PolicyRunner) Run(ctx context.Context, dir string, args ...string) (*Result, error) {
	if pr.mode == PolicyDisabled {
		return pr.inner.Run(ctx, dir, args...)
	}

	op := &PolicyOperation{
		WorkingDir: dir,
		Command:    args,
		Args:       make(map[string]string),
		Flags:      make(map[string]any),
	}

	// Extract manager and operation from command if possible
	if len(args) > 0 {
		op.Manager = args[0]
	}
	if len(args) > 1 {
		op.Operation = args[1]
	}

	for _, policy := range pr.policies {
		result, err := policy.Check(ctx, op)
		if err != nil {
			return nil, &ErrPolicyCheck{Policy: policy.Name(), Err: err}
		}

		if pr.handler != nil {
			pr.handler.OnPolicyResult(op, policy, result)
		}

		if !result.Allowed && pr.mode == PolicyEnforce {
			return nil, &ErrPolicyViolation{
				Policy:  policy.Name(),
				Reason:  result.Reason,
				Command: args,
			}
		}
	}

	return pr.inner.Run(ctx, dir, args...)
}

// RunWithContext executes the command with additional operation context.
// Use this when you have more information about the operation than just the command.
func (pr *PolicyRunner) RunWithContext(ctx context.Context, op *PolicyOperation) (*Result, error) {
	if pr.mode == PolicyDisabled {
		return pr.inner.Run(ctx, op.WorkingDir, op.Command...)
	}

	for _, policy := range pr.policies {
		result, err := policy.Check(ctx, op)
		if err != nil {
			return nil, &ErrPolicyCheck{Policy: policy.Name(), Err: err}
		}

		if pr.handler != nil {
			pr.handler.OnPolicyResult(op, policy, result)
		}

		if !result.Allowed && pr.mode == PolicyEnforce {
			return nil, &ErrPolicyViolation{
				Policy:  policy.Name(),
				Reason:  result.Reason,
				Command: op.Command,
			}
		}
	}

	return pr.inner.Run(ctx, op.WorkingDir, op.Command...)
}

// AllowAllPolicy is a no-op policy that allows all operations.
// Useful as a placeholder or for testing.
type AllowAllPolicy struct{}

func (AllowAllPolicy) Name() string { return "allow-all" }

func (AllowAllPolicy) Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error) {
	return &PolicyResult{Allowed: true}, nil
}

// DenyAllPolicy is a policy that denies all operations.
// Useful for testing or as a circuit breaker.
type DenyAllPolicy struct {
	Reason string
}

func (DenyAllPolicy) Name() string { return "deny-all" }

func (p DenyAllPolicy) Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error) {
	reason := p.Reason
	if reason == "" {
		reason = "all operations denied by policy"
	}
	return &PolicyResult{Allowed: false, Reason: reason}, nil
}

// PackageBlocklistPolicy denies operations on specific packages.
type PackageBlocklistPolicy struct {
	Blocked map[string]string // package name -> reason
}

func (PackageBlocklistPolicy) Name() string { return "package-blocklist" }

func (p PackageBlocklistPolicy) Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error) {
	for _, pkg := range op.Packages {
		if reason, blocked := p.Blocked[pkg]; blocked {
			return &PolicyResult{
				Allowed: false,
				Reason:  reason,
				Metadata: map[string]any{
					"blocked_package": pkg,
				},
			}, nil
		}
	}
	return &PolicyResult{Allowed: true}, nil
}
