package runner

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/iacinput/providerenvvars"
)

// IdentityAnnotation makes a scenario deploy as an identity the harness
// creates for it rather than the harness's own administrative credentials.
// The value is provider-interpreted (see provider.IdentityProvisioner); the
// Kubernetes harness accepts "declared" and
// "declared-minus:<apiGroup>/<resource>:<verb>,<verb>". The identity applies
// to the kind under test only: its fixture chain deploys as the
// harness does, because the fixtures are the lane's stage, not its subject.
const IdentityAnnotation = "planton.dev/e2e-identity"

// PhaseIdentity is the phase that creates the lane's identity, after the
// fixtures and SETUP (the identity may need to exist in a fixture-created
// namespace) and before VALIDATE (the binding reads it).
const PhaseIdentity Phase = "IDENTITY"

// provisionIdentity resolves the annotation, requires the harness capability,
// and records the provider configuration on the context so every manifest
// binding uses it. The returned cleanup is a no-op when the scenario declares
// no identity.
func provisionIdentity(ctx context.Context, tc *provider.KindTestContext, harness provider.Harness) (func(), error) {
	spec, err := ManifestAnnotation(tc.ManifestPath, IdentityAnnotation)
	if err != nil || spec == "" {
		return func() {}, nil
	}
	provisioner, ok := harness.(provider.IdentityProvisioner)
	if !ok {
		return func() {}, errors.Errorf("scenario declares %s: %q but the %s harness does not implement provider.IdentityProvisioner",
			IdentityAnnotation, spec, tc.Provider)
	}
	lanef(tc.T, "  [identity] deploying as %q", spec)
	path, cleanup, err := provisioner.ProvisionIdentity(ctx, tc, spec)
	if err != nil {
		return func() {}, errors.Wrapf(err, "provisioning the lane's identity (%s)", spec)
	}
	tc.IdentityProviderConfig = path
	return cleanup, nil
}

// laneProviderConfig is the provider configuration a manifest binding uses:
// the lane's identity when the scenario declared one, else the kind's
// opt-in fixture (or none, the harness's ambient posture).
func laneProviderConfig(tc *provider.KindTestContext, manifestPath string) (*iacinputproviderconfig.ProviderConfig, error) {
	if tc.IdentityProviderConfig == "" {
		return LoadProviderConfigFixture(tc.ModuleDir, manifestPath)
	}
	detected, err := LoadProviderConfigFixture(tc.ModuleDir, manifestPath)
	if err != nil {
		return nil, err
	}
	config := &iacinputproviderconfig.ProviderConfig{Path: tc.IdentityProviderConfig}
	if detected != nil {
		config.Provider = detected.Provider
		return config, nil
	}
	detection, err := detectProvider(manifestPath)
	if err != nil {
		return nil, err
	}
	config.Provider = detection
	return config, nil
}

// laneEngineEnv is the lane identity's credential as environment variables,
// for the Pulumi calls that deploy, re-plan and destroy the kind under test.
// The identity's provider configuration already reaches the program through
// the iac-input, and the modules build their providers from it; this closes
// the one path it does not cover. A Pulumi program inherits the test
// process's environment, so a resource or lookup that rides the default
// provider would otherwise authenticate as the harness's own administrative
// credential and the lane would prove nothing. The variables are the ones the
// Terraform arm already receives (BuildTerraformInput derives them from the
// same iac-input), so both engines see the same single identity. Empty when
// the scenario declares no identity: the lane keeps the harness's posture.
func laneEngineEnv(tc *provider.KindTestContext) ([]string, error) {
	if tc.IdentityProviderConfig == "" {
		return nil, nil
	}
	iacInput, err := os.ReadFile(tc.IacInputFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "reading the lane's iac-input to derive its identity's credential environment")
	}
	vars, err := providerenvvars.GetEnvVarsWithOptions(string(iacInput), providerenvvars.Options{
		FileCacheLoc: filepath.Dir(tc.IacInputFilePath),
		Engine:       providerenvvars.EngineReadsEnvironment,
	})
	if err != nil {
		return nil, errors.Wrap(err, "deriving the lane identity's credential environment")
	}
	env := make([]string, 0, len(vars))
	for name, value := range vars {
		env = append(env, name+"="+value)
	}
	sort.Strings(env)
	return env, nil
}

// Bounds on retrying a deploy refused for a permission the lane's identity
// holds (provider.GrantPropagationJudge): measured on Discovery Engine,
// Dialogflow and Cloud Deploy, whose caches lagged a minute past the
// identity's own acceptance.
var (
	propagationRetryInterval = 60 * time.Second
	propagationRetryBudget   = 6 * time.Minute
)

// deployWhileGrantsPropagate runs deploy and, when the harness judges its
// failure a refusal of a permission the lane's identity holds, waits and runs
// it again until it succeeds, fails another way, or the budget is spent.
// Lanes without a declared identity, and harnesses without the capability,
// deploy once.
func deployWhileGrantsPropagate(tc *provider.KindTestContext, harness provider.Harness, deploy func() error) error {
	judge, ok := harness.(provider.GrantPropagationJudge)
	err := deploy()
	if err == nil || !ok || tc.IdentityProviderConfig == "" {
		return err
	}
	deadline := time.Now().Add(propagationRetryBudget)
	for attempt := 1; err != nil && time.Now().Before(deadline); attempt++ {
		permission, held := judge.HeldPermissionRefused(tc, err)
		if !held {
			return err
		}
		lanef(tc.T, "  [identity] refused %s, which the lane's identity holds: the grant is still propagating; deploying again in %s (retry %d)",
			permission, propagationRetryInterval, attempt)
		time.Sleep(propagationRetryInterval)
		err = deploy()
	}
	return err
}
