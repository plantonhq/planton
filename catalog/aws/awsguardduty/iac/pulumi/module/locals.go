package module

import (
	"strconv"

	awsguarddutyv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsguardduty/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsguarddutyv1alpha1.AwsGuardDuty
	Spec   *awsguarddutyv1alpha1.AwsGuardDutySpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsguarddutyv1alpha1.AwsGuardDutyIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// The detector, filters, and IP/threat lists are taggable; the
	// feature patches, org surface, and members are not, and the
	// publishing destination is deliberately untagged (ForceNew tags -
	// see detector.go).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsGuardDuty.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
