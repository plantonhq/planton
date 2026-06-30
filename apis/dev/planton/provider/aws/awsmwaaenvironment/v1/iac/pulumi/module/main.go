package module

import (
	"github.com/pkg/errors"
	awsmwaaenvironmentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsmwaaenvironment/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MWAA Environment related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmwaaenvironmentv1.AwsMwaaEnvironmentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsMwaaEnvironment.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Managed security group (self-referencing + HTTPS ingress from source SGs/CIDRs)
	createdSg, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "security group")
	}

	// MWAA Environment
	env, err := environment(ctx, locals, provider, createdSg)
	if err != nil {
		return errors.Wrap(err, "mwaa environment")
	}

	// Export outputs
	ctx.Export(OpEnvironmentArn, env.Arn)
	ctx.Export(OpEnvironmentName, env.Name)
	ctx.Export(OpWebserverUrl, env.WebserverUrl)
	ctx.Export(OpAirflowVersion, env.AirflowVersion)
	ctx.Export(OpServiceRoleArn, env.ServiceRoleArn)
	ctx.Export(OpEnvironmentClass, env.EnvironmentClass)
	ctx.Export(OpStatus, env.Status)
	if createdSg != nil {
		ctx.Export(OpSecurityGroupId, createdSg.ID())
	}

	return nil
}
