package module

import (
	"strconv"

	awswafregexpatternsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafregexpatternset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set (the Name tag is what the WAF console displays alongside the set).
type Locals struct {
	AwsWafRegexPatternSet *awswafregexpatternsetv1alpha1.AwsWafRegexPatternSet
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awswafregexpatternsetv1alpha1.AwsWafRegexPatternSetIacInput) *Locals {
	locals := &Locals{}
	locals.AwsWafRegexPatternSet = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsWafRegexPatternSet.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsWafRegexPatternSet.Metadata.Org,
		awstagkeys.Environment:  locals.AwsWafRegexPatternSet.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsWafRegexPatternSet.String(),
		awstagkeys.ResourceId:   locals.AwsWafRegexPatternSet.Metadata.Id,
	}

	return locals
}
