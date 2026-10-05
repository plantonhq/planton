package module

import (
	"strconv"

	awsec2instancev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsec2instance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEc2Instance *awsec2instancev1alpha1.AwsEc2Instance

	// InstanceName is metadata.name. EC2 instances have no name argument
	// -- the Name tag IS the instance's display identity -- so both
	// engines carry metadata.name in the Name tag and a manifest deploys
	// identically on either.
	InstanceName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsec2instancev1alpha1.AwsEc2InstanceIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEc2Instance = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.InstanceName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEc2Instance.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
