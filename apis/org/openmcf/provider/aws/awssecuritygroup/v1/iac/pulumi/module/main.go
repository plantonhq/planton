package module

import (
	"github.com/pkg/errors"
	awssecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssecuritygroup/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the aws_security_group Pulumi module.
// It reads the AwsSecurityGroupStackInput, sets up AWS credentials if provided,
// and delegates to the securityGroup() function to create the resource.
func Resources(ctx *pulumi.Context, stackInput *awssecuritygroupv1.AwsSecurityGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsSecurityGroup.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsSecurityGroup.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Create the AWS Security Group resource
	if err := securityGroup(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create aws_security_group resource")
	}

	return nil
}
