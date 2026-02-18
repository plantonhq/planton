package module

import (
	"github.com/pkg/errors"
	awseksnodegroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awseksnodegroup/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the aws_eks_node_group Pulumi module.
func Resources(ctx *pulumi.Context, stackInput *awseksnodegroupv1.AwsEksNodeGroupStackInput) error {
	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(stackInput.Target.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(stackInput.Target.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	target := stackInput.Target
	spec := target.Spec

	// Build node group arguments using helper function from locals.go
	args := buildNodeGroupArgs(spec)

	created, err := eks.NewNodeGroup(ctx, target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create EKS node group")
	}

	// Exports (align to AwsEksNodeGroupStackOutputs)
	ctx.Export(OpNodeGroupName, created.NodeGroupName)
	ctx.Export(OpAsgName, pulumi.String(""))
	if spec.SshKeyName != "" {
		ctx.Export(OpRemoteAccessSgId, pulumi.String(""))
	}
	ctx.Export(OpInstanceProfileArn, pulumi.String(""))

	return nil
}
