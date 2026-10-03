package module

import (
	awsiamaccountsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamaccountsettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// None of IAM's account-settings resources is taggable at AWS, so this
// module carries no tag map - the one deliberate absence against the
// catalog's tag convention (mirrored in the Terraform module).
type Locals struct {
	Target *awsiamaccountsettingsv1alpha1.AwsIamAccountSettings
	Spec   *awsiamaccountsettingsv1alpha1.AwsIamAccountSettingsSpec
}

func initializeLocals(_ *pulumi.Context, in *awsiamaccountsettingsv1alpha1.AwsIamAccountSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	return locals
}
