package module

import (
	"strconv"

	awswafipsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafipset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set (the Name tag is what the WAF console displays alongside the set).
type Locals struct {
	AwsWafIpSet *awswafipsetv1alpha1.AwsWafIpSet
	AwsTags     map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awswafipsetv1alpha1.AwsWafIpSetIacInput) *Locals {
	locals := &Locals{}
	locals.AwsWafIpSet = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsWafIpSet.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsWafIpSet.Metadata.Org,
		awstagkeys.Environment:  locals.AwsWafIpSet.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsWafIpSet.String(),
		awstagkeys.ResourceId:   locals.AwsWafIpSet.Metadata.Id,
	}

	return locals
}
