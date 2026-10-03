package module

import (
	"github.com/pkg/errors"
	awsplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the runner appliance. The pieces, in dependency
// order: the token secret, the two IAM roles (setup vs runtime -- kept
// separate so neither accumulates the other's permissions), the log
// group, the outbound-only security group, and finally the compute
// (cluster, task definition, service) that runs the container.
//
// ENROLLMENT IS TOKEN-FIRST: the task ships the runner TOKEN (via Secrets
// Manager), never an identity. The runner joins the control plane on
// first boot, registers itself, and receives its own individually
// revocable identity; task replacement re-joins with the same token (its
// lineage re-admits the runner it originally admitted).
func Resources(ctx *pulumi.Context, iacInput *awsplantonrunnerv1alpha1.AwsPlantonRunnerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsPlantonRunner.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSecret, err := tokenSecret(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create token secret")
	}

	createdExecutionRole, err := executionRole(ctx, locals, provider, createdSecret)
	if err != nil {
		return errors.Wrap(err, "failed to create execution role")
	}

	taskRoleArn, err := runtimeRole(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to resolve runtime role")
	}

	createdSecurityGroup, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create security group")
	}

	if err := runnerCompute(ctx, locals, provider, createdSecret, createdExecutionRole, taskRoleArn, createdSecurityGroup); err != nil {
		return errors.Wrap(err, "failed to create runner compute")
	}

	return nil
}
