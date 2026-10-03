package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsiamuserv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamuser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamUser *awsiamuserv1alpha1.AwsIamUser
	AwsTags    map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsiamuserv1alpha1.AwsIamUserIacInput) *Locals {
	locals := &Locals{}
	locals.AwsIamUser = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsIamUser.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamUser.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamUser.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsIamUser.String(),
		awstagkeys.ResourceId:   locals.AwsIamUser.Metadata.Id,
	}

	return locals
}
