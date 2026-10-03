package module

import (
	"github.com/pkg/errors"
	gcpcertificatemapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertificatemap/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpcertificatemapv1alpha1.GcpCertificateMapIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := certificateMap(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create certificate map")
	}

	return nil
}
