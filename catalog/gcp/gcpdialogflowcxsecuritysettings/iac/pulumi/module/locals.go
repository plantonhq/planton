package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdialogflowcxsecuritysettingsv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxsecuritysettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig               *gcpprovider.GcpProviderConfig
	GcpDialogflowCxSecuritySettings *gcpdialogflowcxsecuritysettingsv1alpha1.GcpDialogflowCxSecuritySettings

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// Google requires one, and the Terraform module applies the same
	// fallback in locals.tf.
	DisplayName string
}

// initializeLocals derives the defaulted display name. Dialogflow CX
// resources carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpdialogflowcxsecuritysettingsv1alpha1.GcpDialogflowCxSecuritySettingsStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDialogflowCxSecuritySettings = stackInput.Target

	locals.DisplayName = locals.GcpDialogflowCxSecuritySettings.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = locals.GcpDialogflowCxSecuritySettings.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
