package runner

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
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
	fmt.Printf("  [identity] deploying as %q\n", spec)
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
