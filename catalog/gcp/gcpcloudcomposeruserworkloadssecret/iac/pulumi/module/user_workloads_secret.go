package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/composer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// userWorkloadsSecret provisions the Kubernetes Secret Composer manages
// in the environment's GKE cluster. Airflow DAGs (KubernetesPodOperator
// tasks, connections) consume it by name; the material never has to be
// baked into DAG code.
//
// The data updates in place; name, environment, region, and project are
// immutable. The environment must already exist — it is a first-class
// resource this one composes against by reference.
//
// No API enablement here: the Composer API is enabled by the
// environment this Secret is delivered into (a Secret cannot exist
// without one).
func userWorkloadsSecret(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudComposerUserWorkloadsSecret.Spec

	// Values are base64-encoded secret material (the Kubernetes Secret
	// contract). ToSecret marks the whole map secret in Pulumi state; it
	// is never surfaced in outputs.
	args := &composer.UserWorkloadsSecretArgs{
		Name:        pulumi.StringPtr(spec.SecretName),
		Environment: pulumi.String(spec.Environment.GetValue()),
		Region:      pulumi.StringPtr(spec.Region),
		Data:        pulumi.ToSecret(pulumi.ToStringMap(spec.Data)).(pulumi.StringMapOutput),
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Client-side destroy behavior: DELETE (default), PREVENT (destroy
	// fails — protects credentials live pipelines depend on), or ABANDON
	// (drop from state, keep the Secret in the cluster).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdSecret, err := composer.NewUserWorkloadsSecret(ctx, "user-workloads-secret", args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create user workloads secret")
	}

	// The resource ID is the fully qualified resource name — the bridged
	// provider inherits Terraform's ID format, identical to the Terraform
	// module's output. The Secret's data is deliberately never exported.
	ctx.Export(OpName, createdSecret.ID())
	ctx.Export(OpSecretName, createdSecret.Name)

	return nil
}
