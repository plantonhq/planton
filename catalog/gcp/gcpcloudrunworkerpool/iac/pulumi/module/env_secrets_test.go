package module

// A secret an author puts in env[].secret_value is stored by the module and read by the worker
// pool's instances through a Secret Manager reference; it never becomes a plain value on the
// worker pool, where every viewer of the worker pool reads it. A literal beside it stays a literal.

import (
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpcloudrunworkerpoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunworkerpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

func TestASecretValueBecomesASecretManagerReference_neverAPlainValue(t *testing.T) {
	spec := &gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolSpec{
		Containers: []*gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolContainer{{
			Name:  "worker",
			Image: "us-docker.pkg.dev/acme/worker:1",
			Env: []*gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolEnvVar{
				{Name: "DB_PASSWORD", SecretValue: "hunter2_do_not_leak"},
				{Name: "MODE", Value: "prod"},
			},
		}},
	}

	variables := secretVariables(spec)
	if len(variables) != 1 || variables[0].Name != "DB_PASSWORD" || variables[0].ContainerIndex != 0 {
		t.Fatalf("exactly the secret_value entry is stored; got %+v", variables)
	}
	refs := map[cloudrunenv.Key]cloudrunenv.Ref{
		{ContainerIndex: 0, Name: "DB_PASSWORD"}: {
			Secret:  pulumi.String("runpool-worker-db-password").ToStringOutput(),
			Version: pulumi.String("1").ToStringOutput(),
		},
	}

	containers := buildContainers(spec, refs)
	envs := containers[0].(*cloudrunv2.WorkerPoolTemplateContainerArgs).Envs.(cloudrunv2.WorkerPoolTemplateContainerEnvArray)

	secret := envs[0].(*cloudrunv2.WorkerPoolTemplateContainerEnvArgs)
	if secret.Value != nil {
		t.Errorf("the secret entry carries a plain value on the worker pool: %v", secret.Value)
	}
	if secret.ValueSource == nil {
		t.Fatal("the secret entry must read its value through a Secret Manager reference")
	}
	literal := envs[1].(*cloudrunv2.WorkerPoolTemplateContainerEnvArgs)
	if literal.Value != pulumi.String("prod") || literal.ValueSource != nil {
		t.Errorf("a literal stays a literal; got value %v, source %v", literal.Value, literal.ValueSource)
	}
}
