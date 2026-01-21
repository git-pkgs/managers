package managers

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPolicyRunnerAllowsWhenNoPolicies(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock)

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Fatalf("expected 1 captured command, got %d", len(mock.Captured))
	}
	expected := []string{"npm", "install"}
	if !reflect.DeepEqual(mock.Captured[0], expected) {
		t.Errorf("got %v, want %v", mock.Captured[0], expected)
	}
}

func TestPolicyRunnerAllowAllPolicy(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(AllowAllPolicy{}))

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install", "lodash")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Errorf("expected command to be executed")
	}
}

func TestPolicyRunnerDenyAllPolicy(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(DenyAllPolicy{Reason: "testing"}))

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install", "lodash")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var violation *ErrPolicyViolation
	if !errors.As(err, &violation) {
		t.Fatalf("expected ErrPolicyViolation, got %T", err)
	}

	if violation.Policy != "deny-all" {
		t.Errorf("got policy %q, want %q", violation.Policy, "deny-all")
	}

	if len(mock.Captured) != 0 {
		t.Errorf("expected no commands executed, got %d", len(mock.Captured))
	}
}

func TestPolicyRunnerWarnMode(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock,
		WithPolicies(DenyAllPolicy{Reason: "testing"}),
		WithPolicyMode(PolicyWarn),
	)

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install", "lodash")
	if err != nil {
		t.Fatalf("expected no error in warn mode, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Errorf("expected command to be executed in warn mode")
	}
}

func TestPolicyRunnerDisabledMode(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock,
		WithPolicies(DenyAllPolicy{Reason: "testing"}),
		WithPolicyMode(PolicyDisabled),
	)

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install", "lodash")
	if err != nil {
		t.Fatalf("expected no error in disabled mode, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Errorf("expected command to be executed in disabled mode")
	}
}

func TestPolicyRunnerMultiplePolicies(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(
		AllowAllPolicy{},
		AllowAllPolicy{},
	))

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Errorf("expected command to be executed")
	}
}

func TestPolicyRunnerFirstDenyWins(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(
		AllowAllPolicy{},
		DenyAllPolicy{Reason: "second policy"},
		AllowAllPolicy{},
	))

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var violation *ErrPolicyViolation
	if !errors.As(err, &violation) {
		t.Fatalf("expected ErrPolicyViolation, got %T", err)
	}

	if len(mock.Captured) != 0 {
		t.Errorf("expected no commands executed")
	}
}

func TestPackageBlocklistPolicy(t *testing.T) {
	policy := PackageBlocklistPolicy{
		Blocked: map[string]string{
			"evil-package":   "known malware",
			"deprecated-lib": "no longer maintained",
		},
	}

	tests := []struct {
		name     string
		packages []string
		allowed  bool
	}{
		{"allowed package", []string{"lodash"}, true},
		{"blocked package", []string{"evil-package"}, false},
		{"mixed packages", []string{"lodash", "deprecated-lib"}, false},
		{"empty packages", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &PolicyOperation{Packages: tt.packages}
			result, err := policy.Check(context.Background(), op)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Allowed != tt.allowed {
				t.Errorf("got allowed=%v, want %v", result.Allowed, tt.allowed)
			}
		})
	}
}

func TestPolicyRunnerWithContext(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(AllowAllPolicy{}))

	op := &PolicyOperation{
		Manager:    "npm",
		Operation:  "add",
		Packages:   []string{"lodash"},
		WorkingDir: "/tmp",
		Command:    []string{"npm", "install", "lodash"},
	}

	_, err := pr.RunWithContext(context.Background(), op)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.Captured) != 1 {
		t.Fatalf("expected 1 captured command, got %d", len(mock.Captured))
	}
}

func TestPolicyRunnerAddPolicy(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock)
	pr.AddPolicy(DenyAllPolicy{Reason: "added later"})

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err == nil {
		t.Fatalf("expected error after adding deny policy")
	}
}

type handlerRecorder struct {
	results []*PolicyResult
}

func (h *handlerRecorder) OnPolicyResult(op *PolicyOperation, policy Policy, result *PolicyResult) {
	h.results = append(h.results, result)
}

func TestPolicyRunnerHandler(t *testing.T) {
	mock := NewMockRunner()
	handler := &handlerRecorder{}
	pr := NewPolicyRunner(mock,
		WithPolicies(AllowAllPolicy{}),
		WithPolicyHandler(handler),
	)

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(handler.results) != 1 {
		t.Fatalf("expected 1 handler call, got %d", len(handler.results))
	}

	if !handler.results[0].Allowed {
		t.Errorf("expected allowed=true in handler")
	}
}

type errorPolicy struct{}

func (errorPolicy) Name() string { return "error-policy" }

func (errorPolicy) Check(ctx context.Context, op *PolicyOperation) (*PolicyResult, error) {
	return nil, errors.New("policy check failed")
}

func TestPolicyRunnerCheckError(t *testing.T) {
	mock := NewMockRunner()
	pr := NewPolicyRunner(mock, WithPolicies(errorPolicy{}))

	_, err := pr.Run(context.Background(), "/tmp", "npm", "install")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var checkErr *ErrPolicyCheck
	if !errors.As(err, &checkErr) {
		t.Fatalf("expected ErrPolicyCheck, got %T", err)
	}

	if checkErr.Policy != "error-policy" {
		t.Errorf("got policy %q, want %q", checkErr.Policy, "error-policy")
	}
}

func TestPolicyModeString(t *testing.T) {
	tests := []struct {
		mode PolicyMode
		want string
	}{
		{PolicyEnforce, "enforce"},
		{PolicyWarn, "warn"},
		{PolicyDisabled, "disabled"},
		{PolicyMode(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("PolicyMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}
