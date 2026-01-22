package managers

import (
	"context"
	"time"
)

type Manager interface {
	Name() string
	Ecosystem() string

	Install(ctx context.Context, opts InstallOptions) (*Result, error)
	Add(ctx context.Context, pkg string, opts AddOptions) (*Result, error)
	Remove(ctx context.Context, pkg string) (*Result, error)
	List(ctx context.Context) (*Result, error)
	Outdated(ctx context.Context) (*Result, error)
	Update(ctx context.Context, pkg string) (*Result, error)
	Path(ctx context.Context, pkg string) (*PathResult, error)

	Supports(cap Capability) bool
	Capabilities() []Capability
}

type InstallOptions struct {
	Frozen     bool
	Clean      bool
	Production bool
}

type AddOptions struct {
	Dev       bool
	Optional  bool
	Exact     bool
	Workspace string
}

type Result struct {
	Command  []string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Cwd      string
	Context  ExecContext
}

func (r *Result) Success() bool {
	return r.ExitCode == 0
}

type PathResult struct {
	Path   string // extracted path to the package
	Result *Result // underlying command result
}

type ExecContext int

const (
	ContextProject ExecContext = iota
	ContextGlobal
	ContextWorkspace
)

type Capability int

const (
	CapInstall Capability = iota
	CapInstallFrozen
	CapInstallClean
	CapAdd
	CapAddDev
	CapAddOptional
	CapRemove
	CapUpdate
	CapList
	CapOutdated
	CapAudit
	CapWorkspace
	CapJSONOutput
	CapSBOMCycloneDX
	CapSBOMSPDX
	CapPath
)

var capabilityNames = map[Capability]string{
	CapInstall:       "install",
	CapInstallFrozen: "install_frozen",
	CapInstallClean:  "install_clean",
	CapAdd:           "add",
	CapAddDev:        "add_dev",
	CapAddOptional:   "add_optional",
	CapRemove:        "remove",
	CapUpdate:        "update",
	CapList:          "list",
	CapOutdated:      "outdated",
	CapAudit:         "audit",
	CapWorkspace:     "workspace",
	CapJSONOutput:    "json_output",
	CapSBOMCycloneDX: "sbom_cyclonedx",
	CapSBOMSPDX:      "sbom_spdx",
	CapPath:          "path",
}

func (c Capability) String() string {
	if name, ok := capabilityNames[c]; ok {
		return name
	}
	return "unknown"
}

func CapabilityFromString(s string) (Capability, bool) {
	for cap, name := range capabilityNames {
		if name == s {
			return cap, true
		}
	}
	return 0, false
}
