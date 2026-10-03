package module

import (
	"strconv"

	awsiaminstanceprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiaminstanceprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamInstanceProfile *awsiaminstanceprofilev1alpha1.AwsIamInstanceProfile
	AwsTags               map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsiaminstanceprofilev1alpha1.AwsIamInstanceProfileIacInput) *Locals {
	locals := &Locals{}
	locals.AwsIamInstanceProfile = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsIamInstanceProfile.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
