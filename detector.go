package managers

import (
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/git-pkgs/managers/definitions"
)

type DetectOptions struct {
	RequireCLI    bool
	OnConflict    ConflictBehavior
	SearchParents bool
	Manager       string
}

type ConflictBehavior int

const (
	ConflictError ConflictBehavior = iota
	ConflictUseFirst
	ConflictUseNewest
)

type Detector struct {
	definitions []*definitions.Definition
	translator  *Translator
	runner      Runner
}

func NewDetector(translator *Translator, runner Runner) *Detector {
	return &Detector{
		translator: translator,
		runner:     runner,
	}
}

func (d *Detector) Register(def *definitions.Definition) {
	d.definitions = append(d.definitions, def)
	d.translator.Register(def)
	d.sortDefinitions()
}

func (d *Detector) sortDefinitions() {
	sort.Slice(d.definitions, func(i, j int) bool {
		return d.definitions[i].Detection.Priority > d.definitions[j].Detection.Priority
	})
}

func (d *Detector) Detect(dir string, opts DetectOptions) (Manager, error) {
	if opts.Manager != "" {
		return d.detectExplicit(dir, opts.Manager)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f.Name()] = true
	}

	var lockfileMatches []*definitions.Definition
	var lockfileNames []string

	for _, def := range d.definitions {
		for _, lockfile := range def.Detection.Lockfiles {
			if fileSet[lockfile] {
				lockfileMatches = append(lockfileMatches, def)
				lockfileNames = append(lockfileNames, lockfile)
			}
		}
	}

	if len(lockfileMatches) > 1 && opts.OnConflict == ConflictError {
		return nil, ErrConflictingLockfiles{
			Dir:       dir,
			Lockfiles: lockfileNames,
		}
	}

	if len(lockfileMatches) >= 1 {
		def := lockfileMatches[0]
		return d.buildManager(def, dir, lockfileNames[:1], opts.RequireCLI)
	}

	for _, def := range d.definitions {
		for _, manifest := range def.Detection.Manifests {
			if fileSet[manifest] {
				return d.buildManager(def, dir, []string{manifest}, opts.RequireCLI)
			}
		}
	}

	return nil, ErrNoManifest{Dir: dir}
}

func (d *Detector) detectExplicit(dir, managerName string) (Manager, error) {
	for _, def := range d.definitions {
		if def.Name == managerName {
			return d.buildManager(def, dir, nil, true)
		}
	}
	return nil, ErrNoManifest{Dir: dir}
}

func (d *Detector) buildManager(def *definitions.Definition, dir string, files []string, requireCLI bool) (Manager, error) {
	if requireCLI {
		if _, err := exec.LookPath(def.Binary); err != nil {
			return nil, ErrCLINotFound{
				Manager: def.Name,
				Binary:  def.Binary,
				Files:   files,
			}
		}
	}

	return &GenericManager{
		def:        def,
		dir:        dir,
		translator: d.translator,
		runner:     d.runner,
	}, nil
}

func (d *Detector) DetectVersion(def *definitions.Definition) (string, error) {
	if len(def.VersionDetection.Command) == 0 {
		return "", nil
	}

	binary := def.Binary
	args := def.VersionDetection.Command

	cmd := exec.Command(binary, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if def.VersionDetection.Pattern == "" {
		return strings.TrimSpace(string(output)), nil
	}

	re, err := regexp.Compile(def.VersionDetection.Pattern)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1], nil
	}

	return strings.TrimSpace(string(output)), nil
}
