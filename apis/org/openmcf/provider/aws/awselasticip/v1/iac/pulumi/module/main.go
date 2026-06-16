package module

import (
	"github.com/pkg/errors"
	awselasticipv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awselasticip/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awselasticipv1.AwsElasticIpStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsElasticIp.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := eip(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create elastic ip")
	}

	ctx.Export(OpAllocationId, result.AllocationId)
	ctx.Export(OpPublicIp, result.PublicIp)
	ctx.Export(OpArn, result.Arn)
	ctx.Export(OpPublicDns, result.PublicDns)

	return nil
}
