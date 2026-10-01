package module

// A secret an author puts in env[].secret_value is stored by the module and read by the revision
// through a Secret Manager reference; it never becomes a plain value on the service, where every
// viewer of the revision reads it. A literal beside it stays a literal.

import (
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

func TestASecretValueBecomesASecretManagerReference_neverAPlainValue(t *testing.T) {
	spec := &gcpcloudrunv1alpha1.GcpCloudRunSpec{
		Containers: []*gcpcloudrunv1alpha1.GcpCloudRunContainer{{
			Name:  "app",
			Image: "us-docker.pkg.dev/acme/app:1",
			Env: []*gcpcloudrunv1alpha1.GcpCloudRunEnvVar{
				{Name: "STRIPE_KEY", SecretValue: "sk_live_do_not_leak"},
				{Name: "MODE", Value: "prod"},
			},
		}},
	}

	variables := secretVariables(spec)
	if len(variables) != 1 || variables[0].Name != "STRIPE_KEY" || variables[0].ContainerIndex != 0 {
		t.Fatalf("exactly the secret_value entry is stored; got %+v", variables)
	}
	refs := map[cloudrunenv.Key]cloudrunenv.Ref{
		{ContainerIndex: 0, Name: "STRIPE_KEY"}: {
			Secret:  pulumi.String("run-app-stripe-key").ToStringOutput(),
			Version: pulumi.String("1").ToStringOutput(),
		},
	}

	containers := buildContainers(spec, refs)
	envs := containers[0].(*cloudrunv2.ServiceTemplateContainerArgs).Envs.(cloudrunv2.ServiceTemplateContainerEnvArray)

	secret := envs[0].(*cloudrunv2.ServiceTemplateContainerEnvArgs)
	if secret.Value != nil {
		t.Errorf("the secret entry carries a plain value on the service: %v", secret.Value)
	}
	if secret.ValueSource == nil {
		t.Fatal("the secret entry must read its value through a Secret Manager reference")
	}
	literal := envs[1].(*cloudrunv2.ServiceTemplateContainerEnvArgs)
	if literal.Value != pulumi.String("prod") || literal.ValueSource != nil {
		t.Errorf("a literal stays a literal; got value %v, source %v", literal.Value, literal.ValueSource)
	}
}
