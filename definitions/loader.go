package definitions

import (
	"embed"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var definitionFiles embed.FS

func LoadEmbedded() ([]*Definition, error) {
	entries, err := definitionFiles.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var defs []*Definition
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		data, err := definitionFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}

		var def Definition
		if err := yaml.Unmarshal(data, &def); err != nil {
			return nil, err
		}

		defs = append(defs, &def)
	}

	return defs, nil
}

func LoadFromBytes(data []byte) (*Definition, error) {
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, err
	}
	return &def, nil
}
