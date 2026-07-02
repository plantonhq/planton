package root

import (
	"testing"

	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/spf13/cobra"
)

// The engine set is the embedding contract: every user-facing engine command
// must be present, and binary self-management (version/upgrade/downgrade) and
// developer tools (e2e) must not be.
func TestRegisterCommands_EngineSet(t *testing.T) {
	parent := &cobra.Command{Use: "host"}
	RegisterCommands(parent, Options{})

	want := []string{
		"apply", "checkout", "destroy", "init", "kustomize", "load-manifest",
		"modules-version", "plan", "pull", "pulumi", "refresh",
		"secret-coverage", "terraform", "tofu", "validate-manifest",
		"validate-outputs", "validate-refs",
	}
	got := map[string]bool{}
	for _, c := range parent.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("engine command %q not registered", name)
		}
	}
	for _, excluded := range []string{"version", "upgrade", "downgrade", "e2e"} {
		if got[excluded] {
			t.Errorf("command %q must not be part of the engine set", excluded)
		}
	}
}

func TestRegisterCommands_PersistentFlags(t *testing.T) {
	parent := &cobra.Command{Use: "host"}
	RegisterCommands(parent, Options{})

	if f := parent.PersistentFlags().Lookup("local-module"); f == nil {
		t.Error("persistent flag --local-module not registered")
	}
	f := parent.PersistentFlags().Lookup("planton-git-repo")
	if f == nil {
		t.Fatal("persistent flag --planton-git-repo not registered")
	}
	if f.DefValue != DefaultPlantonGitRepo {
		t.Errorf("--planton-git-repo default = %q, want %q", f.DefValue, DefaultPlantonGitRepo)
	}
}

func TestSetModulesVersion(t *testing.T) {
	original := version.Version
	defer func() { version.Version = original }()

	version.Version = ""
	SetModulesVersion("v0.3.0")
	if version.Version != "v0.3.0" {
		t.Errorf("version = %q, want v0.3.0", version.Version)
	}

	// Empty input must never erase a stamped version.
	SetModulesVersion("")
	if version.Version != "v0.3.0" {
		t.Errorf("empty SetModulesVersion overwrote stamped version: %q", version.Version)
	}
}
