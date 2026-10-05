package module

import (
	awsiamgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamgroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// IAM groups (and their membership/policy satellites) are untaggable at
// AWS, so this module carries no tag map - the one deliberate absence
// against the catalog's tag convention (mirrored in the Terraform
// module).
type Locals struct {
	Target *awsiamgroupv1alpha1.AwsIamGroup
	Spec   *awsiamgroupv1alpha1.AwsIamGroupSpec
}

func initializeLocals(_ *pulumi.Context, in *awsiamgroupv1alpha1.AwsIamGroupIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	return locals
}
