package module

import (
	"strconv"

	awseksaccessentryv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksaccessentry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEksAccessEntry *awseksaccessentryv1alpha1.AwsEksAccessEntry

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awseksaccessentryv1alpha1.AwsEksAccessEntryIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEksAccessEntry = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEksAccessEntry.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
