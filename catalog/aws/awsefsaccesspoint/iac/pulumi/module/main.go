package module

import (
	"github.com/pkg/errors"
	awsefsaccesspointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsefsaccesspoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsefsaccesspointv1alpha1.AwsEfsAccessPointIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEfsAccessPoint.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := accessPoint(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create access point")
	}

	// --- Exports ---
	ctx.Export(OpAccessPointId, result.AccessPoint.ID())
	ctx.Export(OpAccessPointArn, result.AccessPoint.Arn)
	ctx.Export(OpFileSystemId, result.AccessPoint.FileSystemId)
	ctx.Export(OpFileSystemArn, result.AccessPoint.FileSystemArn)

	return nil
}
