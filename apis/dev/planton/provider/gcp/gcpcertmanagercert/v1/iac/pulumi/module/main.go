package module

import (
	"github.com/pkg/errors"
	gcpcertv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcertmanagercert/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the gcp_cert_manager_cert Pulumi module.
// It prepares context, configures the GCP provider, and calls certManagerCert().
func Resources(ctx *pulumi.Context, stackInput *gcpcertv1.GcpCertManagerCertStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Create gcp provider using credentials from the input
	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Call the core logic for certificate creation and DNS validation setup.
	if err := certManagerCert(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create gcp cert manager cert resource")
	}

	return nil
}
