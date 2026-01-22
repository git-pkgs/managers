package definitions

type Definition struct {
	Name             string              `yaml:"name"`
	Ecosystem        string              `yaml:"ecosystem"`
	Binary           string              `yaml:"binary"`
	Version          string              `yaml:"version,omitempty"`
	Status           string              `yaml:"status,omitempty"`
	MinTested        string              `yaml:"min_tested,omitempty"`
	MaxTested        string              `yaml:"max_tested,omitempty"`
	Detection        Detection           `yaml:"detection"`
	VersionDetection VersionDetection    `yaml:"version_detection,omitempty"`
	Commands         map[string]Command  `yaml:"commands"`
	Capabilities     []string            `yaml:"capabilities"`
}

type Detection struct {
	Lockfiles  []string    `yaml:"lockfiles,omitempty"`
	Manifests  []string    `yaml:"manifests,omitempty"`
	Priority   int         `yaml:"priority"`
	FileChecks []FileCheck `yaml:"file_checks,omitempty"`
}

type FileCheck struct {
	File    string `yaml:"file"`
	Exists  bool   `yaml:"exists,omitempty"`
	Match   string `yaml:"match,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type VersionDetection struct {
	Command []string `yaml:"command,omitempty"`
	Pattern string   `yaml:"pattern,omitempty"`
}

type Command struct {
	Base          []string            `yaml:"base"`
	BaseOverrides map[string][]string `yaml:"base_overrides,omitempty"` // flag name -> replacement base
	Args          map[string]Arg      `yaml:"args,omitempty"`
	Flags         map[string]Flag     `yaml:"flags,omitempty"`
	DefaultFlags  []string            `yaml:"default_flags,omitempty"`
	ExitCodes     map[int]string      `yaml:"exit_codes,omitempty"`
	Then          []Command           `yaml:"then,omitempty"` // commands to run after this one
	Extract       *Extract            `yaml:"extract,omitempty"`
}

type Extract struct {
	Type          string `yaml:"type"`                     // raw, json, line_prefix, regex, json_array, template
	Field         string `yaml:"field,omitempty"`          // for json: field name to extract
	Prefix        string `yaml:"prefix,omitempty"`         // for line_prefix: prefix to match
	Pattern       string `yaml:"pattern,omitempty"`        // for regex: pattern with capture group; for template: path pattern with {package}
	ArrayField    string `yaml:"array_field,omitempty"`    // for json_array: array field to search
	MatchField    string `yaml:"match_field,omitempty"`    // for json_array: field to match against pkg name
	ExtractField  string `yaml:"extract_field,omitempty"`  // for json_array: field to extract from matched element
	StripFilename bool   `yaml:"strip_filename,omitempty"` // remove filename from path, returning directory
}

type Arg struct {
	Position       int    `yaml:"position"`
	Required       bool   `yaml:"required"`
	Validate       string `yaml:"validate,omitempty"`
	Flag           string `yaml:"flag,omitempty"`
	Suffix         string `yaml:"suffix,omitempty"`          // append user value with this prefix, e.g. "@" for pkg@version
	FixedSuffix    string `yaml:"fixed_suffix,omitempty"`    // always append this suffix, e.g. "@none" for go remove
	ExtractionOnly bool   `yaml:"extraction_only,omitempty"` // arg is only used for output extraction, not passed to command
}

type Flag struct {
	Values []FlagValue
}

type FlagValue struct {
	Literal string
	Field   string
	Join    string // if set, join literal and field value with this (e.g., "=" for --flag=value)
}

func (f *Flag) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw []interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	for _, v := range raw {
		switch val := v.(type) {
		case string:
			f.Values = append(f.Values, FlagValue{Literal: val})
		case map[string]interface{}:
			fv := FlagValue{}
			if field, ok := val["value"].(string); ok {
				fv.Field = field
			}
			if join, ok := val["join"].(string); ok {
				fv.Join = join
			}
			if fv.Field != "" {
				f.Values = append(f.Values, fv)
			}
		}
	}
	return nil
}

type Validator struct {
	Pattern   string `yaml:"pattern"`
	MaxLength int    `yaml:"max_length,omitempty"`
}
