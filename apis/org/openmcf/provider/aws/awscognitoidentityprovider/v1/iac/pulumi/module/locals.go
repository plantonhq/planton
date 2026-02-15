package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	cogidpv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscognitoidentityprovider/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Target  *cogidpv1.AwsCognitoIdentityProvider
	Spec    *cogidpv1.AwsCognitoIdentityProviderSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *cogidpv1.AwsCognitoIdentityProviderStackInput) *Locals {
	locals := &Locals{}
	locals.Target = stackInput.Target
	locals.Spec = stackInput.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsCognitoIdentityProvider.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
