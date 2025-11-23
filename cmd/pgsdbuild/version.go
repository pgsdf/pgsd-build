package main

import "fmt"

var (
	// Version is the current version of pgsdbuild (set via ldflags during build).
	Version = "0.1.0"

	// BuildDate is set during build via ldflags.
	BuildDate = "dev"

	// GitCommit is set during build via ldflags.
	GitCommit = "dev"
)

// VersionInfo returns a formatted version string.
func VersionInfo() string {
	return fmt.Sprintf("pgsdbuild version %s (commit: %s, built: %s)",
		Version, GitCommit, BuildDate)
}
