package module

import (
	"strconv"

	awsnatgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsnatgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsNatGateway *awsnatgatewayv1alpha1.AwsNatGateway
	AwsTags       map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsnatgatewayv1alpha1.AwsNatGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.AwsNatGateway = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsNatGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
