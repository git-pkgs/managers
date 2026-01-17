package managers

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type Runner interface {
	Run(ctx context.Context, dir string, args ...string) (*Result, error)
}

type ExecRunner struct{}

func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

func (r *ExecRunner) Run(ctx context.Context, dir string, args ...string) (*Result, error) {
	if len(args) == 0 {
		return nil, ErrNoCommand
	}

	start := time.Now()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Command:  args,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
		Cwd:      dir,
		Context:  ContextProject,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else {
		result.ExitCode = -1
	}

	if err != nil && result.ExitCode == -1 {
		return result, err
	}

	return result, nil
}

type MockRunner struct {
	Captured [][]string
	Results  []*Result
	Errors   []error
	callIdx  int
}

func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

func (m *MockRunner) Run(ctx context.Context, dir string, args ...string) (*Result, error) {
	m.Captured = append(m.Captured, args)

	idx := m.callIdx
	m.callIdx++

	if idx < len(m.Errors) && m.Errors[idx] != nil {
		return nil, m.Errors[idx]
	}

	if idx < len(m.Results) {
		return m.Results[idx], nil
	}

	return &Result{
		Command:  args,
		ExitCode: 0,
		Cwd:      dir,
	}, nil
}

func (m *MockRunner) LastCaptured() []string {
	if len(m.Captured) == 0 {
		return nil
	}
	return m.Captured[len(m.Captured)-1]
}
