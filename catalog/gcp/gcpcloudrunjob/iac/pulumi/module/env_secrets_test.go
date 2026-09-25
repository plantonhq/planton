package module

// A secret an author puts in env[].secret_value is stored by the module and read by the job's
// task through a Secret Manager reference; it never becomes a plain value on the job, where every
// viewer of the job reads it. A literal beside it stays a literal.

import (
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpcloudrunjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

func TestASecretValueBecomesASecretManagerReference_neverAPlainValue(t *testing.T) {
	tmpl := &gcpcloudrunjobv1alpha1.GcpCloudRunJobTemplate{
		Containers: []*gcpcloudrunjobv1alpha1.GcpCloudRunJobContainer{{
			Name:  "task",
			Image: "us-docker.pkg.dev/acme/task:1",
			Env: []*gcpcloudrunjobv1alpha1.GcpCloudRunJobEnvVar{
				{Name: "DB_PASSWORD", SecretValue: "hunter2_do_not_leak"},
				{Name: "MODE", Value: "prod"},
			},
		}},
	}

	variables := secretVariables(tmpl)
	if len(variables) != 1 || variables[0].Name != "DB_PASSWORD" || variables[0].ContainerIndex != 0 {
		t.Fatalf("exactly the secret_value entry is stored; got %+v", variables)
	}
	refs := map[cloudrunenv.Key]cloudrunenv.Ref{
		{ContainerIndex: 0, Name: "DB_PASSWORD"}: {
			Secret:  pulumi.String("job-task-db-password").ToStringOutput(),
			Version: pulumi.String("1").ToStringOutput(),
		},
	}

	containers := buildContainers(tmpl, refs)
	envs := containers[0].(*cloudrunv2.JobTemplateTemplateContainerArgs).Envs.(cloudrunv2.JobTemplateTemplateContainerEnvArray)

	secret := envs[0].(*cloudrunv2.JobTemplateTemplateContainerEnvArgs)
	if secret.Value != nil {
		t.Errorf("the secret entry carries a plain value on the job: %v", secret.Value)
	}
	if secret.ValueSource == nil {
		t.Fatal("the secret entry must read its value through a Secret Manager reference")
	}
	literal := envs[1].(*cloudrunv2.JobTemplateTemplateContainerEnvArgs)
	if literal.Value != pulumi.String("prod") || literal.ValueSource != nil {
		t.Errorf("a literal stays a literal; got value %v, source %v", literal.Value, literal.ValueSource)
	}
}
