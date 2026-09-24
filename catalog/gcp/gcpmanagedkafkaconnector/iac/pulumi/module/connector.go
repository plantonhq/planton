package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/managedkafka"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// connector creates one connector on its Connect cluster.
func connector(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpManagedKafkaConnector.Spec
	resourceName := locals.GcpManagedKafkaConnector.Metadata.Name

	args := &managedkafka.ConnectorArgs{
		Location:       pulumi.String(spec.Location),
		ConnectCluster: pulumi.String(locals.ConnectClusterId),
		ConnectorId:    pulumi.String(locals.ConnectorId),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if len(spec.Configs) > 0 {
		args.Configs = pulumi.ToStringMap(spec.Configs)
	}

	// Omitted, failed tasks stay failed; declared, each backoff is sent
	// only when set so Google's default fills the other -- the Terraform
	// module's rule.
	if policy := spec.TaskRestartPolicy; policy != nil {
		policyArgs := &managedkafka.ConnectorTaskRestartPolicyArgs{}
		if policy.MinimumBackoff != "" {
			policyArgs.MinimumBackoff = pulumi.String(policy.MinimumBackoff)
		}
		if policy.MaximumBackoff != "" {
			policyArgs.MaximumBackoff = pulumi.String(policy.MaximumBackoff)
		}
		args.TaskRestartPolicy = policyArgs
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := managedkafka.NewConnector(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create managed kafka connector")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpConnectorId, created.ConnectorId)
	return nil
}
