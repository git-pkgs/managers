package managers

import (
	"context"

	"github.com/git-pkgs/managers/definitions"
)

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

func (m *GenericManager) Install(ctx context.Context, opts InstallOptions) (*Result, error) {
	input := CommandInput{
		Args: map[string]string{},
		Flags: map[string]any{
			"frozen":     opts.Frozen,
			"clean":      opts.Clean,
			"production": opts.Production,
		},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "install", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Add(ctx context.Context, pkg string, opts AddOptions) (*Result, error) {
	input := CommandInput{
		Args: map[string]string{
			"package": pkg,
		},
		Flags: map[string]any{
			"dev":       opts.Dev,
			"optional":  opts.Optional,
			"exact":     opts.Exact,
			"workspace": opts.Workspace,
		},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "add", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Remove(ctx context.Context, pkg string) (*Result, error) {
	input := CommandInput{
		Args: map[string]string{
			"package": pkg,
		},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "remove", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) List(ctx context.Context) (*Result, error) {
	input := CommandInput{
		Args:  map[string]string{},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "list", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Outdated(ctx context.Context) (*Result, error) {
	input := CommandInput{
		Args:  map[string]string{},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "outdated", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Update(ctx context.Context, pkg string) (*Result, error) {
	input := CommandInput{
		Args:  map[string]string{},
		Flags: map[string]any{},
	}

	if pkg != "" {
		input.Args["package"] = pkg
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "update", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
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
	input := CommandInput{
		Args:  map[string]string{},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "vendor", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Resolve(ctx context.Context) (*Result, error) {
	input := CommandInput{
		Args:  map[string]string{},
		Flags: map[string]any{},
	}

	cmd, err := m.translator.BuildCommand(m.def.Name, "resolve", input)
	if err != nil {
		return nil, err
	}

	return m.runner.Run(ctx, m.dir, cmd...)
}

func (m *GenericManager) Path(ctx context.Context, pkg string) (*PathResult, error) {
	input := CommandInput{
		Args: map[string]string{
			"package": pkg,
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

	path, err := ExtractPath(result.Stdout, extract, pkg)
	if err != nil {
		return &PathResult{Result: result}, err
	}

	return &PathResult{
		Path:   path,
		Result: result,
	}, nil
}
