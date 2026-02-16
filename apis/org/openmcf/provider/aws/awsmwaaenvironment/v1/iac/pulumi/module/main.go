package module

import (
	"github.com/pkg/errors"
	awsmwaaenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmwaaenvironment/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MWAA Environment related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmwaaenvironmentv1.AwsMwaaEnvironmentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
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
