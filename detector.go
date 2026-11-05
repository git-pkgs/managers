package managers

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
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

// FileDetectOptions controls package manager detection without command execution.
type FileDetectOptions struct {
	OnConflict ConflictBehavior
	Manager    string
}

// FileDetection describes the files that identify a package manager.
type FileDetection struct {
	Manager   string
	Ecosystem string
	Directory string
	Manifest  string
	Lockfile  string
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
	sort.SliceStable(d.definitions, func(i, j int) bool {
		return d.definitions[i].Detection.Priority > d.definitions[j].Detection.Priority
	})
}

func (d *Detector) Detect(dir string, opts DetectOptions) (Manager, error) { //nolint:ireturn
	if opts.Manager != "" {
		return d.detectExplicit(dir, opts.Manager, opts.RequireCLI)
	}

	match, def, err := d.detectFS(os.DirFS(dir), ".", dir, FileDetectOptions{
		OnConflict: opts.OnConflict,
	})
	if err != nil {
		return nil, err
	}

	files := []string{match.Lockfile}
	if match.Lockfile == "" {
		files = []string{match.Manifest}
	}

	return d.buildManager(def, dir, files, opts.RequireCLI)
}

// DetectFS detects a package manager from files in dir. It does not look for or
// execute the package manager CLI.
func (d *Detector) DetectFS(fsys fs.FS, dir string, opts FileDetectOptions) (FileDetection, error) {
	match, _, err := d.detectFS(fsys, dir, dir, opts)
	return match, err
}

// DetectAllFS detects one package manager per ecosystem from files in dir. It
// does not look for or execute package manager CLIs.
func (d *Detector) DetectAllFS(fsys fs.FS, dir string, opts FileDetectOptions) ([]FileDetection, error) {
	matcher, err := newFileMatcher(fsys, dir)
	if err != nil {
		return nil, err
	}

	if opts.Manager != "" {
		match, _, err := d.detectExplicitFS(matcher, dir, opts.Manager)
		if err != nil {
			return nil, err
		}
		return []FileDetection{match}, nil
	}

	matches, err := d.matchDefinitions(matcher)
	if err != nil {
		return nil, err
	}

	byEcosystem := make(map[string][]matchedDefinition)
	for _, match := range matches {
		if len(match.matches.manifests) == 0 && len(match.matches.lockfiles) == 0 {
			continue
		}
		byEcosystem[match.def.Ecosystem] = append(byEcosystem[match.def.Ecosystem], match)
	}

	detections := make([]FileDetection, 0, len(byEcosystem))
	for _, ecosystemMatches := range byEcosystem {
		detection, err := selectEcosystemDetection(dir, ecosystemMatches, opts.OnConflict)
		if err != nil {
			return nil, err
		}
		detections = append(detections, detection)
	}

	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Ecosystem < detections[j].Ecosystem
	})
	return detections, nil
}

type matchedDefinition struct {
	def     *definitions.Definition
	matches definitionMatch
}

func (d *Detector) detectFS(
	fsys fs.FS,
	dir string,
	reportedDir string,
	opts FileDetectOptions,
) (FileDetection, *definitions.Definition, error) {
	matcher, err := newFileMatcher(fsys, dir)
	if err != nil {
		return FileDetection{}, nil, err
	}

	if opts.Manager != "" {
		return d.detectExplicitFS(matcher, reportedDir, opts.Manager)
	}

	type lockfileMatch struct {
		def      *definitions.Definition
		lockfile string
		manifest string
	}

	matchedDefinitions, err := d.matchDefinitions(matcher)
	if err != nil {
		return FileDetection{}, nil, err
	}

	var lockfileMatches []lockfileMatch
	for _, matched := range matchedDefinitions {
		def := matched.def
		matches := matched.matches
		for _, lockfile := range matches.lockfiles {
			lockfileMatches = append(lockfileMatches, lockfileMatch{
				def:      def,
				lockfile: lockfile,
				manifest: first(matches.manifests),
			})
		}
	}

	if len(lockfileMatches) > 1 && opts.OnConflict == ConflictError {
		lockfiles := make([]string, 0, len(lockfileMatches))
		for _, match := range lockfileMatches {
			lockfiles = append(lockfiles, match.lockfile)
		}
		return FileDetection{}, nil, ErrConflictingLockfiles{
			Dir:       reportedDir,
			Lockfiles: lockfiles,
		}
	}

	if len(lockfileMatches) > 0 {
		match := lockfileMatches[0]
		return newFileDetection(match.def, reportedDir, match.manifest, match.lockfile), match.def, nil
	}

	for i := len(matchedDefinitions) - 1; i >= 0; i-- {
		matched := matchedDefinitions[i]
		def := matched.def
		matches := matched.matches
		if matches.valid && len(matches.manifests) > 0 {
			return newFileDetection(def, reportedDir, matches.manifests[0], ""), def, nil
		}
	}

	return FileDetection{}, nil, ErrNoManifest{Dir: reportedDir}
}

func (d *Detector) matchDefinitions(matcher *fileMatcher) ([]matchedDefinition, error) {
	var matched []matchedDefinition
	for _, def := range d.definitions {
		matches, err := matcher.matchDefinition(def)
		if err != nil {
			return nil, err
		}
		if matches.valid {
			matched = append(matched, matchedDefinition{def: def, matches: matches})
		}
	}
	return matched, nil
}

func selectEcosystemDetection(
	dir string,
	matches []matchedDefinition,
	onConflict ConflictBehavior,
) (FileDetection, error) {
	type lockfileMatch struct {
		definition matchedDefinition
		lockfile   string
	}

	var lockfiles []lockfileMatch
	for _, match := range matches {
		for _, lockfile := range match.matches.lockfiles {
			lockfiles = append(lockfiles, lockfileMatch{definition: match, lockfile: lockfile})
		}
	}

	if len(lockfiles) > 1 && onConflict == ConflictError {
		names := make([]string, 0, len(lockfiles))
		for _, match := range lockfiles {
			names = append(names, match.lockfile)
		}
		return FileDetection{}, ErrConflictingLockfiles{Dir: dir, Lockfiles: names}
	}
	if len(lockfiles) > 0 {
		match := lockfiles[0]
		return newFileDetection(
			match.definition.def,
			dir,
			first(match.definition.matches.manifests),
			match.lockfile,
		), nil
	}

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if len(match.matches.manifests) > 0 {
			return newFileDetection(match.def, dir, match.matches.manifests[0], ""), nil
		}
	}

	return FileDetection{}, ErrNoManifest{Dir: dir}
}

func (d *Detector) detectExplicitFS(
	matcher *fileMatcher,
	reportedDir string,
	managerName string,
) (FileDetection, *definitions.Definition, error) {
	for _, def := range d.definitions {
		if def.Name != managerName {
			continue
		}

		matches, err := matcher.matchDefinition(def)
		if err != nil {
			return FileDetection{}, nil, err
		}
		return newFileDetection(
			def,
			reportedDir,
			first(matches.manifests),
			first(matches.lockfiles),
		), def, nil
	}

	return FileDetection{}, nil, ErrNoManifest{Dir: reportedDir}
}

type definitionMatch struct {
	valid     bool
	manifests []string
	lockfiles []string
}

type fileMatcher struct {
	fsys  fs.FS
	dir   string
	files []string
}

func newFileMatcher(fsys fs.FS, dir string) (*fileMatcher, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return &fileMatcher{fsys: fsys, dir: dir, files: files}, nil
}

func (m *fileMatcher) matchDefinition(def *definitions.Definition) (definitionMatch, error) {
	valid, err := m.matchFileChecks(def.Detection.FileChecks)
	if err != nil || !valid {
		return definitionMatch{valid: valid}, err
	}

	manifests, err := m.matchPatterns(def.Detection.Manifests)
	if err != nil {
		return definitionMatch{}, err
	}
	lockfiles, err := m.matchPatterns(def.Detection.Lockfiles)
	if err != nil {
		return definitionMatch{}, err
	}

	return definitionMatch{
		valid:     true,
		manifests: manifests,
		lockfiles: lockfiles,
	}, nil
}

func (m *fileMatcher) matchFileChecks(checks []definitions.FileCheck) (bool, error) {
	for _, check := range checks {
		matches, err := m.matchPatterns([]string{check.File})
		if err != nil {
			return false, err
		}

		if check.Match != "" {
			matched, err := m.matchFileContents(matches, check.Match)
			if err != nil || !matched {
				return matched, err
			}
			continue
		}

		if check.Exists != (len(matches) > 0) {
			return false, nil
		}
	}

	return true, nil
}

func (m *fileMatcher) matchFileContents(files []string, pattern string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	for _, name := range files {
		contents, err := fs.ReadFile(m.fsys, joinFSPath(m.dir, name))
		if err != nil {
			return false, err
		}
		if re.Match(contents) {
			return true, nil
		}
	}

	return false, nil
}

func (m *fileMatcher) matchPatterns(patterns []string) ([]string, error) {
	var matches []string
	for _, pattern := range patterns {
		if !strings.Contains(pattern, "/") {
			for _, name := range m.files {
				matched, err := path.Match(pattern, name)
				if err != nil {
					return nil, err
				}
				if matched {
					matches = append(matches, name)
				}
			}
			continue
		}

		found, err := fs.Glob(m.fsys, joinFSPath(m.dir, pattern))
		if err != nil {
			return nil, err
		}
		for _, name := range found {
			info, err := fs.Stat(m.fsys, name)
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				matches = append(matches, trimFSDir(m.dir, name))
			}
		}
	}
	return matches, nil
}

func joinFSPath(dir, name string) string {
	if dir == "." {
		return name
	}
	return path.Join(dir, name)
}

func trimFSDir(dir, name string) string {
	if dir == "." {
		return name
	}
	return strings.TrimPrefix(name, path.Clean(dir)+"/")
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func newFileDetection(
	def *definitions.Definition,
	dir string,
	manifest string,
	lockfile string,
) FileDetection {
	return FileDetection{
		Manager:   def.Name,
		Ecosystem: def.Ecosystem,
		Directory: dir,
		Manifest:  manifest,
		Lockfile:  lockfile,
	}
}

func (d *Detector) detectExplicit(dir, managerName string, requireCLI bool) (Manager, error) { //nolint:ireturn
	for _, def := range d.definitions {
		if def.Name == managerName {
			return d.buildManager(def, dir, nil, requireCLI)
		}
	}
	return nil, ErrNoManifest{Dir: dir}
}

func (d *Detector) buildManager(def *definitions.Definition, dir string, files []string, requireCLI bool) (Manager, error) { //nolint:ireturn
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
