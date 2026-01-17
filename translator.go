package managers

import (
	"fmt"

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

	// Check for base overrides (e.g., frozen flag changes "install" to "ci" for npm)
	baseOverrideUsed := ""
	base := cmd.Base
	for flagName, override := range cmd.BaseOverrides {
		if val, ok := input.Flags[flagName]; ok && isTruthy(val) {
			base = override
			baseOverrideUsed = flagName
			break
		}
	}
	args = append(args, base...)

	// Process args in a deterministic order
	// First handle package, then version (for suffix handling)
	packageVal := ""
	if val, ok := input.Args["package"]; ok {
		packageVal = val
	}

	for name, argDef := range cmd.Args {
		val, provided := input.Args[name]
		if !provided {
			if argDef.Required {
				return nil, ErrMissingArgument{Argument: name}
			}
			continue
		}

		if argDef.Validate != "" {
			if err := t.validate(argDef.Validate, val); err != nil {
				return nil, err
			}
		}

		if argDef.Flag != "" {
			// Flag-style arg: --version "1.0"
			args = append(args, argDef.Flag, val)
		} else if argDef.FixedSuffix != "" {
			// Fixed suffix: package@none
			args = append(args, val+argDef.FixedSuffix)
		} else if argDef.Suffix != "" && name == "version" {
			// Version suffix: find package arg and append @version
			// Skip here, handled below
			continue
		} else {
			args = append(args, val)
		}
	}

	// Handle version suffix (append to package)
	if versionDef, hasVersion := cmd.Args["version"]; hasVersion && versionDef.Suffix != "" {
		if version, hasVersionVal := input.Args["version"]; hasVersionVal {
			// Find and update the package arg
			for i, a := range args {
				if a == packageVal {
					args[i] = a + versionDef.Suffix + version
					break
				}
			}
		}
	}

	// Add default flags
	for _, flagStr := range cmd.DefaultFlags {
		args = append(args, flagStr)
	}

	// Add user-specified flags
	for name, val := range input.Flags {
		if val == false || val == "" || val == nil {
			continue
		}

		// Skip flag if it was used for base override
		if name == baseOverrideUsed {
			continue
		}

		flagDef, ok := cmd.Flags[name]
		if !ok {
			continue
		}

		expanded := t.expandFlag(flagDef, input.Flags)
		args = append(args, expanded...)
	}

	// Append any extra raw arguments (escape hatch for manager-specific flags)
	args = append(args, input.Extra...)

	return args, nil
}

func (t *Translator) expandFlag(flag definitions.Flag, flags map[string]any) []string {
	var result []string
	for _, v := range flag.Values {
		if v.Literal != "" && v.Field != "" && v.Join != "" {
			// Joined flag: --group=development
			if val, ok := flags[v.Field]; ok {
				if s, ok := val.(string); ok && s != "" {
					result = append(result, v.Literal+v.Join+s)
				}
			}
		} else if v.Literal != "" {
			result = append(result, v.Literal)
		} else if v.Field != "" {
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
		return nil
	}

	if v.MaxLength > 0 && len(value) > v.MaxLength {
		return ErrInvalidPackageName{
			Name:   value,
			Reason: fmt.Sprintf("exceeds maximum length of %d", v.MaxLength),
		}
	}

	return nil
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
