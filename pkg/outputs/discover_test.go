package outputs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverOverride_Executable(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "with-executable")
	got := discoverOverride(dir)
	if got != OverrideExecutable {
		t.Errorf("expected OverrideExecutable, got %s", got)
	}
}

func TestDiscoverOverride_Mapping(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "with-mapping")
	got := discoverOverride(dir)
	if got != OverrideMapping {
		t.Errorf("expected OverrideMapping, got %s", got)
	}
}

func TestDiscoverOverride_BothPreferExecutable(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "with-both")
	got := discoverOverride(dir)
	if got != OverrideExecutable {
		t.Errorf("expected OverrideExecutable (takes precedence), got %s", got)
	}
}

func TestDiscoverOverride_EmptyDir(t *testing.T) {
	dir := filepath.Join("testdata", "modules", "empty")
	got := discoverOverride(dir)
	if got != OverrideNone {
		t.Errorf("expected OverrideNone, got %s", got)
	}
}

func TestDiscoverOverride_EmptyString(t *testing.T) {
	got := discoverOverride("")
	if got != OverrideNone {
		t.Errorf("expected OverrideNone for empty string, got %s", got)
	}
}

func TestDiscoverOverride_NonexistentDir(t *testing.T) {
	got := discoverOverride("/tmp/does-not-exist-planton-test-dir")
	if got != OverrideNone {
		t.Errorf("expected OverrideNone for nonexistent dir, got %s", got)
	}
}

func TestDiscoverOverride_ExecutableWithoutPermission(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, executableFileName)
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho ok"), 0644); err != nil {
		t.Fatal(err)
	}

	got := discoverOverride(dir)
	if got != OverrideNone {
		t.Errorf("expected OverrideNone for non-executable file, got %s", got)
	}
}

func TestDiscoverOverride_ExecutableIsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, executableFileName), 0755); err != nil {
		t.Fatal(err)
	}

	got := discoverOverride(dir)
	if got != OverrideNone {
		t.Errorf("expected OverrideNone when transform-outputs is a directory, got %s", got)
	}
}

func TestOverrideKind_String(t *testing.T) {
	cases := []struct {
		kind OverrideKind
		want string
	}{
		{OverrideNone, "none"},
		{OverrideExecutable, "executable"},
		{OverrideMapping, "mapping"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("OverrideKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}
