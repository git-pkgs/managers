package managers

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	ErrNoCommand            = errors.New("no command provided")
	ErrUnsupportedOperation = errors.New("operation not supported by this manager")
	ErrUnsupportedOption    = errors.New("option not supported by this manager")
)

type ErrNoManifest struct {
	Dir string
}

func (e ErrNoManifest) Error() string {
	return fmt.Sprintf("no package manifest found in %s", e.Dir)
}

type ErrCLINotFound struct {
	Manager string
	Binary  string
	Files   []string
}

func (e ErrCLINotFound) Error() string {
	return fmt.Sprintf("%s not found (detected from %s). Install %s or add it to PATH",
		e.Binary, strings.Join(e.Files, ", "), e.Manager)
}

type ErrUnsupportedVersion struct {
	Manager string
	Version string
	Nearest string
}

func (e ErrUnsupportedVersion) Error() string {
	if e.Nearest != "" {
		return fmt.Sprintf("%s %s not supported (nearest: %s)", e.Manager, e.Version, e.Nearest)
	}
	return fmt.Sprintf("%s %s not supported", e.Manager, e.Version)
}

type ErrConflictingLockfiles struct {
	Dir       string
	Lockfiles []string
}

func (e ErrConflictingLockfiles) Error() string {
	return fmt.Sprintf("multiple lockfiles in %s: %s. Remove all but one or specify manager explicitly",
		e.Dir, strings.Join(e.Lockfiles, ", "))
}

type ErrManifestNotInRoot struct {
	Dir      string
	Found    string
	Expected string
}

func (e ErrManifestNotInRoot) Error() string {
	return fmt.Sprintf("found %s but not in project root (found in %s)",
		filepath.Base(e.Found), filepath.Dir(e.Found))
}

type ErrOrphanedWorkspaceMember struct {
	Dir        string
	MemberFile string
}

func (e ErrOrphanedWorkspaceMember) Error() string {
	return fmt.Sprintf("appears to be a workspace member but no workspace root found above %s", e.Dir)
}

type ErrInvalidPackageName struct {
	Name   string
	Reason string
}

func (e ErrInvalidPackageName) Error() string {
	return fmt.Sprintf("invalid package name %q: %s", e.Name, e.Reason)
}

type ErrMissingArgument struct {
	Argument string
}

func (e ErrMissingArgument) Error() string {
	return fmt.Sprintf("missing required argument: %s", e.Argument)
}
