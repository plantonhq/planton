package module

import (
	"github.com/pkg/errors"
	gcpdialogflowcxsecuritysettingsv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxsecuritysettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpdialogflowcxsecuritysettingsv1alpha1.GcpDialogflowCxSecuritySettingsStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := securitySettings(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create dialogflow cx security settings")
	}

	return nil
}
