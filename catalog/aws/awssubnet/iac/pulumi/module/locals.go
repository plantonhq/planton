package module

import (
	"strconv"

	awssubnetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssubnet/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsSubnet *awssubnetv1alpha1.AwsSubnet
	AwsTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awssubnetv1alpha1.AwsSubnetIacInput) *Locals {
	locals := &Locals{}
	locals.AwsSubnet = iacInput.Target

	metadata := iacInput.Target.Metadata

	// The settled tag convention, matching the Terraform module key-for-key:
	// user labels merge in FIRST so the Name + planton.ai/* identity keys can
	// never be overridden by a label. Labels reaching AWS as tags is what
	// lets a composition stamp cloud-side discovery tags on a subnet (e.g.
	// Karpenter's karpenter.sh/discovery tag, which its EC2NodeClass subnet
	// selector matches) without a per-kind tags field.
	locals.AwsTags = map[string]string{}
	for key, value := range metadata.Labels {
		locals.AwsTags[key] = value
	}
	locals.AwsTags[awstagkeys.Name] = metadata.Name
	locals.AwsTags[awstagkeys.Resource] = strconv.FormatBool(true)
	locals.AwsTags[awstagkeys.Organization] = metadata.Org
	locals.AwsTags[awstagkeys.Environment] = metadata.Env
	locals.AwsTags[awstagkeys.ResourceKind] = catalogkind.CatalogKind_AwsSubnet.String()
	locals.AwsTags[awstagkeys.ResourceId] = metadata.Id

	return locals
}
