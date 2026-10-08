package module

import (
	"github.com/pkg/errors"
	gcpsccnotificationconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccnotificationconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsccnotificationconfigv1alpha1.GcpSccNotificationConfigIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetScope().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := notificationConfig(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the Security Command Center notification config")
	}

	return nil
}
