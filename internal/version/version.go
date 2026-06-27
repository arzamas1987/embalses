package version

import "runtime/debug"

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "unknown"
)

// Info returns version metadata.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Go      string `json:"go"`
}

// Get returns the current version info.
func Get() Info {
	goVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		goVersion = info.GoVersion
	}
	return Info{
		Version: Version,
		Commit:  Commit,
		Go:      goVersion,
	}
}
