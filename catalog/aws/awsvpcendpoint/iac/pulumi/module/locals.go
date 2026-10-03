package module

import (
	"strconv"

	awsvpcendpointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpcendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsVpcEndpoint *awsvpcendpointv1alpha1.AwsVpcEndpoint

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsvpcendpointv1alpha1.AwsVpcEndpointIacInput) *Locals {
	locals := &Locals{}
	locals.AwsVpcEndpoint = iacInput.Target

	// VPC endpoints carry no name parameter in AWS -- identity lives
	// entirely in tags, so the Name tag is what the console displays.
	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsVpcEndpoint.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
