package module

import (
	"github.com/pkg/errors"
	awswafipsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafipset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/wafv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one WAFv2 IP set. Name, scope, and address version are
// create-time immutable (ForceNew); the address list itself updates in place,
// which is the point of the resource — rules referencing the set's ARN see
// address changes without a web ACL redeploy.
func Resources(ctx *pulumi.Context, iacInput *awswafipsetv1alpha1.AwsWafIpSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsWafIpSet.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain). For CLOUDFRONT scope the spec's CEL pins
	// region to us-east-1 — the WAF global region.
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	args := &wafv2.IpSetArgs{
		// The set's AWS name is the Planton resource name — the stable
		// identity web ACL statements and operators see.
		Name:             pulumi.String(locals.AwsWafIpSet.Metadata.Name),
		Scope:            pulumi.String(spec.Scope),
		IpAddressVersion: pulumi.String(spec.IpAddressVersion),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}

	// CIDR entries only (the spec's CEL enforces the /nn suffix up front).
	// An empty list is a valid placeholder set that matches nothing until
	// ranges are added.
	if len(spec.Addresses) > 0 {
		args.Addresses = pulumi.ToStringArray(spec.Addresses)
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	createdIpSet, err := wafv2.NewIpSet(ctx,
		locals.AwsWafIpSet.Metadata.Name,
		args,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IP set")
	}

	ctx.Export(OpIpSetArn, createdIpSet.Arn)
	ctx.Export(OpIpSetId, createdIpSet.ID())
	ctx.Export(OpIpSetName, createdIpSet.Name)

	return nil
}
