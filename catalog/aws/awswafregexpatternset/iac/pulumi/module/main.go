package module

import (
	"github.com/pkg/errors"
	awswafregexpatternsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafregexpatternset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/wafv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one WAFv2 regex pattern set. Name and scope are
// create-time immutable (ForceNew); the expression list itself updates in
// place, which is the point of the resource — rules referencing the set's ARN
// see pattern changes without a web ACL redeploy.
func Resources(ctx *pulumi.Context, iacInput *awswafregexpatternsetv1alpha1.AwsWafRegexPatternSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsWafRegexPatternSet.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain). For CLOUDFRONT scope the spec's CEL pins
	// region to us-east-1 — the WAF global region.
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// One entry per expression. AWS validates the regex dialect server-side
	// (PCRE subset: no backreferences or lookaround), so an unsupported
	// pattern fails the deploy with AWS's own message rather than a module
	// guess.
	regularExpressions := make(wafv2.RegexPatternSetRegularExpressionArray, 0, len(spec.RegularExpressions))
	for _, expression := range spec.RegularExpressions {
		regularExpressions = append(regularExpressions, &wafv2.RegexPatternSetRegularExpressionArgs{
			RegexString: pulumi.String(expression),
		})
	}

	args := &wafv2.RegexPatternSetArgs{
		// The set's AWS name is the Planton resource name — the stable
		// identity web ACL statements and operators see.
		Name:               pulumi.String(locals.AwsWafRegexPatternSet.Metadata.Name),
		Scope:              pulumi.String(spec.Scope),
		RegularExpressions: regularExpressions,
		Tags:               pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	createdRegexPatternSet, err := wafv2.NewRegexPatternSet(ctx,
		locals.AwsWafRegexPatternSet.Metadata.Name,
		args,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create regex pattern set")
	}

	ctx.Export(OpRegexPatternSetArn, createdRegexPatternSet.Arn)
	ctx.Export(OpRegexPatternSetId, createdRegexPatternSet.ID())
	ctx.Export(OpRegexPatternSetName, createdRegexPatternSet.Name)

	return nil
}
