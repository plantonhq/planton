package module

import (
	"strconv"

	awseksnodegroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksnodegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEksNodeGroup *awseksnodegroupv1alpha1.AwsEksNodeGroup

	// NodeGroupName is metadata.name truncated to AWS's 63-character node
	// group limit, deterministically, so the same manifest always yields
	// the same name on both engines.
	NodeGroupName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awseksnodegroupv1alpha1.AwsEksNodeGroupIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEksNodeGroup = iacInput.Target

	locals.NodeGroupName = iacInput.Target.Metadata.Name
	if len(locals.NodeGroupName) > 63 {
		locals.NodeGroupName = locals.NodeGroupName[:63]
	}

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEksNodeGroup.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
