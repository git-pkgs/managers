package managers

import (
	"errors"
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
	Policy     *FileDetectionPolicy
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
	if opts.Policy != nil {
		detections, err := d.detectAllPolicyFS(fsys, dir, opts)
		if err != nil {
			return FileDetection{}, err
		}
		if len(detections) == 0 {
			return FileDetection{}, ErrNoManifest{Dir: dir}
		}
		return detections[0], nil
	}
	match, _, err := d.detectFS(fsys, dir, dir, opts)
	return match, err
}

// DetectAllFS detects one package manager per ecosystem from files in dir. It
// does not look for or execute package manager CLIs.
func (d *Detector) DetectAllFS(fsys fs.FS, dir string, opts FileDetectOptions) ([]FileDetection, error) {
	if opts.Policy != nil {
		return d.detectAllPolicyFS(fsys, dir, opts)
	}

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
		left, leftConfigured := d.ecosystemOrder(detections[i].Ecosystem)
		right, rightConfigured := d.ecosystemOrder(detections[j].Ecosystem)
		if leftConfigured != rightConfigured {
			return leftConfigured
		}
		if leftConfigured && left != right {
			return left < right
		}
		return detections[i].Ecosystem < detections[j].Ecosystem
	})
	return detections, nil
}

func (d *Detector) detectAllPolicyFS(
	fsys fs.FS,
	dir string,
	opts FileDetectOptions,
) ([]FileDetection, error) {
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

	var detections []FileDetection
	for _, ecosystem := range opts.Policy.Ecosystems {
		manifests, err := matcher.matchPatterns(ecosystem.Manifests)
		if err != nil {
			return nil, err
		}

		type lockfileMatch struct {
			manager  string
			lockfile string
		}

		var lockfiles []lockfileMatch
		for _, rule := range ecosystem.Lockfiles {
			matches, err := matcher.matchPatterns([]string{rule.File})
			if err != nil {
				return nil, err
			}
			for _, name := range matches {
				lockfiles = append(lockfiles, lockfileMatch{manager: rule.Manager, lockfile: name})
			}
		}

		if len(lockfiles) > 1 && opts.OnConflict == ConflictError {
			names := make([]string, 0, len(lockfiles))
			for _, match := range lockfiles {
				names = append(names, match.lockfile)
			}
			return nil, ErrConflictingLockfiles{Dir: dir, Lockfiles: names}
		}
		if len(lockfiles) > 0 {
			detections = append(detections, FileDetection{
				Manager:   lockfiles[0].manager,
				Ecosystem: ecosystem.Ecosystem,
				Directory: dir,
				Manifest:  first(manifests),
				Lockfile:  lockfiles[0].lockfile,
			})
			continue
		}

		if len(manifests) > 0 {
			detections = append(detections, FileDetection{
				Manager:   ecosystem.DefaultManager,
				Ecosystem: ecosystem.Ecosystem,
				Directory: dir,
				Manifest:  manifests[0],
			})
		}
	}

	return detections, nil
}

// Ecosystem returns the ecosystem registered for managerName.
func (d *Detector) Ecosystem(managerName string) (string, bool) {
	for _, def := range d.definitions {
		if def.Name == managerName {
			return def.Ecosystem, true
		}
	}
	return "", false
}

func (d *Detector) ecosystemOrder(ecosystem string) (int, bool) {
	order := 0
	configured := false
	for _, def := range d.definitions {
		if def.Ecosystem != ecosystem || def.Detection.EcosystemOrder <= 0 {
			continue
		}
		if !configured || def.Detection.EcosystemOrder < order {
			order = def.Detection.EcosystemOrder
			configured = true
		}
	}
	return order, configured
}

// AffectedDirectories returns project directories whose detection result may
// change when filePath changes.
func (d *Detector) AffectedDirectories(filePath string) ([]string, error) {
	directories := make(map[string]bool)
	for _, def := range d.definitions {
		patterns := append([]string(nil), def.Detection.Manifests...)
		patterns = append(patterns, def.Detection.Lockfiles...)
		for _, check := range def.Detection.FileChecks {
			patterns = append(patterns, check.File)
		}
		for _, pattern := range patterns {
			directory, matched, err := affectedDirectory(filePath, pattern)
			if err != nil {
				return nil, err
			}
			if matched {
				directories[directory] = true
			}
		}
	}

	result := make([]string, 0, len(directories))
	for directory := range directories {
		result = append(result, directory)
	}
	if len(result) == 0 {
		return nil, nil
	}
	sort.Strings(result)
	return result, nil
}

func affectedDirectory(filePath, pattern string) (string, bool, error) {
	filePath = path.Clean(filePath)
	for offset := 0; offset < len(filePath); {
		suffix := filePath[offset:]
		matched, err := path.Match(pattern, suffix)
		if err != nil {
			return "", false, err
		}
		if matched {
			directory := strings.TrimSuffix(filePath[:offset], "/")
			if directory == "" {
				directory = "."
			}
			return directory, true, nil
		}

		next := strings.IndexByte(suffix, '/')
		if next < 0 {
			break
		}
		offset += next + 1
	}
	return "", false, nil
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

	if matched, ok := selectManifestDefinition(matchedDefinitions); ok {
		return newFileDetection(
			matched.def,
			reportedDir,
			matched.matches.manifests[0],
			"",
		), matched.def, nil
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
		if len(matches.manifests) > 0 || len(matches.lockfiles) > 0 {
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

	if match, ok := selectManifestDefinition(matches); ok {
		return newFileDetection(match.def, dir, match.matches.manifests[0], ""), nil
	}

	return FileDetection{}, ErrNoManifest{Dir: dir}
}

func selectManifestDefinition(matches []matchedDefinition) (matchedDefinition, bool) {
	for _, match := range matches {
		if match.matches.checksMatch && hasPositiveFileCheck(match.def) && len(match.matches.manifests) > 0 {
			return match, true
		}
	}

	for _, match := range matches {
		if match.def.Detection.Default && len(match.matches.manifests) > 0 {
			return match, true
		}
	}

	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if match.matches.checksMatch && len(match.matches.manifests) > 0 {
			return match, true
		}
	}

	return matchedDefinition{}, false
}

func hasPositiveFileCheck(def *definitions.Definition) bool {
	for _, check := range def.Detection.FileChecks {
		if check.Match != "" || check.Exists {
			return true
		}
	}
	return false
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
	checksMatch bool
	manifests   []string
	lockfiles   []string
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
	checksMatch, err := m.matchFileChecks(def.Detection.FileChecks)
	if err != nil {
		return definitionMatch{}, err
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
		checksMatch: checksMatch,
		manifests:   manifests,
		lockfiles:   lockfiles,
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
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
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
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
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
