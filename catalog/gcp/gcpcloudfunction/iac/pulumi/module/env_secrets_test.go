package module

// A secret an author puts in secret_environment_variables[].value is stored by the module and read
// by the function through a Secret Manager reference; it never becomes a plain value on the
// function, where every viewer reads it. An entry naming a secret the author owns stays exactly
// that reference.

import (
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudfunctionsv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpcloudfunctionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
)

func secretSpec() *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec {
	return &gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec{
		Region: "us-central1",
		ServiceConfig: &gcpcloudfunctionv1alpha1.GcpCloudFunctionServiceConfig{
			SecretEnvironmentVariables: []*gcpcloudfunctionv1alpha1.GcpCloudFunctionSecretEnvVar{
				{Key: "STRIPE_KEY", Value: "sk_live_do_not_leak"},
				{Key: "DB_PASSWORD", Secret: "db-password", Version: "3", ProjectId: "shared-secrets"},
			},
		},
	}
}

func TestAStoredValueBecomesASecretManagerReference_neverAPlainValue(t *testing.T) {
	spec := secretSpec()
	variables := secretVariables(spec)
	if len(variables) != 1 || variables[0].Name != "STRIPE_KEY" {
		t.Fatalf("exactly the value entry is stored; got %+v", variables)
	}
	refs := map[envsecrets.Key]envsecrets.Ref{
		{Name: "STRIPE_KEY"}: {
			Secret:  pulumi.String("function_us-central1_hook_STRIPE_KEY").ToStringOutput(),
			Version: pulumi.String("1").ToStringOutput(),
			Project: pulumi.String("acme-prod").ToStringOutput(),
		},
	}

	envs := serviceConfig(spec, "acme-prod", refs).SecretEnvironmentVariables.(cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArray)
	if len(envs) != 2 {
		t.Fatalf("both entries reach the function; got %d", len(envs))
	}
	stored := envs[0].(*cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArgs)
	if stored.Secret != refs[envsecrets.Key{Name: "STRIPE_KEY"}].Secret || stored.Version != refs[envsecrets.Key{Name: "STRIPE_KEY"}].Version {
		t.Errorf("the stored entry must read the module's own secret and version; got %v %v", stored.Secret, stored.Version)
	}
	owned := envs[1].(*cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArgs)
	if owned.Secret != pulumi.String("db-password") || owned.Version != pulumi.String("3") || owned.ProjectId != pulumi.String("shared-secrets") {
		t.Errorf("an owned secret stays the author's reference; got %v %v %v", owned.Secret, owned.Version, owned.ProjectId)
	}
}

func TestAStoredValueNeedsNoAmbientProjectLookup(t *testing.T) {
	spec := secretSpec()
	spec.ServiceConfig.SecretEnvironmentVariables = spec.ServiceConfig.SecretEnvironmentVariables[:1]
	if needsEffectiveProject(spec) {
		t.Error("a stored value takes its project from the module's own secret; no lookup is needed")
	}
}

func TestThePlacementNamesTheFunctionsSecrets(t *testing.T) {
	spec := secretSpec()
	placement := secretPlacement(&Locals{GcpCloudFunction: &gcpcloudfunctionv1alpha1.GcpCloudFunction{Spec: spec}, FunctionName: "hook"})
	if got := envsecrets.SecretID(placement, secretVariables(spec)[0]); got != "function_us-central1_hook_STRIPE_KEY" {
		t.Errorf("secret id = %q", got)
	}
	if placement.RuntimeServiceAccount != "" || placement.DefaultAccount != "" {
		t.Errorf("a function without an identity falls back to the Compute Engine default; got %+v", placement)
	}
}
