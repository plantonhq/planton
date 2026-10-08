package module

import (
	"github.com/pkg/errors"
	gcpkmsautokeyconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmsautokeyconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpkmsautokeyconfigv1alpha1.GcpKmsAutokeyConfigIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetScope().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := autokeyConfig(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to apply the KMS Autokey configuration")
	}

	return nil
}
