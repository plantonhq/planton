package module

import (
	"github.com/pkg/errors"
	awseksnodegroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awseksnodegroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the aws_eks_node_group Pulumi module.
func Resources(ctx *pulumi.Context, stackInput *awseksnodegroupv1.AwsEksNodeGroupStackInput) error {
	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, stackInput.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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
