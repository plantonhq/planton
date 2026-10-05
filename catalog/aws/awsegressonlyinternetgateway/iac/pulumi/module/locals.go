package module

import (
	"strconv"

	awsegressonlyinternetgatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsegressonlyinternetgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEgressOnlyInternetGateway *awsegressonlyinternetgatewayv1alpha1.AwsEgressOnlyInternetGateway
	AwsTags                      map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsegressonlyinternetgatewayv1alpha1.AwsEgressOnlyInternetGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEgressOnlyInternetGateway = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEgressOnlyInternetGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
