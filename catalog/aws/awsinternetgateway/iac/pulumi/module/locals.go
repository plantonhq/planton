package module

import (
	"strconv"

	awsinternetgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsinternetgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsInternetGateway *awsinternetgatewayv1alpha1.AwsInternetGateway
	AwsTags            map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsinternetgatewayv1alpha1.AwsInternetGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.AwsInternetGateway = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsInternetGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
