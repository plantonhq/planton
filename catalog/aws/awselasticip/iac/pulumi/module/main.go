package module

import (
	"github.com/pkg/errors"
	awselasticipv1alpha1 "github.com/plantonhq/planton/catalog/aws/awselasticip/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awselasticipv1alpha1.AwsElasticIpIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsElasticIp.Spec.Region)
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
	ctx.Export(OpAssociationId, result.AssociationId)
	ctx.Export(OpPtrRecord, result.PtrRecord)

	return nil
}
