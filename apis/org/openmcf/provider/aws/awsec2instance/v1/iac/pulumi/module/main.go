package module

import (
	"github.com/pkg/errors"
	awsec2instancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsec2instance/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry‑point invoked by OpenMCF’s CLI.
// It wires provider credentials, initialises locals, and delegates
// to ec2Instance(...) to create the EC2 VM.
func Resources(ctx *pulumi.Context, stackInput *awsec2instancev1.AwsEc2InstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsEc2Instance.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsEc2Instance.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	if err := ec2Instance(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "create aws ec2 instance resource")
	}

	return nil
}
