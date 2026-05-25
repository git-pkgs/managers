package managers

import (
	"fmt"
	"sort"

	"github.com/git-pkgs/managers/definitions"
)

type Translator struct {
	definitions map[string]*definitions.Definition
	validators  map[string]*definitions.Validator
}

func NewTranslator() *Translator {
	return &Translator{
		definitions: make(map[string]*definitions.Definition),
		validators:  make(map[string]*definitions.Validator),
	}
}

func (t *Translator) Register(def *definitions.Definition) {
	t.definitions[def.Name] = def
}

func (t *Translator) RegisterValidator(name string, v *definitions.Validator) {
	t.validators[name] = v
}

func (t *Translator) Definition(name string) (*definitions.Definition, bool) {
	def, ok := t.definitions[name]
	return def, ok
}

type CommandInput struct {
	Args  map[string]string
	Flags map[string]any
	Extra []string // Raw arguments appended to the command (escape hatch)
}

func (t *Translator) BuildCommand(managerName, operation string, input CommandInput) ([]string, error) {
	def, ok := t.definitions[managerName]
	if !ok {
		return nil, fmt.Errorf("unknown manager: %s", managerName)
	}

	cmd, ok := def.Commands[operation]
	if !ok {
		return nil, ErrUnsupportedOperation
	}

	return t.buildSingleCommand(def.Binary, cmd, input)
}

// BuildCommands returns all commands for an operation (including "then" chains)
func (t *Translator) BuildCommands(managerName, operation string, input CommandInput) ([][]string, error) {
	def, ok := t.definitions[managerName]
	if !ok {
		return nil, fmt.Errorf("unknown manager: %s", managerName)
	}

	cmd, ok := def.Commands[operation]
	if !ok {
		return nil, ErrUnsupportedOperation
	}

	return t.buildCommandChain(def.Binary, cmd, input)
}

func (t *Translator) buildCommandChain(binary string, cmd definitions.Command, input CommandInput) ([][]string, error) {
	first, err := t.buildSingleCommand(binary, cmd, input)
	if err != nil {
		return nil, err
	}

	result := [][]string{first}

	for _, next := range cmd.Then {
		nextCmd, err := t.buildSingleCommand(binary, next, input)
		if err != nil {
			return nil, err
		}
		result = append(result, nextCmd)
	}

	return result, nil
}

func (t *Translator) buildSingleCommand(binary string, cmd definitions.Command, input CommandInput) ([]string, error) {
	args := []string{binary}

	suppressDefaultFlags := false

	baseOverrideUsed := t.applyBaseOverrides(&args, cmd, input)

	packageVal := input.Args["package"]

	sortedArgs := t.sortArgs(cmd)

	for _, entry := range sortedArgs {
		val, err := t.processArg(entry.name, entry.argDef, input, &args)
		if err != nil {
			return nil, err
		}
		if val == "" {
			continue
		}
		if entry.argDef.SuppressDefaultFlags {
			suppressDefaultFlags = true
		}
	}

	t.applyVersionSuffix(&args, cmd, input, packageVal)

	if !suppressDefaultFlags {
		args = append(args, cmd.DefaultFlags...)
	}

	t.applyUserFlags(&args, cmd, input, baseOverrideUsed)

	args = append(args, input.Extra...)

	return args, nil
}

type argEntry struct {
	name   string
	argDef definitions.Arg
}

func (t *Translator) applyBaseOverrides(args *[]string, cmd definitions.Command, input CommandInput) string {
	base := cmd.Base
	baseOverrideUsed := ""
	for flagName, override := range cmd.BaseOverrides {
		if val, ok := input.Flags[flagName]; ok && isTruthy(val) {
			base = override
			baseOverrideUsed = flagName
			break
		}
	}
	*args = append(*args, base...)
	return baseOverrideUsed
}

func (t *Translator) sortArgs(cmd definitions.Command) []argEntry {
	var sorted []argEntry
	for name, argDef := range cmd.Args {
		sorted = append(sorted, argEntry{name, argDef})
	}
	sort.Slice(sorted, func(i, j int) bool {
		iIsFlag := sorted[i].argDef.Flag != ""
		jIsFlag := sorted[j].argDef.Flag != ""
		if iIsFlag != jIsFlag {
			return !iIsFlag
		}
		return sorted[i].argDef.Position < sorted[j].argDef.Position
	})
	return sorted
}

func (t *Translator) processArg(name string, argDef definitions.Arg, input CommandInput, args *[]string) (string, error) {
	val, provided := input.Args[name]
	if !provided {
		if argDef.Required && !argDef.ExtractionOnly {
			return "", ErrMissingArgument{Argument: name}
		}
		return "", nil
	}

	if argDef.ExtractionOnly {
		return "", nil
	}

	if argDef.Validate != "" {
		if err := t.validate(argDef.Validate, val); err != nil {
			return "", err
		}
	}

	switch {
	case argDef.Flag != "":
		*args = append(*args, argDef.Flag, val)
	case argDef.FixedSuffix != "":
		*args = append(*args, val+argDef.FixedSuffix)
	case argDef.Suffix != "" && name == "version":
		// Handled in applyVersionSuffix
	default:
		*args = append(*args, val)
	}

	return val, nil
}

func (t *Translator) applyVersionSuffix(args *[]string, cmd definitions.Command, input CommandInput, packageVal string) {
	versionDef, hasVersion := cmd.Args["version"]
	if !hasVersion || versionDef.Suffix == "" {
		return
	}
	version, hasVersionVal := input.Args["version"]
	if !hasVersionVal {
		return
	}
	for i, a := range *args {
		if a == packageVal {
			(*args)[i] = a + versionDef.Suffix + version
			break
		}
	}
}

func (t *Translator) applyUserFlags(args *[]string, cmd definitions.Command, input CommandInput, baseOverrideUsed string) {
	for name, val := range input.Flags {
		if val == false || val == "" || val == nil {
			continue
		}
		if name == baseOverrideUsed {
			continue
		}
		flagDef, ok := cmd.Flags[name]
		if !ok {
			continue
		}
		expanded := t.expandFlag(flagDef, input.Flags)
		*args = append(*args, expanded...)
	}
}

func (t *Translator) expandFlag(flag definitions.Flag, flags map[string]any) []string {
	var result []string
	for _, v := range flag.Values {
		switch {
		case v.Literal != "" && v.Field != "" && v.Join != "":
			if val, ok := flags[v.Field]; ok {
				if s, ok := val.(string); ok && s != "" {
					result = append(result, v.Literal+v.Join+s)
				}
			}
		case v.Literal != "":
			result = append(result, v.Literal)
		case v.Field != "":
			if val, ok := flags[v.Field]; ok {
				if s, ok := val.(string); ok && s != "" {
					result = append(result, s)
				}
			}
		}
	}
	return result
}

func (t *Translator) validate(validatorName, value string) error {
	v, ok := t.validators[validatorName]
	if !ok {
		return ValidatePackageName(validatorName, value)
	}

	if v.MaxLength > 0 && len(value) > v.MaxLength {
		return ErrInvalidPackageName{
			Name:   value,
			Reason: fmt.Sprintf("exceeds maximum length of %d", v.MaxLength),
		}
	}

	return nil
}

// ExitCodeMeaning returns the semantic meaning of an exit code for a
// manager/operation pair, as defined in the YAML. Returns "" if the
// manager, operation, or exit code is not defined.
func (t *Translator) ExitCodeMeaning(managerName, operation string, exitCode int) string {
	def, ok := t.definitions[managerName]
	if !ok {
		return ""
	}
	cmd, ok := def.Commands[operation]
	if !ok {
		return ""
	}
	return cmd.ExitCodes[exitCode]
}

// IsFatalExitCode reports whether exitCode represents a fatal error for
// the given manager/operation. Exit code 0 is never fatal. For non-zero
// codes, the result is fatal unless the YAML definition assigns a
// non-"error" meaning.
func (t *Translator) IsFatalExitCode(managerName, operation string, exitCode int) bool {
	if exitCode == 0 {
		return false
	}
	meaning := t.ExitCodeMeaning(managerName, operation, exitCode)
	if meaning == "" || meaning == "error" {
		return true
	}
	return false
}

func isTruthy(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	default:
		return true
	}
}
