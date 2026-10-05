package module

import (
	"strconv"

	awsbedrockcustommodelv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockcustommodel/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockcustommodelv1alpha1.AwsBedrockCustomModel
	Spec   *awsbedrockcustommodelv1alpha1.AwsBedrockCustomModelSpec

	// CustomModelName is metadata.name -- the naming basis both engines
	// share. AWS allows 1-63 characters.
	CustomModelName string

	// JobName defaults to metadata.name. Job names are unique per account
	// FOREVER (AWS never reuses them, even after delete), so re-running a
	// customization needs an explicit spec.job_name.
	JobName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockcustommodelv1alpha1.AwsBedrockCustomModelIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.CustomModelName = metadata.Name

	locals.JobName = locals.Spec.JobName
	if locals.JobName == "" {
		locals.JobName = metadata.Name
	}

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockCustomModel.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
