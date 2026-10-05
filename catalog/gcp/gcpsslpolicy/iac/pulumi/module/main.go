package module

import (
	"github.com/pkg/errors"
	gcpsslpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsslpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpsslpolicyv1alpha1.GcpSslPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := sslPolicy(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create ssl policy")
	}

	return nil
}
