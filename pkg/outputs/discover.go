package outputs

import (
	"os"
	"path/filepath"
)

// OverrideKind identifies which output transformation override mechanism
// was discovered in a module directory.
type OverrideKind int

const (
	// OverrideNone means no override files were found; the generic
	// reflection-based transformer will be used.
	OverrideNone OverrideKind = iota

	// OverrideExecutable means a transform-outputs executable was found.
	OverrideExecutable

	// OverrideMapping means an output_transform.yaml file was found.
	OverrideMapping
)

const (
	executableFileName = "transform-outputs"
	mappingFileName    = "output_transform.yaml"
)

// String returns a human-readable label for the override kind.
func (k OverrideKind) String() string {
	switch k {
	case OverrideExecutable:
		return "executable"
	case OverrideMapping:
		return "mapping"
	default:
		return "none"
	}
}

// discoverOverride checks a module directory for output transformation
// overrides using the most-specific-wins priority:
//
//  1. transform-outputs executable (+x bit)
//  2. output_transform.yaml
//  3. neither → OverrideNone
//
// If moduleDir is empty or does not exist, returns OverrideNone.
func discoverOverride(moduleDir string) OverrideKind {
	if moduleDir == "" {
		return OverrideNone
	}

	if isExecutableFile(filepath.Join(moduleDir, executableFileName)) {
		return OverrideExecutable
	}

	if isRegularFile(filepath.Join(moduleDir, mappingFileName)) {
		return OverrideMapping
	}

	return OverrideNone
}

// isExecutableFile returns true if path is a regular file with at least
// one execute permission bit set (owner, group, or other).
func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}

// isRegularFile returns true if path exists and is a regular file.
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}
