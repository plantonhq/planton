package module

import (
	awssesaccountsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssesaccountsettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
//
// No identity-tag map here: neither upstream account-attribute
// resource carries a tags argument (settings singletons, not taggable
// objects), and metadata.name never reaches AWS.
type Locals struct {
	Target *awssesaccountsettingsv1alpha1.AwsSesAccountSettings
	Spec   *awssesaccountsettingsv1alpha1.AwsSesAccountSettingsSpec
}

func initializeLocals(_ *pulumi.Context, in *awssesaccountsettingsv1alpha1.AwsSesAccountSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	return locals
}
