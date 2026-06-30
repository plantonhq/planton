package module

import (
	"github.com/pkg/errors"
	awsapprunnerservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsapprunnerservice/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that prepares locals, initialises the AWS provider,
// orchestrates VPC Connector, Auto Scaling Configuration, and App Runner Service creation,
// and exports outputs as defined in AwsAppRunnerServiceStackOutputs.
func Resources(ctx *pulumi.Context, stackInput *awsapprunnerservicev1.AwsAppRunnerServiceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Initialise the AWS provider (with or without explicit credentials).
	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsAppRunnerService.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsAppRunnerService.Spec

	// --- VPC Connector (conditional) ---
	// Create an inline VPC Connector when subnet_ids are provided and no existing
	// vpc_connector_arn is referenced. This lets the service reach resources in the VPC.
	var createdVpcConnector *apprunner.VpcConnector
	if len(spec.GetSubnetIds()) > 0 && spec.GetVpcConnectorArn().GetValue() == "" {
		createdVpcConnector, err = vpcConnector(ctx, locals, provider)
		if err != nil {
			return errors.Wrap(err, "failed to create VPC connector")
		}
	}

	// --- Auto Scaling Configuration ---
	// Create an Auto Scaling Configuration Version when the auto_scaling block is provided.
	var createdAutoScaling *apprunner.AutoScalingConfigurationVersion
	if spec.GetAutoScaling() != nil {
		createdAutoScaling, err = autoScalingConfig(ctx, locals, provider)
		if err != nil {
			return errors.Wrap(err, "failed to create auto scaling configuration")
		}
	}

	// --- App Runner Service ---
	if err := service(ctx, locals, provider, createdVpcConnector, createdAutoScaling); err != nil {
		return errors.Wrap(err, "failed to create App Runner service")
	}

	return nil
}
