package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsiamoidcproviderv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamoidcprovider/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamOidcProvider *awsiamoidcproviderv1alpha1.AwsIamOidcProvider
	AwsTags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsiamoidcproviderv1alpha1.AwsIamOidcProviderIacInput) *Locals {
	locals := &Locals{}
	locals.AwsIamOidcProvider = iacInput.Target

	// Resource-identity tags match the Terraform module key-for-key
	// (Name plus the planton.ai identity keys).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsIamOidcProvider.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamOidcProvider.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamOidcProvider.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsIamOidcProvider.String(),
		awstagkeys.ResourceId:   locals.AwsIamOidcProvider.Metadata.Id,
	}

	return locals
}
