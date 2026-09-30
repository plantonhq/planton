package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
)

// A prerequisite deploys on its own kind's engine, whatever the lane runs: an
// undeclared kind keeps Pulumi (every existing chain), a kind that declares
// only HCL engines takes the HCL arm, and the HCL arm is refused for a kind
// that does not declare the binary the lane would run.
func TestDependencyEngine(t *testing.T) {
	cases := []struct {
		slug, binary, want, refusal string
	}{
		{slug: "awsvpc", want: "pulumi"},
		{slug: "stripewebhookendpoint", want: "terraform"},
		{slug: "stripewebhookendpoint", binary: "terraform", refusal: "OpenTofu only"},
		{slug: "openfgastore", binary: "terraform", want: "terraform"},
	}
	for _, c := range cases {
		t.Run(c.slug+"/"+c.binary, func(t *testing.T) {
			t.Setenv("PLANTON_E2E_TF_BINARY", c.binary)
			got, err := dependencyEngine(c.slug)
			if c.refusal != "" {
				if err == nil || !strings.Contains(err.Error(), c.refusal) || !strings.Contains(err.Error(), c.slug) {
					t.Fatalf("dependencyEngine(%q) with binary %q: err %v, want a refusal naming the dependency and %q", c.slug, c.binary, err, c.refusal)
				}
				return
			}
			if err != nil || got != c.want {
				t.Fatalf("dependencyEngine(%q) with binary %q = %q, %v; want %q", c.slug, c.binary, got, err, c.want)
			}
		})
	}
}

// stubTerraformSeams replaces the HCL engine seams for one test and restores
// them afterwards, with teardown retries that never sleep.
func stubTerraformSeams(t *testing.T) {
	t.Helper()
	origDeploy, origOutputs, origDestroy := terraformDeployFn, terraformOutputsFn, terraformDestroyFn
	origPulumiDestroy, origRemove := pulumiDestroyFn, pulumiRemoveStackFn
	origBackoff := dependencyDestroyBackoff
	dependencyDestroyBackoff = 0
	t.Cleanup(func() {
		terraformDeployFn, terraformOutputsFn, terraformDestroyFn = origDeploy, origOutputs, origDestroy
		pulumiDestroyFn, pulumiRemoveStackFn = origPulumiDestroy, origRemove
		dependencyDestroyBackoff = origBackoff
	})
}

// stripeWebhookDependency is a real, loadable OpenTofu-only prerequisite: the
// webhook endpoint's own E2E manifest and module.
func stripeWebhookDependency(t *testing.T) (moduleDir string, dep Dependency) {
	t.Helper()
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locating repo root: %v", err)
	}
	base := filepath.Join(repoRoot, "catalog", "stripe", "stripewebhookendpoint")
	return filepath.Join(base, "iac", "tf"), Dependency{KindSlug: "stripewebhookendpoint", ManifestPath: filepath.Join(base, "e2e", "manifest.yaml")}
}

// A failed apply may still have created resources, so the HCL arm hands back a
// tracked state -- its working copy is what teardown destroys from.
func TestDeployTofuDependency_FailedApplyIsStillTracked(t *testing.T) {
	stubTerraformSeams(t)
	moduleDir, dep := stripeWebhookDependency(t)
	terraformDeployFn = func(testing.TB, *tt.Options) (*TerraformResult, error) {
		return nil, errors.New("Error: creating webhook endpoint: rate limited")
	}

	state, err := deployTofuDependency(t, moduleDir, dep, "dep-stripewebhookendpoint-x", "x")
	if err == nil {
		t.Fatal("expected the failed apply to surface")
	}
	if !state.tracked() || state.Engine != "terraform" || state.terraformOpts == nil {
		t.Fatalf("failed apply returned an untracked state %+v; teardown would leak what the apply created", state)
	}
	if _, statErr := os.Stat(filepath.Join(state.WorkDir, "terraform.tfvars")); statErr != nil {
		t.Fatalf("working copy %s has no generated tfvars: %v", state.WorkDir, statErr)
	}
	state.terraformCleanup()
}

// A successful apply hands the module's outputs back, so the dependent's
// value_from references resolve against them.
func TestDeployTofuDependency_CapturesOutputs(t *testing.T) {
	stubTerraformSeams(t)
	moduleDir, dep := stripeWebhookDependency(t)
	terraformDeployFn = func(testing.TB, *tt.Options) (*TerraformResult, error) { return &TerraformResult{}, nil }
	terraformOutputsFn = func(testing.TB, *tt.Options) (map[string]interface{}, error) {
		return map[string]interface{}{"id": "we_123"}, nil
	}

	state, err := deployTofuDependency(t, moduleDir, dep, "dep-stripewebhookendpoint-x", "x")
	if err != nil {
		t.Fatalf("deployTofuDependency: %v", err)
	}
	t.Cleanup(state.terraformCleanup)
	if state.Outputs["id"] != "we_123" {
		t.Fatalf("outputs = %v, want the module's id", state.Outputs)
	}
}

// The portal's plan-switching scenario names StripePrice; the price's registry
// prerequisite brings the product first. Both deploy on OpenTofu from their own
// install profiles, and the price's product reference resolves against the
// product's output before its variables are written -- an unresolved reference
// would fail the price's tfvars generation outright.
func TestDeployDependencies_StripePriceOnProductChainsOnOpenTofu(t *testing.T) {
	stubTerraformSeams(t)
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locating repo root: %v", err)
	}
	scenario := filepath.Join(repoRoot, "catalog", "stripe", "stripebillingportalconfiguration", "e2e", "scenarios", "plan-switching.yaml")

	deps, err := ResolveDependencies(repoRoot, "stripe", "stripebillingportalconfiguration", scenario)
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 2 || deps[0].KindSlug != "stripeproduct" || deps[1].KindSlug != "stripeprice" {
		t.Fatalf("chain = %+v, want stripeproduct then stripeprice", deps)
	}

	var priceTfvars string
	terraformDeployFn = func(_ testing.TB, opts *tt.Options) (*TerraformResult, error) {
		raw, err := os.ReadFile(opts.VarFiles[0])
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(raw), "currency") {
			priceTfvars = string(raw)
		}
		return &TerraformResult{}, nil
	}
	// The product deploys first, so until the price's variables are seen, outputs are the product's.
	terraformOutputsFn = func(testing.TB, *tt.Options) (map[string]interface{}, error) {
		if priceTfvars == "" {
			return map[string]interface{}{"id": "prod_chain"}, nil
		}
		return map[string]interface{}{"id": "price_chain"}, nil
	}

	states, err := DeployDependencies(context.Background(), t, repoRoot, "stripe", "stripebillingportalconfiguration", scenario, "", "a1b2c3d4", &recordingHarness{})
	t.Cleanup(func() {
		for _, s := range states {
			s.terraformCleanup()
		}
	})
	if err != nil {
		t.Fatalf("DeployDependencies: %v", err)
	}
	if len(states) != 2 || states[0].Engine != "terraform" || states[1].Engine != "terraform" {
		t.Fatalf("states = %+v, want two OpenTofu deploys", states)
	}
	if !strings.Contains(priceTfvars, `"prod_chain"`) {
		t.Fatalf("the price's variables do not carry the product's id:\n%s", priceTfvars)
	}
	if states[1].Outputs["id"] != "price_chain" {
		t.Errorf("price outputs = %v, want its id for the portal to reference", states[1].Outputs)
	}
}

// Teardown runs in reverse across engines: the HCL arm destroys from its
// working copy, never removes a Pulumi stack, and deletes the copy only after
// a successful destroy; a failed destroy keeps the copy and names it.
func TestTeardownDependencies_MixedEngines(t *testing.T) {
	stubTerraformSeams(t)

	var order []string
	pulumiDestroyFn = func(moduleDir, stackName, backendURL, stackInputFilePath string) (*PulumiResult, error) {
		order = append(order, "pulumi:"+stackName)
		return &PulumiResult{}, nil
	}
	var removed []string
	pulumiRemoveStackFn = func(moduleDir, stackName, backendURL string) error {
		removed = append(removed, stackName)
		return nil
	}
	failing := &tt.Options{TerraformDir: "failing"}
	terraformDestroyFn = func(_ testing.TB, opts *tt.Options) (*TerraformResult, error) {
		order = append(order, "tofu:"+opts.TerraformDir)
		if opts == failing {
			return nil, errors.New("Error: archiving product: api error")
		}
		return &TerraformResult{}, nil
	}
	cleaned := map[string]bool{}
	tofuState := func(slug, workDir string, opts *tt.Options) DependencyState {
		return DependencyState{
			Dependency: Dependency{KindSlug: slug}, Engine: "terraform", StackName: "dep-" + slug, WorkDir: workDir,
			terraformOpts: opts, terraformCleanup: func() { cleaned[workDir] = true }, t: t,
		}
	}

	deployed := []DependencyState{
		{Dependency: Dependency{KindSlug: "awsvpc"}, Engine: "pulumi", StackName: "stack-a"},
		tofuState("stripeproduct", "/tmp/wd-product", failing),
		tofuState("stripeprice", "/tmp/wd-price", &tt.Options{TerraformDir: "ok"}),
	}

	err := TeardownDependencies(deployed)
	if err == nil || !strings.Contains(err.Error(), "stripeproduct") || !strings.Contains(err.Error(), "/tmp/wd-product") {
		t.Fatalf("err = %v, want a failure naming the dependency and the kept working copy", err)
	}
	if order[0] != "tofu:ok" || order[len(order)-1] != "pulumi:stack-a" {
		t.Fatalf("teardown order = %v, want the price first and the vpc last", order)
	}
	if got := strings.Count(strings.Join(order, ","), "tofu:failing"); got != dependencyDestroyAttempts {
		t.Errorf("failing HCL destroy ran %d times, want the retry budget %d", got, dependencyDestroyAttempts)
	}
	if !cleaned["/tmp/wd-price"] || cleaned["/tmp/wd-product"] {
		t.Errorf("cleaned = %v, want only the destroyed dependency's copy removed", cleaned)
	}
	if len(removed) != 1 || removed[0] != "stack-a" {
		t.Errorf("removed stacks = %v, want only the Pulumi one", removed)
	}
}
