package module

import (
	"github.com/pkg/errors"
	gcpmodelarmorfloorsettingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmodelarmorfloorsetting/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpmodelarmorfloorsettingv1alpha1.GcpModelArmorFloorSettingIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetScope().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := floorSetting(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to apply model armor floor setting")
	}

	return nil
}
