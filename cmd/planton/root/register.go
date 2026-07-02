package root

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/spf13/cobra"
)

// DefaultPlantonGitRepo is the default local clone path of the Planton
// open-source repository, used by --local-module to run IaC modules from a
// working tree instead of downloaded release artifacts.
const DefaultPlantonGitRepo = "~/scm/github.com/plantonhq/planton"

// Options configures the engine command set for the binary that registers it.
type Options struct {
	// ModulesVersion pins the released IaC module artifacts
	// (downloads.planton.dev/releases/<version>/...) that match the proto
	// schemas compiled into the binary. The standalone planton binary carries
	// its own version stamped via ldflags and leaves this empty; a host binary
	// that embeds the engine as a Go module passes the resolved module version
	// here so module downloads stay schema-consistent with the compiled-in
	// protos. When neither is set, module resolution falls back to the git
	// staging area.
	ModulesVersion string
}

// RegisterCommands wires the Planton open-source engine's user-facing command
// set onto parent, along with the persistent flags those commands require.
//
// This is the embedding contract: the standalone planton binary and any host
// binary that embeds the engine (such as the Planton Platform CLI) register
// commands through the same exported surface, so the two cannot drift.
//
// Binary self-management commands (version, upgrade, downgrade) and developer
// tools (e2e) are deliberately not part of the engine set -- a binary owns its
// own lifecycle. The standalone binary adds them separately.
func RegisterCommands(parent *cobra.Command, opts Options) {
	SetModulesVersion(opts.ModulesVersion)
	RegisterPersistentFlags(parent)

	parent.AddCommand(
		Apply,
		Checkout,
		Destroy,
		Init,
		Kustomize,
		LoadManifest,
		ModulesVersion,
		Plan,
		Pull,
		Pulumi,
		Refresh,
		SecretCoverage,
		Terraform,
		Tofu,
		ValidateManifest,
		ValidateOutputs,
		ValidateRefs,
	)
}

// RegisterPersistentFlags registers the flags every engine command resolves
// through cobra's parent-child flag inheritance. A host that mounts individual
// exported commands (rather than the full set via RegisterCommands) must still
// call this on the mount point, or module resolution inside the commands fails
// at flag lookup.
func RegisterPersistentFlags(parent *cobra.Command) {
	parent.PersistentFlags().Bool(string(flag.LocalModule), false,
		"Use local planton git repository for IaC modules instead of downloading")
	parent.PersistentFlags().String(string(flag.PlantonGitRepo), DefaultPlantonGitRepo,
		"Path to local planton git repository (used with --local-module)")
}

// SetModulesVersion pins the released IaC module artifacts the engine
// downloads (see Options.ModulesVersion). Exported for hosts that mount
// individual commands instead of calling RegisterCommands. Empty input is a
// no-op so a stamped standalone binary is never overwritten with nothing.
func SetModulesVersion(v string) {
	if v != "" {
		version.Version = v
	}
}
