package module

import (
	"strconv"

	awseksfargateprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksfargateprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEksFargateProfile *awseksfargateprofilev1alpha1.AwsEksFargateProfile

	// FargateProfileName is metadata.name truncated to AWS's 63-character
	// profile limit, deterministically, so the same manifest always yields
	// the same name on both engines.
	FargateProfileName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awseksfargateprofilev1alpha1.AwsEksFargateProfileIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEksFargateProfile = iacInput.Target

	locals.FargateProfileName = iacInput.Target.Metadata.Name
	if len(locals.FargateProfileName) > 63 {
		locals.FargateProfileName = locals.FargateProfileName[:63]
	}

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEksFargateProfile.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
