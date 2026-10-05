package module

import (
	"strconv"

	awsrdsinstancev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrdsinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRdsInstance *awsrdsinstancev1alpha1.AwsRdsInstance

	// InstanceIdentifier is metadata.name -- the basis both engines share
	// so a manifest deploys identically on either.
	InstanceIdentifier string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsrdsinstancev1alpha1.AwsRdsInstanceIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRdsInstance = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.InstanceIdentifier = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRdsInstance.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
