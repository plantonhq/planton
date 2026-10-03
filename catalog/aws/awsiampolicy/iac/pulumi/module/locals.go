package module

import (
	"strconv"

	awsiampolicyv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiampolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamPolicy *awsiampolicyv1alpha1.AwsIamPolicy
	AwsTags      map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsiampolicyv1alpha1.AwsIamPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.AwsIamPolicy = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsIamPolicy.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
