package module

import (
	"github.com/pkg/errors"
	gcpbigquerycapacitycommitmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigquerycapacitycommitment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcpbigquerycapacitycommitmentv1alpha1.GcpBigQueryCapacityCommitmentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := capacityCommitment(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create bigquery capacity commitment")
	}

	return nil
}
