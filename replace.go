package managers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const opReplace = "replace"

type replaceMode int

const (
	replaceModeNone replaceMode = iota
	replaceModePath
	replaceModeGit
	replaceModeDrop
)

var goModuleVersionRE = regexp.MustCompile(`^v\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)

// manifestReplaceFiles maps managers whose replace operation is a direct
// manifest edit (no CLI subprocess) to the manifest filename relative to
// the project directory.
var manifestReplaceFiles = map[string]string{
	"cargo":   "Cargo.toml",
	"uv":      "pyproject.toml",
	"bundler": "Gemfile",
}

func (o ReplaceOptions) mode() (replaceMode, error) {
	modes := 0
	var mode replaceMode
	if o.Path != "" {
		mode = replaceModePath
		modes++
	}
	if o.Git != "" {
		mode = replaceModeGit
		modes++
	}
	if o.Drop {
		mode = replaceModeDrop
		modes++
	}
	if modes == 0 {
		return replaceModeNone, errors.New("specify one of Path, Git, or Drop")
	}
	if modes > 1 {
		return replaceModeNone, errors.New("specify only one of Path, Git, or Drop")
	}
	if o.Ref != "" && mode != replaceModeGit {
		return replaceModeNone, errors.New("ref can only be used with a git replacement")
	}
	return mode, nil
}

func replaceCapability(mode replaceMode) Capability {
	switch mode {
	case replaceModePath:
		return CapReplacePath
	case replaceModeGit:
		return CapReplaceGit
	case replaceModeDrop:
		return CapReplaceDrop
	default:
		return CapReplacePath
	}
}

// Replace redirects pkg to a local path or git source, or removes a prior
// redirect with Drop. For most managers this runs a CLI command; cargo, uv
// and bundler edit the manifest file directly since those ecosystems have
// no CLI for source overrides.
func (m *GenericManager) Replace(ctx context.Context, pkg string, opts ReplaceOptions) (*Result, error) {
	mode, err := opts.mode()
	if err != nil {
		return nil, err
	}
	if !m.Supports(replaceCapability(mode)) {
		return nil, ErrUnsupportedOperation
	}

	if file, ok := manifestReplaceFiles[m.def.Name]; ok {
		return m.replaceInManifest(file, pkg, mode, opts)
	}

	inputs, err := m.buildReplaceInputs(pkg, mode, opts)
	if err != nil {
		return nil, err
	}

	var result *Result
	for _, in := range inputs {
		cmd, err := m.translator.BuildCommand(m.def.Name, in.operation, in.input)
		if err != nil {
			return nil, err
		}
		result, err = m.runner.Run(ctx, m.dir, cmd...)
		if err != nil {
			return result, err
		}
		if result.ExitCode != 0 {
			return result, nil
		}
	}
	return result, nil
}

type replaceInput struct {
	operation string
	input     CommandInput
}

func (m *GenericManager) buildReplaceInputs(pkg string, mode replaceMode, opts ReplaceOptions) ([]replaceInput, error) {
	switch m.def.Name {
	case "gomod":
		return buildGoReplaceInputs(pkg, mode, opts)
	case "npm", "pnpm", "yarn", "bun":
		return buildNPMReplaceInputs(m.def.Name, pkg, mode, opts)
	case "composer":
		return buildComposerReplaceInputs(pkg, mode, opts)
	default:
		return nil, ErrUnsupportedOperation
	}
}

func buildGoReplaceInputs(pkg string, mode replaceMode, opts ReplaceOptions) ([]replaceInput, error) {
	input := CommandInput{Args: map[string]string{}, Flags: map[string]any{}, Extra: opts.Extra}
	switch mode {
	case replaceModePath:
		input.Args["spec"] = pkg + "=" + opts.Path
	case replaceModeGit:
		target, err := normalizeGoReplaceTarget(opts.Git)
		if err != nil {
			return nil, err
		}
		if opts.Ref == "" {
			return nil, errors.New("gomod replace to a git module requires Ref (a Go module version or pseudo-version)")
		}
		if !goModuleVersionRE.MatchString(opts.Ref) {
			return nil, fmt.Errorf("gomod Ref must be a Go module version or pseudo-version, got %q", opts.Ref)
		}
		input.Args["spec"] = pkg + "=" + target + "@" + opts.Ref
	case replaceModeDrop:
		input.Args[argPackage] = pkg
		input.Flags["drop"] = true
	}
	// go mod edit is a text edit; tidy afterwards so go.sum reflects the
	// replacement's transitive requirements. Matches Add, which chains
	// tidy via then: in gomod.yaml.
	return []replaceInput{
		{operation: opReplace, input: input},
		{operation: "tidy", input: CommandInput{Args: map[string]string{}, Flags: map[string]any{}}},
	}, nil
}

func buildNPMReplaceInputs(manager, pkg string, mode replaceMode, opts ReplaceOptions) ([]replaceInput, error) {
	input := CommandInput{Args: map[string]string{}, Flags: map[string]any{}, Extra: opts.Extra}
	switch mode {
	case replaceModePath:
		input.Args["spec"] = pkg + "@file:" + opts.Path
	case replaceModeGit:
		input.Args["spec"] = pkg + "@" + npmGitSpec(opts.Git, opts.Ref)
	case replaceModeDrop:
		return nil, fmt.Errorf("%s cannot drop a file/git replacement without the original version; use Add with a registry version instead", manager)
	}
	return []replaceInput{{operation: opReplace, input: input}}, nil
}

func buildComposerReplaceInputs(pkg string, mode replaceMode, opts ReplaceOptions) ([]replaceInput, error) {
	repoKey := "repositories." + composerRepositoryName(pkg)
	input := CommandInput{
		Args:  map[string]string{"repository": repoKey},
		Flags: map[string]any{},
	}
	switch mode {
	case replaceModePath:
		input.Args["payload"] = fmt.Sprintf(`{"type":"path","url":%q}`, opts.Path)
	case replaceModeGit:
		input.Args["payload"] = fmt.Sprintf(`{"type":"vcs","url":%q}`, opts.Git)
		inputs := []replaceInput{{operation: opReplace, input: input}}
		if opts.Ref != "" {
			inputs = append(inputs, replaceInput{
				operation: "add",
				input: CommandInput{
					Args:  map[string]string{argPackage: pkg + ":dev-" + opts.Ref},
					Flags: map[string]any{},
					Extra: opts.Extra,
				},
			})
		}
		return inputs, nil
	case replaceModeDrop:
		input.Flags["drop"] = true
	}
	input.Extra = opts.Extra
	return []replaceInput{{operation: opReplace, input: input}}, nil
}

func (m *GenericManager) replaceInManifest(file, pkg string, mode replaceMode, opts ReplaceOptions) (*Result, error) {
	start := time.Now()
	path := filepath.Join(m.dir, file)
	var err error
	switch m.def.Name {
	case "cargo":
		err = editTOMLReplaceEntry(path, "[patch.crates-io]", pkg, mode, opts)
	case "uv":
		err = editTOMLReplaceEntry(path, "[tool.uv.sources]", pkg, mode, opts)
	case "bundler":
		err = editGemfileReplaceEntry(path, pkg, mode, opts)
	default:
		err = ErrUnsupportedOperation
	}
	if err != nil {
		return nil, err
	}
	return &Result{
		ExitCode: 0,
		Stdout:   "edited " + path,
		Duration: time.Since(start),
		Cwd:      m.dir,
		Context:  ContextProject,
	}, nil
}

func npmGitSpec(repo, ref string) string {
	spec := repo
	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		spec = "git+" + spec
	}
	if ref != "" {
		spec += "#" + ref
	}
	return spec
}

// normalizeGoReplaceTarget converts a git URL into the module-path form
// that go.mod replace directives accept, e.g.
// https://github.com/foo/bar.git -> github.com/foo/bar.
func normalizeGoReplaceTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", errors.New("go replacement target cannot be empty")
	}

	if strings.HasPrefix(target, "git@") {
		withoutPrefix := strings.TrimPrefix(target, "git@")
		host, path, ok := strings.Cut(withoutPrefix, ":")
		if !ok || host == "" || path == "" {
			return "", fmt.Errorf("invalid go replacement git target %q", target)
		}
		return strings.TrimSuffix(host+"/"+strings.Trim(path, "/"), ".git"), nil
	}

	if strings.Contains(target, "://") {
		parsed, err := url.Parse(target)
		if err != nil || parsed.Host == "" || parsed.Path == "" {
			return "", fmt.Errorf("invalid go replacement git target %q", target)
		}
		switch parsed.Scheme {
		case "https", "http", "ssh", "git":
		default:
			return "", fmt.Errorf("unsupported go replacement git URL scheme %q", parsed.Scheme)
		}
		return strings.TrimSuffix(parsed.Host+"/"+strings.Trim(parsed.Path, "/"), ".git"), nil
	}

	if strings.HasPrefix(target, "git+") {
		return "", fmt.Errorf("unsupported go replacement git target %q; use a module path or URL", target)
	}
	return strings.TrimSuffix(target, ".git"), nil
}

func composerRepositoryName(pkg string) string {
	r := strings.NewReplacer("/", "-", "_", "-", ".", "-")
	return "git-pkgs-" + r.Replace(pkg)
}
