package version

import (
	"fmt"
	"runtime"
)

var (
	version = "0.0.0" // Filled out during release cutting and provided by ldflags during build
	commit  string    // Provided by ldflags during build
	branch  string    // Provided by ldflags during build
)

// String returns a human-readable version string.
func String() string {
	hasVersion := version != ""
	hasBuildInfo := commit != ""

	switch {
	case hasVersion && hasBuildInfo:
		return fmt.Sprintf("%s (commit %s, branch %s)", version, commit, branch)
	case !hasVersion && hasBuildInfo:
		return fmt.Sprintf("(commit %s, branch %s)", commit, branch)
	case hasVersion && !hasBuildInfo:
		return fmt.Sprintf("%s (no build information)", version)
	default:
		return "(no version or build info)"
	}
}

// Version returns the version string.
func Version() string { return version }

// CommitHash returns the commit hash at which the binary was built.
func CommitHash() string { return commit }

// Branch returns the branch at which the binary was built.
func Branch() string { return branch }

// GoString returns the compiler, compiler version and architecture of the build.
func GoString() string {
	return fmt.Sprintf("%s / %s / %s", runtime.Compiler, runtime.Version(), runtime.GOARCH)
}
