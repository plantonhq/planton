package module

import (
	"github.com/pkg/errors"
	gcpsccmuteconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccmuteconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsccmuteconfigv1alpha1.GcpSccMuteConfigIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := muteConfig(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create the Security Command Center mute config")
	}

	return nil
}
