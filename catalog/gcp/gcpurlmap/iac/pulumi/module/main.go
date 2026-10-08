package module

import (
	"github.com/pkg/errors"
	gcpurlmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpurlmap/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpurlmapv1alpha1.GcpUrlMapIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, iacInput.Target.GetSpec().GetProjectId().GetValue())
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := urlMap(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create url map")
	}

	return nil
}
