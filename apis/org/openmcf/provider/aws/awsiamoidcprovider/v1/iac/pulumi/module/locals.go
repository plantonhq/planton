package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	awsiamoidcproviderv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsiamoidcprovider/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamOidcProvider *awsiamoidcproviderv1.AwsIamOidcProvider
	AwsTags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsiamoidcproviderv1.AwsIamOidcProviderStackInput) *Locals {
	locals := &Locals{}
	locals.AwsIamOidcProvider = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamOidcProvider.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamOidcProvider.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsIamOidcProvider.String(),
		awstagkeys.ResourceId:   locals.AwsIamOidcProvider.Metadata.Id,
	}

	return locals
}
