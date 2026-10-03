package module

import (
	"github.com/pkg/errors"
	gcpsecretmanagersecretv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsecretmanagersecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsecretmanagersecretv1alpha1.GcpSecretManagerSecretIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := secret(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	return nil
}
