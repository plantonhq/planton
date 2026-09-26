package root

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The engine set is the embedding contract: every user-facing engine command
// must be present, and binary self-management (version/upgrade/downgrade) and
// developer tools (e2e, provider-parity) must not be.
func TestRegisterCommands_EngineSet(t *testing.T) {
	parent := &cobra.Command{Use: "host"}
	RegisterCommands(parent, Options{})

	want := []string{
		"apply", "checkout", "destroy", "init", "kustomize", "load-manifest",
		"module", "modules-version", "plan", "pull", "pulumi", "refresh",
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
	for _, excluded := range []string{"version", "upgrade", "downgrade", "e2e", "provider-parity"} {
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

// Every engine command that RUNS an IaC module reads --module-dir with an
// EMPTY default: empty means "no explicit choice", and the module runtime
// then probes the current directory before falling through to the published
// module for this release. A group that registered a required or defaulted
// --module-dir would break that contract for every command under it, and a
// handler that refused an empty value would contradict the flag's own help.
// (The module-authoring tools, `module verify` and `validate-outputs`,
// inspect a directory the author names and are rightly required.)
func TestEngineCommands_ModuleDirIsOptionalEverywhere(t *testing.T) {
	parent := &cobra.Command{Use: "host"}
	RegisterCommands(parent, Options{})

	runsAModule := map[string]bool{
		"apply": true, "destroy": true, "plan": true, "refresh": true, "init": true,
		"pulumi": true, "tofu": true, "terraform": true,
	}

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if f := c.InheritedFlags().Lookup("module-dir"); f != nil || c.Flags().Lookup("module-dir") != nil {
			if f == nil {
				f = c.Flags().Lookup("module-dir")
			}
			if f.DefValue != "" {
				t.Errorf("%s: --module-dir default = %q, want empty (the resolver owns the default)", c.CommandPath(), f.DefValue)
			}
			if ann := f.Annotations[cobra.BashCompOneRequiredFlag]; len(ann) > 0 {
				t.Errorf("%s: --module-dir is marked required; the module runtime resolves an empty value", c.CommandPath())
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	for _, c := range parent.Commands() {
		if runsAModule[c.Name()] {
			walk(c)
		}
	}
}

// Every command that deploys to Kubernetes reads --kube-context; each engine
// group must therefore register it, or the handler silently sees "" and the
// deploy lands on whatever context the kubeconfig currently selects.
func TestEngineGroups_RegisterKubeContext(t *testing.T) {
	parent := &cobra.Command{Use: "host"}
	RegisterCommands(parent, Options{})

	for _, group := range []string{"pulumi", "tofu", "terraform"} {
		var found *cobra.Command
		for _, c := range parent.Commands() {
			if c.Name() == group {
				found = c
			}
		}
		if found == nil {
			t.Fatalf("engine group %q not registered", group)
		}
		if found.PersistentFlags().Lookup("kube-context") == nil {
			t.Errorf("%s: --kube-context not registered on the group; its handlers read it", group)
		}
	}
	for _, lifecycle := range []string{"apply", "destroy", "plan", "refresh", "init"} {
		for _, c := range parent.Commands() {
			if c.Name() == lifecycle && c.PersistentFlags().Lookup("kube-context") == nil {
				t.Errorf("%s: --kube-context not registered", lifecycle)
			}
		}
	}
}

// A host that embeds the engine owns global flags of its own, declared
// persistently on its root: the Planton Platform CLI's --org, --env/-e,
// --output-format/-o, --instance and --account. An engine command that
// declares a flag with one of those long names hides the host's (the person
// types one, the handler reads the other), and one reusing a shorthand panics
// when cobra merges the flag sets -- `planton kustomize schema -o` did exactly
// that. The engine's own persistent flags are held to the same rule. So no
// engine command, including the chart validate command hosts mount, may
// declare a flag, long name or shorthand, that an ancestor declares
// persistently.
//
// The engine's commands are package-level instances, and cobra merges a
// parent's persistent flags into a command's own set for good; a test that
// ran under another host earlier leaves that host's flags behind. So the walk
// runs in a fresh run of this binary, where this host is the only one.
func TestEngineCommands_ShadowNoPersistentFlag(t *testing.T) {
	if os.Getenv(shadowWalkEnv) == "1" {
		walkEngineUnderHost(t)
		return
	}
	child := exec.Command(os.Args[0], "-test.run=^TestEngineCommands_ShadowNoPersistentFlag$", "-test.count=1", "-test.v")
	child.Env = append(os.Environ(), shadowWalkEnv+"=1")
	out, err := child.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "walked the engine under a host") {
		t.Fatalf("the walk under a fresh host failed (%v):\n%s", err, out)
	}
}

const shadowWalkEnv = "ENGINE_SHADOW_WALK"

func walkEngineUnderHost(t *testing.T) {
	host := &cobra.Command{Use: "host"}
	host.PersistentFlags().String("org", "", "")
	host.PersistentFlags().StringP("env", "e", "", "")
	host.PersistentFlags().StringP("output-format", "o", "table", "")
	host.PersistentFlags().String("instance", "", "")
	host.PersistentFlags().String("account", "", "")
	RegisterCommands(host, Options{})
	chart := &cobra.Command{Use: "chart"}
	chart.AddCommand(NewChartValidateCommand())
	host.AddCommand(chart)

	var walk func(cmd *cobra.Command, inherited []*pflag.Flag)
	walk = func(cmd *cobra.Command, inherited []*pflag.Flag) {
		for _, sub := range cmd.Commands() {
			for _, declared := range []*pflag.FlagSet{sub.Flags(), sub.PersistentFlags()} {
				declared.VisitAll(func(f *pflag.Flag) {
					for _, p := range inherited {
						switch {
						case f == p:
						case f.Name == p.Name:
							t.Errorf("%s declares its own --%s, shadowing the persistent flag", sub.CommandPath(), f.Name)
						case f.Shorthand != "" && f.Shorthand == p.Shorthand:
							t.Errorf("%s declares -%s for --%s, which is the persistent --%s's shorthand", sub.CommandPath(), f.Shorthand, f.Name, p.Name)
						}
					}
				})
			}
			// Cobra merges the inherited flags into the command's own set on every
			// run and every help, and panics on a clash; with nothing reported
			// above, do that merge so any clash the checks miss fails here too.
			if !t.Failed() {
				_ = sub.LocalFlags()
				_ = sub.InheritedFlags()
			}
			var own []*pflag.Flag
			sub.PersistentFlags().VisitAll(func(f *pflag.Flag) { own = append(own, f) })
			walk(sub, append(append([]*pflag.Flag{}, inherited...), own...))
		}
	}
	var hostPersistent []*pflag.Flag
	host.PersistentFlags().VisitAll(func(f *pflag.Flag) { hostPersistent = append(hostPersistent, f) })
	walk(host, hostPersistent)
	t.Log("walked the engine under a host")
}
