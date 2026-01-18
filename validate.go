package managers

import (
	"regexp"
)

var defaultValidators = map[string]*regexp.Regexp{
	"package_name":   regexp.MustCompile(`^[@a-zA-Z0-9][\w\-\./]*$`),
	"npm_package":    regexp.MustCompile(`^(@[a-z0-9-~][a-z0-9-._~]*/)?[a-z0-9-~][a-z0-9-._~]*$`),
	"gem_name":       regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
	"cargo_crate":    regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
	"go_module":      regexp.MustCompile(`^[a-zA-Z0-9][\w\-\.\/]*$`),
	"maven_artifact": regexp.MustCompile(`^[a-zA-Z0-9._-]+:[a-zA-Z0-9._-]+$`),
}

var maxLengths = map[string]int{
	"package_name": 214,
	"npm_package":  214,
	"gem_name":     128,
	"cargo_crate":  64,
	"go_module":    256,
}

func ValidatePackageName(validatorName, name string) error {
	if name == "" {
		return ErrInvalidPackageName{Name: name, Reason: "empty name"}
	}

	if maxLen, ok := maxLengths[validatorName]; ok {
		if len(name) > maxLen {
			return ErrInvalidPackageName{
				Name:   name,
				Reason: "exceeds maximum length",
			}
		}
	}

	pattern, ok := defaultValidators[validatorName]
	if !ok {
		pattern = defaultValidators["package_name"]
	}

	if !pattern.MatchString(name) {
		return ErrInvalidPackageName{
			Name:   name,
			Reason: "contains invalid characters",
		}
	}

	return nil
}
