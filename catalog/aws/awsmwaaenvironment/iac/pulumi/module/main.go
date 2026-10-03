package module

import (
	"github.com/pkg/errors"
	awsmwaaenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmwaaenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the AWS MWAA Environment and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsmwaaenvironmentv1alpha1.AwsMwaaEnvironmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsMwaaEnvironment.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	env, err := environment(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "mwaa environment")
	}

	ctx.Export(OpEnvironmentArn, env.Arn)
	ctx.Export(OpEnvironmentName, env.Name)
	ctx.Export(OpWebserverUrl, env.WebserverUrl)
	ctx.Export(OpAirflowVersion, env.AirflowVersion)
	ctx.Export(OpServiceRoleArn, env.ServiceRoleArn)
	ctx.Export(OpEnvironmentClass, env.EnvironmentClass)
	ctx.Export(OpStatus, env.Status)
	ctx.Export(OpCreatedAt, env.CreatedAt)
	ctx.Export(OpDatabaseVpcEndpointService, env.DatabaseVpcEndpointService)
	ctx.Export(OpWebserverVpcEndpointService, env.WebserverVpcEndpointService)

	return nil
}
