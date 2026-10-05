package module

import (
	"strconv"

	awsbedrockinferenceprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockinferenceprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbedrockinferenceprofilev1alpha1.AwsBedrockInferenceProfile
	Spec   *awsbedrockinferenceprofilev1alpha1.AwsBedrockInferenceProfileSpec

	// ProfileName is metadata.name -- the naming basis both engines
	// share. AWS allows 1-64 characters.
	ProfileName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbedrockinferenceprofilev1alpha1.AwsBedrockInferenceProfileIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.ProfileName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBedrockInferenceProfile.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
