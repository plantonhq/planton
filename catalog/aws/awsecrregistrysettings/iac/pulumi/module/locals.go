package module

import (
	awsecrregistrysettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecrregistrysettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No AwsTags map here: every registry-level ECR resource this module
// manages is untaggable at AWS (resource_tags on creation templates
// are the STAMPED repositories' tags - user surface, not identity
// tags). The Terraform module carries the same absence.
type Locals struct {
	Target *awsecrregistrysettingsv1alpha1.AwsEcrRegistrySettings
	Spec   *awsecrregistrysettingsv1alpha1.AwsEcrRegistrySettingsSpec
}

func initializeLocals(_ *pulumi.Context, in *awsecrregistrysettingsv1alpha1.AwsEcrRegistrySettingsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	return locals
}
