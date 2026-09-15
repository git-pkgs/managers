package managers

import (
	"context"

	"github.com/git-pkgs/managers/definitions"
)

const argPackage = "package"

type GenericManager struct {
	def        *definitions.Definition
	dir        string
	translator *Translator
	runner     Runner
	warnings   []string
}

func (m *GenericManager) Name() string {
	return m.def.Name
}

func (m *GenericManager) Ecosystem() string {
	return m.def.Ecosystem
}

func (m *GenericManager) Dir() string {
	return m.dir
}

func (m *GenericManager) Warnings() []string {
	return m.warnings
}

// run builds the command chain for an operation (base command plus any
// then: entries) and executes them in order. It returns the first
// command's result with subsequent results attached in Then. Execution
// stops after the first non-zero exit; Success() on the returned result
// reflects the whole chain.
func (m *GenericManager) run(ctx context.Context, operation string, input CommandInput) (*Result, error) {
	cmds, err := m.translator.BuildCommands(m.def.Name, operation, input)
	if err != nil {
		return nil, err
	}
	var first *Result
	for i, cmd := range cmds {
		res, err := m.runner.Run(ctx, m.dir, cmd...)
		if i == 0 {
			first = res
		} else if first != nil {
			first.Then = append(first.Then, res)
		}
		if err != nil {
			return first, err
		}
		if res.ExitCode != 0 {
			return first, nil
		}
	}
	return first, nil
}

func (m *GenericManager) Init(ctx context.Context) (*Result, error) {
	return m.run(ctx, "init", CommandInput{})
}

func (m *GenericManager) Install(ctx context.Context, opts InstallOptions) (*Result, error) {
	return m.run(ctx, "install", CommandInput{
		Flags: map[string]any{
			"frozen":     opts.Frozen,
			"clean":      opts.Clean,
			"production": opts.Production,
		},
	})
}

func (m *GenericManager) Add(ctx context.Context, pkg string, opts AddOptions) (*Result, error) {
	input := CommandInput{
		Args: map[string]string{argPackage: pkg},
		Flags: map[string]any{
			"dev":       opts.Dev,
			"optional":  opts.Optional,
			"exact":     opts.Exact,
			"workspace": opts.Workspace,
		},
	}
	if opts.Version != "" {
		input.Args["version"] = opts.Version
	}
	return m.run(ctx, "add", input)
}

func (m *GenericManager) Remove(ctx context.Context, pkg string) (*Result, error) {
	return m.run(ctx, "remove", CommandInput{Args: map[string]string{argPackage: pkg}})
}

func (m *GenericManager) List(ctx context.Context) (*Result, error) {
	return m.run(ctx, "list", CommandInput{})
}

func (m *GenericManager) Outdated(ctx context.Context) (*Result, error) {
	return m.run(ctx, "outdated", CommandInput{})
}

func (m *GenericManager) Update(ctx context.Context, pkg string) (*Result, error) {
	input := CommandInput{}
	if pkg != "" {
		input.Args = map[string]string{argPackage: pkg}
	}
	return m.run(ctx, "update", input)
}

func (m *GenericManager) Supports(cap Capability) bool {
	capName := cap.String()
	for _, c := range m.def.Capabilities {
		if c == capName {
			return true
		}
	}
	return false
}

func (m *GenericManager) Capabilities() []Capability {
	var caps []Capability
	for _, name := range m.def.Capabilities {
		if cap, ok := CapabilityFromString(name); ok {
			caps = append(caps, cap)
		}
	}
	return caps
}

func (m *GenericManager) Vendor(ctx context.Context) (*Result, error) {
	return m.run(ctx, "vendor", CommandInput{})
}

func (m *GenericManager) Resolve(ctx context.Context) (*Result, error) {
	return m.run(ctx, "resolve", CommandInput{})
}

func (m *GenericManager) Path(ctx context.Context, pkg string) (*PathResult, error) {
	input := CommandInput{
		Args: map[string]string{
			argPackage: pkg,
		},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "path", input)
	if err != nil {
		return nil, err
	}

	result, err := m.runner.Run(ctx, m.dir, cmd...)
	if err != nil {
		return nil, err
	}

	var extract *definitions.Extract
	if pathCmd, ok := m.def.Commands["path"]; ok {
		extract = pathCmd.Extract
	}

	pathResult, err := extractPathResult(result.Stdout, extract, pkg)
	if err != nil {
		return &PathResult{Result: result}, err
	}
	pathResult.Result = result
	return pathResult, nil
}
