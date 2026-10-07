package module

import (
	"github.com/pkg/errors"
	gcpkmskeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmskey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpkmskeyv1alpha1.GcpKmsKeyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// No project of its own: the key's project is carried by its key ring's path.
	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig, "")
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := kmsKey(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create kms key")
	}

	return nil
}
