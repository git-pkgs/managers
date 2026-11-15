package managers

const (
	apmManagerName   = "apm"
	npmManagerName   = "npm"
	nugetManagerName = "nuget"
)

// FileDetectionPolicy defines an ordered set of file-based detection rules.
type FileDetectionPolicy struct {
	Ecosystems []EcosystemDetectionPolicy
}

// EcosystemDetectionPolicy selects one manager for an ecosystem.
type EcosystemDetectionPolicy struct {
	Ecosystem      string
	DefaultManager string
	Manifests      []string
	Lockfiles      []LockfileDetectionPolicy
}

// LockfileDetectionPolicy associates a lockfile with a manager.
type LockfileDetectionPolicy struct {
	File    string
	Manager string
}

// GitPkgsFileDetectionPolicy returns the file rules used by the git-pkgs CLI.
func GitPkgsFileDetectionPolicy() *FileDetectionPolicy {
	return &FileDetectionPolicy{Ecosystems: []EcosystemDetectionPolicy{
		{
			Ecosystem:      npmManagerName,
			DefaultManager: npmManagerName,
			Manifests:      []string{"package.json"},
			Lockfiles: []LockfileDetectionPolicy{
				{File: "bun.lock", Manager: "bun"},
				{File: "pnpm-lock.yaml", Manager: "pnpm"},
				{File: "yarn.lock", Manager: "yarn"},
				{File: "package-lock.json", Manager: npmManagerName},
			},
		},
		fileDetectionPolicy("rubygems", "bundler", "Gemfile", "Gemfile.lock"),
		fileDetectionPolicy("cargo", "cargo", "Cargo.toml", "Cargo.lock"),
		fileDetectionPolicy("go", "gomod", "go.mod", ""),
		{
			Ecosystem:      "pypi",
			DefaultManager: "uv",
			Manifests:      []string{"pyproject.toml"},
			Lockfiles: []LockfileDetectionPolicy{
				{File: "uv.lock", Manager: "uv"},
				{File: "poetry.lock", Manager: "poetry"},
			},
		},
		fileDetectionPolicy("packagist", "composer", "composer.json", "composer.lock"),
		fileDetectionPolicy("hex", "mix", "mix.exs", "mix.lock"),
		fileDetectionPolicy("pub", "pub", "pubspec.yaml", "pubspec.lock"),
		fileDetectionPolicy("cocoapods", "cocoapods", "Podfile", "Podfile.lock"),
		{
			Ecosystem:      "maven",
			DefaultManager: "maven",
			Manifests:      []string{"pom.xml"},
			Lockfiles:      []LockfileDetectionPolicy{{File: "gradle.lockfile", Manager: "gradle"}},
		},
		{
			Ecosystem:      nugetManagerName,
			DefaultManager: nugetManagerName,
			Manifests:      []string{"packages.config"},
			Lockfiles: []LockfileDetectionPolicy{
				{File: "packages.lock.json", Manager: nugetManagerName},
				{File: "project.assets.json", Manager: nugetManagerName},
			},
		},
		fileDetectionPolicy("swift", "swift", "Package.swift", "Package.resolved"),
		fileDetectionPolicy("deno", "deno", "deno.json", "deno.lock"),
		{
			Ecosystem:      "hackage",
			DefaultManager: "cabal",
			Manifests:      []string{"*.cabal"},
			Lockfiles: []LockfileDetectionPolicy{
				{File: "cabal.project.freeze", Manager: "cabal"},
				{File: "stack.yaml.lock", Manager: "stack"},
			},
		},
		fileDetectionPolicy("crystal", "shards", "shard.yml", "shard.lock"),
		fileDetectionPolicy("julia", "pkg", "Project.toml", "Manifest.toml"),
		fileDetectionPolicy("conan", "conan", "conanfile.txt", "conan.lock"),
		fileDetectionPolicy("vcpkg", "vcpkg", "vcpkg.json", ""),
		fileDetectionPolicy("carthage", "carthage", "Cartfile", "Cartfile.resolved"),
		fileDetectionPolicy("cpan", "cpanm", "cpanfile", "cpanfile.snapshot"),
		fileDetectionPolicy("cran", "renv", "DESCRIPTION", "renv.lock"),
		fileDetectionPolicy("clojars", "lein", "project.clj", ""),
		fileDetectionPolicy("elm", "elm", "elm.json", ""),
		fileDetectionPolicy("dub", "dub", "dub.json", ""),
		fileDetectionPolicy("haxelib", "haxelib", "haxelib.json", ""),
		fileDetectionPolicy("nix", "nix", "flake.nix", "flake.lock"),
		fileDetectionPolicy("conda", "conda", "environment.yml", ""),
		{
			Ecosystem:      apmManagerName,
			DefaultManager: apmManagerName,
			Manifests:      []string{"apm.yml"},
			Lockfiles: []LockfileDetectionPolicy{
				{File: "apm.lock.yaml", Manager: apmManagerName},
				{File: "apm.lock", Manager: apmManagerName},
			},
		},
	}}
}

func fileDetectionPolicy(
	ecosystem string,
	manager string,
	manifest string,
	lockfile string,
) EcosystemDetectionPolicy {
	policy := EcosystemDetectionPolicy{
		Ecosystem:      ecosystem,
		DefaultManager: manager,
	}
	if manifest != "" {
		policy.Manifests = []string{manifest}
	}
	if lockfile != "" {
		policy.Lockfiles = []LockfileDetectionPolicy{{File: lockfile, Manager: manager}}
	}
	return policy
}
