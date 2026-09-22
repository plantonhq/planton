package module

import (
	"github.com/pkg/errors"
	gcppscserviceattachmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppscserviceattachment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcppscserviceattachmentv1alpha1.GcpPscServiceAttachmentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := serviceAttachment(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create service attachment")
	}
	return nil
}
